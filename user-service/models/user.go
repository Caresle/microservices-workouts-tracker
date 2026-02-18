package models

import "github.com/jackc/pgx/v5"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func FromRowToUser(row pgx.Row) (*User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Username, &user.Email)

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
