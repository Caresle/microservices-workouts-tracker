package queries

import (
	"github.com/caresle/microservices-workouts-tracker/user-service/lib"
	"github.com/caresle/microservices-workouts-tracker/user-service/models"
)

type UserQueries struct {
	GetAll                 string
	GetById                string
	GetByEmail             string
	GetByEmailWithPassword string
	Create                 string
	Update                 string
	Delete                 string
}

var queries = UserQueries{
	GetAll:                 "SELECT user_id, name, email FROM tbl_mwt_users",
	GetById:                "SELECT user_id, name, email FROM tbl_mwt_users WHERE user_id = $1",
	GetByEmail:             "SELECT user_id, name, email FROM tbl_mwt_users WHERE email = $1",
	GetByEmailWithPassword: "SELECT user_id, name, email, password FROM tbl_mwt_users WHERE email = $1",
	Create:                 "INSERT INTO tbl_mwt_users (name, email, password) VALUES ($1, $2, $3) RETURNING user_id, name, email",
	Update:                 "UPDATE tbl_mwt_users SET name = $1, email = $2, password = $3 WHERE user_id = $4 RETURNING user_id, name, email",
	Delete:                 "DELETE FROM tbl_mwt_users WHERE user_id = $1",
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

func GetUserByEmail(email string) (*models.User, error) {
	rows, err := lib.Pg(queries.GetByEmail, email)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, rows.Err()
	}

	return models.FromRowToUser(rows)
}

func getUserWithPasswordByEmail(email string) (*models.User, string, error) {
	rows, err := lib.Pg(queries.GetByEmailWithPassword, email)

	if err != nil {
		return nil, "", err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, "", rows.Err()
	}

	return models.FromRowToUserWithPassword(rows)
}

func ValidateUserCredentials(email string, password string) (*models.User, error) {
	user, encryptedPassword, err := getUserWithPasswordByEmail(email)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	err = lib.VerifyPassword(encryptedPassword, password)

	if err != nil {
		return nil, nil
	}

	return user, nil
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

func UpdateUser(user models.User, password string) (*models.User, error) {
	encryptedPassword, err := lib.EncryptPassword(password)

	rows, err := lib.Pg(queries.Update, user.Name, user.Email, encryptedPassword, user.Id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, rows.Err()
	}

	return models.FromRowToUser(rows)
}

func DeleteUser(id int) error {
	_, err := lib.Pg(queries.Delete, id)

	return err
}
