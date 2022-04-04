package graph

//go:generate go run github.com/99designs/gqlgen generate
import (
	"gql/database"
	"gql/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserStore map[string]model.User
}

// UpdateInsertUser Convert model a map then call the db method to update or insert a user
func (r Resolver) UpdateInsertUser(insertionData model.User) (*model.User, error) {

	// Unpack data for the database model to map
	userData := map[string]string{"uuid": insertionData.ID, "name": insertionData.Name, "userType": insertionData.UserType.String()}

	result, databaseErr := database.UpdateInsertQuery(database.SearchNode{NodeName: "User", SearchKey: "uuid", SearchValue: insertionData.ID},
		userData)

	// Database error returned
	if databaseErr != nil {
		return nil, databaseErr
	}

	// Return the node created/updated data
	return &model.User{
		ID:       result["uuid"],
		Name:     result["name"],
		UserType: model.UserType(result["userType"]),
	}, nil

}

func (r Resolver) QueryUser(userData model.User) (*model.User, error) {

	returnParams := []string{"uuid", "name", "userType"}

	result, databaseErr := database.SimpleQuery(database.SearchNode{
		NodeName:    "User",
		SearchKey:   "uuid",
		SearchValue: userData.ID,
	}, returnParams)

	// Database error returned
	if databaseErr != nil {
		return nil, databaseErr
	}

	// Return the node created/updated data
	return &model.User{
		ID:       result["uuid"],
		Name:     result["name"],
		UserType: model.UserType(result["userType"]),
	}, nil

}

func (r Resolver) QueryUsers(userData model.User) ([]*model.User, error) {

	searchParameters := map[string]string{"userType": userData.UserType.String()}
	resultPtr, databaseErr := database.NodeQuery(database.MultiParamSearchNode{
		NodeName:     "User",
		SearchParams: searchParameters,
		SearchLimit:  0,
		Ordering:     nil,
		Descending:   false,
	})
	// Database error returned
	if databaseErr != nil {
		return nil, databaseErr
	}

	var u []*model.User

	//var u []model.User

	for _, currentData := range *resultPtr {
		// change map to users
		u = append(u, &model.User{
			ID:       currentData["uuid"],
			Name:     currentData["name"],
			UserType: model.UserType(currentData["userType"]),
		})
	}

	return u, nil

}
