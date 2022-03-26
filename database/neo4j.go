package database

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"log"
)

var Driver neo4j.Driver

func CreateDriver(uri, username, password string) error {
	var err error
	Driver, err = neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))

	// Local driver error
	if err != nil {
		return err
	}

	// Verify Connectivity
	err = Driver.VerifyConnectivity()
	return err
}

// CloseDriver call on application exit
func CloseDriver() error {
	log.Printf("Closing DB")
	return Driver.Close()
}
