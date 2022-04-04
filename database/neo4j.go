package database

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"log"
	"strconv"
	"strings"
)

var Driver neo4j.Driver

type SearchNode struct {
	NodeName    string
	SearchKey   string
	SearchValue string
}

type MultiParamSearchNode struct {
	NodeName     string
	SearchParams map[string]string
	SearchLimit  int64
	Ordering     []string
	Descending   bool
}

// Public functions

// CreateDriver call once at start of application
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

// UpdateInsertQuery Insert or Update a node into the database
func UpdateInsertQuery(node SearchNode, insertionData map[string]string) (map[string]string, error) {

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

	neo4jWriteResult, neo4jWriteErr := writeSingleNodeToDB(query.String(), queryData)

	//  write failed
	if neo4jWriteErr != nil {
		return nil, neo4jWriteErr
	}

	// write success
	if neo4jWriteResult != nil {
		return neo4jWriteResult.(map[string]string), nil
	}

	return nil, fmt.Errorf("single node write operation did not return a result")
}

// SimpleQuery Find a node in the database on a single property
func SimpleQuery(node SearchNode, propertyData []string) (map[string]string, error) {

	queryReturnParameters := ""
	for _, property := range propertyData {
		queryReturnParameters += " n." + property + " AS " + property + ","
	}
	queryReturnParameters = strings.Trim(queryReturnParameters, ",")
	//queryReturnParameters += " n"

	var queryData = make(map[string]interface{})
	queryData[node.SearchKey] = node.SearchValue

	var query strings.Builder
	query.WriteString("MATCH (n:")
	query.WriteString(node.NodeName)
	query.WriteString("{" + node.SearchKey + ": $" + node.SearchKey + "})")
	query.WriteString(" RETURN")
	query.WriteString(queryReturnParameters)

	neo4jReadResult, neo4jReadErr := readSingleNodeFromDB(query.String(), queryData)

	//  read failed
	if neo4jReadErr != nil {
		return nil, fmt.Errorf("single node search did not find node %s with a property %s containing the value %s", node.NodeName, node.SearchKey, node.SearchValue)
		//return nil, neo4jReadErr
	}

	// read found a result
	if neo4jReadResult != nil {
		return neo4jReadResult.(map[string]string), nil
	}

	// Catch all statement shouldn't execute but as a safety net.  Would require nil, nil readSingleNodeFromDB return
	return nil, fmt.Errorf("single node search returned an nil record set with nil error for node %s with a property %s containing the value %s",
		node.NodeName, node.SearchKey, node.SearchValue)
}

func NodeQuery(node MultiParamSearchNode) ([]map[string]string, error) {
	// MATCH (n:User) WHERE n.userType = "STUDENT" RETURN n
	querySearchProperties := ""

	var queryData = make(map[string]interface{})

	for propertyName, propertyVal := range node.SearchParams {
		querySearchProperties += " n." + propertyName + " = $" + propertyName + ","
		queryData[propertyName] = propertyVal
	}

	querySearchProperties = " WHERE" + strings.Trim(querySearchProperties, ",")

	queryReturnParameters := " n"

	queryOrdering := ""

	for _, propertyVal := range node.Ordering {
		queryOrdering += "n." + propertyVal + ", "
	}

	if len(queryOrdering) > 0 {
		queryOrdering = " ORDER BY " + strings.Trim(queryOrdering, ",")
		if node.Descending {
			queryOrdering += queryOrdering + " DESC"
		}
	}

	var query strings.Builder
	query.WriteString("MATCH (n:")
	query.WriteString(node.NodeName)
	query.WriteString(")" + querySearchProperties + "")
	query.WriteString(" RETURN")
	query.WriteString(queryReturnParameters)
	query.WriteString(queryOrdering)
	if node.SearchLimit > 0 {
		query.WriteString(" LIMIT " + strconv.FormatInt(node.SearchLimit, 10))
	}

	neo4jReadResult, neo4jReadErr := readNodesFromDB(query.String(), queryData)

	//  read failed
	if neo4jReadErr != nil {
		return nil, fmt.Errorf("node search did not find any nodes %s", node.NodeName)
		//return nil, neo4jReadErr
	}

	// read found a result
	if neo4jReadResult != nil {

		return neo4jReadResult, nil
	}

	// Catch all statement shouldn't execute but as a safety net.  Would require nil, nil readSingleNodeFromDB return
	return nil, fmt.Errorf("node search returned an nil record set with nil errors")

}

// Private functions
func mapToString(mapData map[string]string, pairStart string, separator string, pairEnd string) string {

	keyValuePairs := make([]string, 0, len(mapData))

	for key := range mapData {
		keyValuePairs = append(keyValuePairs, pairStart+key+separator+mapData[key]+pairEnd)
	}

	return strings.Join(keyValuePairs, ", ")
}
func writeSingleNodeToDB(cypher string, params map[string]interface{}) (interface{}, error) {

	// Open session
	session := Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {
			log.Println(err)
		}
	}(session)

	// Start write data to neo4j
	neo4jWriteResult, neo4jWriteErr := session.WriteTransaction(
		func(transaction neo4j.Transaction) (interface{}, error) {

			transactionResult, driverNativeErr :=
				transaction.Run(cypher, params)

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
			return nodeProperties, transactionResult.Err()
		})

	return neo4jWriteResult, neo4jWriteErr

}

func readSingleNodeFromDB(cypher string, params map[string]interface{}) (interface{}, error) {

	// Open session
	session := Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {
			log.Println(err)
		}
	}(session)

	neo4jReadResult, neo4jReadErr := session.ReadTransaction(
		func(transaction neo4j.Transaction) (interface{}, error) {

			transactionResult, driverNativeErr :=
				transaction.Run(cypher, params)

			// Raw driver error
			if driverNativeErr != nil {
				return nil, driverNativeErr
			}

			nodeProperties := make(map[string]string)

			record, err := transactionResult.Single()

			if err != nil {
				return nil, transactionResult.Err()
			}

			// If result returned
			for index, property := range record.Keys {
				nodeProperties[property] = record.Values[index].(string)
			}
			// Return the created nodes data
			return nodeProperties, nil

		})

	return neo4jReadResult, neo4jReadErr

}
func readNodesFromDB(cypher string, params map[string]interface{}) ([]map[string]string, error) {
	// Open session
	session := Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {
			log.Println(err)
		}
	}(session)

	neo4jReadResult, neo4jReadErr := session.ReadTransaction(
		func(transaction neo4j.Transaction) (interface{}, error) {

			transactionResult, driverNativeErr :=
				transaction.Run(cypher, params)

			// Raw driver error
			if driverNativeErr != nil {
				return nil, driverNativeErr
			}

			// Return the created nodes data
			return transactionResult.Collect()
		})

	var m []map[string]string

	for _, node := range neo4jReadResult.([]*neo4j.Record) {

		nodeProps := make(map[string]string, len(node.Values[0].(neo4j.Node).Props))

		for key, val := range node.Values[0].(neo4j.Node).Props {
			nodeProps[key] = fmt.Sprintf("%v", val)
		}

		m = append(m, nodeProps)

	}

	return m, neo4jReadErr
}
