package graph

//go:generate go run github.com/99designs/gqlgen generate
import "gql/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	CharacterStore map[string]model.Character
	UserStore      map[string]model.User
}
