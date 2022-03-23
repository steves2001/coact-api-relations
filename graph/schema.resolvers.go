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

func (r *mutationResolver) UpsertUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	// Data holder for
	var userData model.User

	userData.Name = input.Name
	userData.UserType = input.UserType

	session := r.DbDriver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func(session neo4j.Session) {
		err := session.Close()
		if err != nil {

		}
	}(session)
	// Update or insert?
	if input.ID != nil {
		userData.ID = *input.ID
	} else {
		// Create a Version 4 UUID.
		newUuid, err := uuid.NewV4()
		if err != nil {
			return nil, fmt.Errorf("UUID creation error %v", err)
		}
		userData.ID = newUuid.String()
	}

	result, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		transactionResult, err := transaction.Run(`MERGE (u:User {uuid: $uuid}) 
 ON CREATE SET u.uuid = $uuid, u.name = $name, u.userType = $userType 
 ON MATCH SET  u.uuid = $uuid, u.name = $name, u.userType = $userType  
 RETURN u.uuid, u.name, u.userType`,
			map[string]interface{}{"uuid": userData.ID, "name": userData.Name, "userType": userData.UserType})

		if err != nil {
			return nil, err
		}

		if transactionResult.Next() {

			return &model.User{
				ID:       transactionResult.Record().Values[0].(string),
				Name:     transactionResult.Record().Values[1].(string),
				UserType: model.UserType(transactionResult.Record().Values[2].(string)),
			}, nil

		}

		return nil, transactionResult.Err()
	})

	if err != nil {
		return nil, err
	}
	return result.(*model.User), nil
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
