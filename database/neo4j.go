package database

import "github.com/neo4j/neo4j-go-driver/v4/neo4j"

func HelloWorld(driver neo4j.Driver) (string, error) {

	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

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

func CreateDriver(uri, username, password string) (neo4j.Driver, error) {
	return neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))
}

// CloseDriver call on application exit
func CloseDriver(driver neo4j.Driver) error {
	return driver.Close()
}
