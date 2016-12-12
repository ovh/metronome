// Package userSrv handle user database operations.
package userSrv

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/runabove/metronome/src/api/models"
	"github.com/runabove/metronome/src/metronome/pg"
)

// Login made a lookup on the database base on username and perform password comparaison.
// It return nil if the username is unknown or the password mismatch.
func Login(username, password string) *models.User {
	db := pg.DB()

	users := models.Users{}
	err := db.Model(&users).Where("name = ?", username).Select()
	if err != nil {
		panic(err)
	}

	if len(users) == 0 {
		return nil
	}

	user := users[0]

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil
	}

	return &user
}

// Create a new user into the database.
// Return true if the username already exist.
func Create(user *models.User) (duplicated bool) {
	user.Password = genPassword(user.Password)

	db := pg.DB()

	res, err := db.Model(&user).OnConflict("DO NOTHING").Insert()
	if err != nil {
		panic(err)
	}
	if res.RowsAffected() == 0 {
		return true
	}

	user.Password = "" // remove password hash
	return false
}

// Edit a user in the database.
// Return true if the username already exist.
func Edit(userID string, user *models.User) (duplicated bool) {
	db := pg.DB()

	var cols []string

	if len(user.Password) > 0 {
		user.Password = genPassword(user.Password)
		cols = append(cols, "password")
	}

	user.ID = userID
	_, err := db.Model(&user).OnConflict("DO NOTHING").Column(cols...).Update()

	if err != nil {
		panic(err)
	}

	user.Password = "" // remove password hash
	return false
}

// Get a user from the database.
// Return nil if the user is not found.
func Get(userID string) *models.User {
	db := pg.DB()

	var users models.Users
	err := db.Model(&users).Where("user_id = ?", userID).Select()
	if err != nil {
		panic(err)
	}

	if len(users) == 0 {
		return nil
	}

	user := users[0]
	user.Password = "" // remove password hash

	return &user
}

// genPassword hash password using bcrypt.
func genPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return string(hash)
}
