package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"gql/graph/generated"
	"gql/graph/model"
	"strconv"
)

// UpsertUser adds or updates a user in the system
func (r *mutationResolver) UpsertUser(ctx context.Context, input model.UserInput) (*model.User, error) {

	// Update or insert?
	var userId string

	if input.ID != nil {
		userId = *input.ID
	} else {
		newUuid, err := uuid.NewV4() // Create a Version 4 UUID.
		if err != nil {
			return nil, fmt.Errorf("UUID creation error %v", err)
		}
		userId = newUuid.String()
	}

	// Open session
	session := r.DbDriver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {

		}
	}(session)

	// Start write data to neo4j
	neo4jWriteResult, neo4jWriteErr := session.WriteTransaction(
		func(transaction neo4j.Transaction) (interface{}, error) {

			transactionResult, driverNativeErr :=
				transaction.Run(
					"MERGE (u:User {uuid: $uuid})  ON CREATE SET u.uuid = $uuid, u.name = $name, u.userType = $userType  ON MATCH SET  u.uuid = $uuid, u.name = $name, u.userType = $userType RETURN u.uuid, u.name, u.userType",
					map[string]interface{}{"uuid": userId, "name": input.Name, "userType": input.UserType})

			// Raw driver error
			if driverNativeErr != nil {
				return nil, driverNativeErr
			}

			// If result returned
			if transactionResult.Next() {

				// Return the created nodes data
				return &model.User{
					ID:       transactionResult.Record().Values[0].(string),
					Name:     transactionResult.Record().Values[1].(string),
					UserType: model.UserType(transactionResult.Record().Values[2].(string)),
				}, nil

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
	return neo4jWriteResult.(*model.User), nil
}

func (r *mutationResolver) UpsertCharacter(ctx context.Context, input model.CharacterInput) (*model.Character, error) {
	id := input.ID
	var character model.Character
	character.Name = input.Name
	character.CliqueType = input.CliqueType

	n := len(r.Resolver.CharacterStore)
	if n == 0 {
		r.Resolver.CharacterStore = make(map[string]model.Character)
	}

	if id != nil {
		cs, ok := r.Resolver.CharacterStore[*id]
		if !ok {
			return nil, fmt.Errorf("not found")
		}
		if input.IsHero != nil {
			character.IsHero = *input.IsHero
		} else {
			character.IsHero = cs.IsHero
		}
		r.Resolver.CharacterStore[*id] = character
	} else {
		// generate unique id
		nid := strconv.Itoa(n + 1)
		character.ID = nid
		if input.IsHero != nil {
			character.IsHero = *input.IsHero
		}
		r.Resolver.CharacterStore[nid] = character
	}

	return &character, nil
}

func (r *queryResolver) Character(ctx context.Context, id string) (*model.Character, error) {
	character, ok := r.Resolver.CharacterStore[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &character, nil
}

func (r *queryResolver) Characters(ctx context.Context, cliqueType model.CliqueType) ([]*model.Character, error) {
	characters := make([]*model.Character, 0)
	for idx := range r.Resolver.CharacterStore {
		character := r.Resolver.CharacterStore[idx]
		if character.CliqueType == cliqueType {

			characters = append(characters, &character)
		}
	}

	return characters, nil
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
