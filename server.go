package main

import (
	"context"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"gql/database"
	"gql/graph"
	"gql/graph/generated"
	"gql/utility"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

/* Runs the server on a thread */
func startHttpServer(wg *sync.WaitGroup, defaultPort string, dbDriver neo4j.Driver) *http.Server {
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{DbDriver: dbDriver}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	serv := &http.Server{Addr: ":" + defaultPort}

	go func() {
		defer wg.Done() // let main know we are done cleaning up
		log.Printf("connect to http://localhost:%s/ for GraphQL playground", defaultPort)
		// always returns error. ErrServerClosed on graceful close
		if err := serv.ListenAndServe(); err != http.ErrServerClosed {
			// unexpected error. port in use?
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return serv
}

func main() {

	config, err := utility.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	// fmt.Println(config)

	// Connect to neo4j
	err = database.CreateDriver(config.Neo4jUri, config.Neo4jUser, config.Neo4jPassword)
	defer func() {
		_ = database.CloseDriver()
	}()

	if err != nil {
		log.Fatal("cannot load database driver ", err)
	}

	// Override default if set in env
	port := os.Getenv("PORT")
	if port == "" {
		port = config.DefaultPort
	}

	log.Printf("main: starting HTTP server")

	// Run server on separate thread
	httpServerExitDone := &sync.WaitGroup{}
	httpServerExitDone.Add(1)
	srv := startHttpServer(httpServerExitDone, config.DefaultPort, database.Driver)

	// Setting up signal capturing then wait for the ctrl+c
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Printf("main: stopping HTTP server")

	// now close the server gracefully ("shutdown")
	// timeout could be given with a proper context
	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}

	// wait for goroutine started in startHttpServer() to stop
	httpServerExitDone.Wait()

	log.Printf("main: done. exiting")
}
