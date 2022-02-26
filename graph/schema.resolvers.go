package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"gql/graph/generated"
	"gql/graph/model"
	"strconv"
)

func (r *mutationResolver) UpsertUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	var userData model.User

	userData.Name = input.Name
	userData.UserType = input.UserType

	// No users
	if len(r.Resolver.UserStore) == 0 {
		r.Resolver.UserStore = make(map[string]model.User)
	}

	// Update or insert?
	if input.ID != nil {
		_, ok := r.Resolver.UserStore[*input.ID]
		if !ok {
			return nil, fmt.Errorf("not found")
		}
		userData.ID = *input.ID
		r.Resolver.UserStore[*input.ID] = userData
	} else {
		// Create a Version 4 UUID.
		newUuid, err := uuid.NewV4()
		if err != nil {
			return nil, fmt.Errorf("UUID creation error %v", err)
		}
		userData.ID = newUuid.String()
		r.Resolver.UserStore[newUuid.String()] = userData
	}

	return &userData, nil
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
