package database

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"log"
	"strings"
)

var Driver neo4j.Driver

type SearchNode struct {
	NodeName    string
	SearchKey   string
	SearchValue string
}

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

// UpdateInsertQuery Insert or Update a user into the database
func UpdateInsertQuery(node SearchNode, insertionData map[string]string) (map[string]string, error) {

	// Open session
	session := Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {
			log.Println(err)
		}
	}(session)

	var queryParameters = ""
	var queryReturnParameters = ""
	var queryData = make(map[string]interface{})
	for property, value := range insertionData {
		queryParameters += " n." + property + " = $" + property + ","
		queryReturnParameters += " n." + property + " AS " + property + ","
		queryData[property] = value
	}

	queryParameters = strings.Trim(queryParameters, ",")
	queryReturnParameters = strings.Trim(queryReturnParameters, ",")

	var query strings.Builder
	query.WriteString("MERGE (n:")
	query.WriteString(node.NodeName)
	query.WriteString("{" + node.SearchKey + ": $" + node.SearchKey + "})")
	query.WriteString(" ON CREATE SET")
	query.WriteString(queryParameters)
	query.WriteString(" ON MATCH SET")
	query.WriteString(queryParameters)
	query.WriteString(" RETURN")
	query.WriteString(queryReturnParameters)

	// Start write data to neo4j
	neo4jWriteResult, neo4jWriteErr := session.WriteTransaction(
		func(transaction neo4j.Transaction) (interface{}, error) {

			transactionResult, driverNativeErr :=
				transaction.Run(query.String(), queryData)

			// Raw driver error
			if driverNativeErr != nil {
				return nil, driverNativeErr
			}
			nodeProperties := make(map[string]string)
			// If result returned
			if transactionResult.Next() {

				for index, property := range transactionResult.Record().Keys {
					nodeProperties[property] = transactionResult.Record().Values[index].(string)
				}
				// Return the created nodes data
				return nodeProperties, nil

			}

			// Node wasn't created there was an error return this
			return nil, transactionResult.Err()
		})
	// End write data to neo4j

	//  write failed
	if neo4jWriteErr != nil {
		return nil, neo4jWriteErr
	}

	// write success
	return neo4jWriteResult.(map[string]string), nil
}
