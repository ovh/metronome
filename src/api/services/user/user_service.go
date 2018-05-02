// Package usersrv handle user database operations.
package usersrv

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/ovh/metronome/src/api/models"
	"github.com/ovh/metronome/src/metronome/pg"
)

// Login made a lookup on the database base on username and perform password comparaison.
// It return nil if the username is unknown or the password mismatch.
func Login(username, password string) (*models.User, error) {
	db := pg.DB()

	users := models.Users{}
	err := db.Model(&users).Where("name = ?", username).Select()
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, err
	}

	user := users[0]
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, err
	}

	return &user, nil
}

// Create a new user into the database.
// Return true if the username already exist.
func Create(user *models.User) (bool, error) {
	password, err := genPassword([]byte(user.Password))
	if err != nil {
		return false, err
	}

	user.Password = string(password)

	db := pg.DB()
	res, err := db.Model(&user).OnConflict("DO NOTHING").Insert()
	if err != nil {
		return false, err
	}
	if res.RowsAffected() == 0 {
		return true, nil
	}

	user.Password = "" // remove password hash
	return false, nil
}

// Edit a user in the database.
// Return true if the username already exist.
func Edit(userID string, user *models.User) (bool, error) {
	db := pg.DB()

	var cols []string
	if len(user.Password) > 0 {
		password, err := genPassword([]byte(user.Password))
		if err != nil {
			return false, err
		}
		user.Password = string(password)
		cols = append(cols, "password")
	}

	user.ID = userID
	_, err := db.Model(&user).OnConflict("DO NOTHING").Column(cols...).Update()
	if err != nil {
		return false, err
	}

	user.Password = "" // remove password hash
	return false, nil
}

// Get a user from the database.
// Return nil if the user is not found.
func Get(userID string) (*models.User, error) {
	db := pg.DB()

	var users models.Users
	err := db.Model(&users).Where("user_id = ?", userID).Select()
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	user := users[0]
	user.Password = "" // remove password hash
	return &user, nil
}

// genPassword hash password using bcrypt.
func genPassword(password []byte) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}
