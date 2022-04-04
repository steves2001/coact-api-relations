package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"gql/graph/generated"
	"gql/graph/model"

	"github.com/gofrs/uuid"
)

func (r *mutationResolver) UpsertUser(ctx context.Context, input model.UserInput) (*model.User, error) {
	// Update or insert defined by the presence of an user ID value?
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

	user := model.User{
		ID:       userId,
		Name:     input.Name,
		UserType: input.UserType}

	result, err := r.UpdateInsertUser(user)

	return result, err
}

func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	user := model.User{
		ID:       id,
		Name:     "",
		UserType: ""}

	result, err := r.QueryUser(user)

	if err != nil {
		return nil, err
	}

	return result, err
}

func (r *queryResolver) Users(ctx context.Context, userType model.UserType) ([]*model.User, error) {

	queryUser := model.User{
		ID:       "",
		Name:     "",
		UserType: userType,
	}

	users, err := r.QueryUsers(queryUser)

	if err != nil {
		return nil, err
	}

	return users, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
