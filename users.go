package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"crypto/sha256"
	"fmt"
	"time"
)

// User holds the ID, username and projects of a user, but NOT their credentials
type User struct {
	// ID the user unique identifier used to refer to them from other structs
	ID int `db:"id"`

	// Username the user connection string
	Username string `db:"username"`

	// Projects the user’s projects
	Projects []*Project
}

// NewUser registers a new user in the database and returns it, alongide a potential error
func NewUser(username, password string) (*User, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	u := &User{
		Username: username,
	}

	passwordSalt := GenerateSalt(username, password)
	passwordHash := HashPassword(password, passwordSalt)

	row := db.QueryRowx(`insert into autochrone.users
		(username, password_hash, password_salt)
		values ($1, $2, $3) returning id`, u.Username, passwordHash, passwordSalt)
	if err := row.Err(); err != nil {
		return nil, err
	}

	if err := row.Scan(&u.ID); err != nil {
		return nil, err
	}

	return u, nil
}

// GenerateSalt generates a long string used as a salt for a user’s password
func GenerateSalt(username, password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s--autochrone--%s--autochrone--%s", username, time.Now().Format(time.RFC3339Nano), password))))
}

// HashPassword hashes the password with the salt
func HashPassword(password, passwordSalt string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s--autochrone--%s--autochrone--%s", password, passwordSalt, password))))
}

// CheckPassword returns true if given password is correct, false otherwise
func (u *User) CheckPassword(password string) bool {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer db.Close()

	row := db.QueryRowx("select password_hash, password_salt from autochrone.users where username = $1", u.Username)
	if err := row.Err(); err != nil {
		return false
	}

	var passwordHash, passwordSalt string
	if err := row.Scan(&passwordHash, &passwordSalt); err != nil {
		return false
	}

	if passwordHash != HashPassword(password, passwordSalt) {
		return false
	}

	return true
}

// UpdatePassword sets a new password in the database for the given user, returns nil on success, an error otherwise
func (u *User) UpdatePassword(password string) error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	passwordSalt := GenerateSalt(u.Username, password)
	passwordHash := HashPassword(password, passwordSalt)
	_, err = db.Queryx("update autochrone.users set (password_salt, password_hash) = ($1, $2) where id = $3", passwordSalt, passwordHash, u.ID)
	if err != nil {
		return err
	}

	_, err = db.Queryx("delete from autochrone.access_tokens where user_id = $1", u.ID)
	return err
}

// UserExistsWithUsername returns true if a user could be found with such username, otherwise false
func UserExistsWithUsername(username string) bool {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer db.Close()

	row := db.QueryRowx("select true from autochrone.users where username = $1", username)
	if err := row.Err(); err != nil {
		return false
	}

	var found bool
	if err := row.Scan(&found); err != nil {
		return false
	}

	return found
}

// UserExistsWithID returns true if a user could be found with such ID, otherwise false
func UserExistsWithID(id int) bool {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return false
	}
	defer db.Close()

	row := db.QueryRowx("select true from autochrone.users where id = $1", id)
	if err := row.Err(); err != nil {
		return false
	}

	var found bool
	if err := row.Scan(&found); err != nil {
		return false
	}

	return found
}

// GetUserByUsername returns the user with given username and a potential an error
func GetUserByUsername(username string) (*User, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	u := &User{}
	err = db.Get(u, "select id, username from autochrone.users where username = $1", username)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetUserByID returns the user with given ID and a potential error
func GetUserByID(id int) (*User, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	u := &User{}
	err = db.Get(u, "select id, username from autochrone.users where id = $1", id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetUsers returns several users
func GetUsers() ([]*User, error) {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Queryx("select id, username from autochrone.users")
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for rows.Next() {
		u := &User{}
		if err := rows.StructScan(u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

// DeleteUser deletes a user from the database
func DeleteUser(user *User) error {
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Queryx("delete from autochrone.users where id = $1", user.ID); err != nil {
		return err
	}

	return nil
}
