package graph

//go:generate go run github.com/99designs/gqlgen generate
import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"gql/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DbDriver       neo4j.Driver
	CharacterStore map[string]model.Character
	UserStore      map[string]model.User
}

// HelloWorld This should not be here, but it works for testing
func (r Resolver) HelloWorld() (string, error) {

	session := r.DbDriver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {

		}
	}(session)

	greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			"CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)",
			map[string]interface{}{"message": "hello, world"})
		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}

	return greeting.(string), nil
}
