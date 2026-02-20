package models

import (
	user_request "github.com/caresle/microservices-workouts-tracker/user-service/request"
	"github.com/jackc/pgx/v5"
)

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func FromRowToUser(row pgx.Row) (*User, error) {
	var user User
	err := row.Scan(&user.Id, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func FromRowsToUsers(rows pgx.Rows) ([]*User, error) {
	var users []*User

	for rows.Next() {
		user, err := FromRowToUser(rows)

		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func FromCreateRequestToUser(request user_request.CreateUserRequest) (*User, string) {
	return &User{
		Name:  request.Name,
		Email: request.Email,
	}, request.Password
}

func FromUpdateRequestToUser(request user_request.UpdateUserRequest) *User {
	return &User{
		Id:    request.Id,
		Name:  request.Name,
		Email: request.Email,
	}
}
