package queries

import (
	"github.com/caresle/microservices-workouts-tracker/user-service/lib"
	"github.com/caresle/microservices-workouts-tracker/user-service/models"
)

type UserQueries struct {
	GetAll  string
	GetById string
	Create  string
}

var queries = UserQueries{
	GetAll:  "SELECT user_id, name, email FROM tbl_mwt_users",
	GetById: "SELECT user_id, name, email FROM tbl_mwt_users WHERE user_id = $1",
	Create:  "INSERT INTO tbl_mwt_users (name, email, password) VALUES ($1, $2, $3) RETURNING user_id, name, email",
}

func GetAllUsers() ([]*models.User, error) {
	rows, err := lib.Pg(queries.GetAll)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return models.FromRowsToUsers(rows)
}

func GetUserById(id int) (*models.User, error) {
	rows, err := lib.Pg(queries.GetById, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, rows.Err()
	}

	return models.FromRowToUser(rows)
}

func CreateUser(user models.User, password string) (*models.User, error) {
	encryptedPassword, err := lib.EncryptPassword(password)

	rows, err := lib.Pg(queries.Create, user.Name, user.Email, encryptedPassword)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, rows.Err()
	}

	return models.FromRowToUser(rows)
}
