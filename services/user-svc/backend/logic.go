package backend

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"spendanalyzer.com/user/api"
)

var CreateUserQuery = "INSERT INTO users (username) VALUES (?)"
var GetUserQuery = "SELECT id, username, created_at FROM users WHERE id = (?)"
var GetUserByUsernameQuery = "SELECT id, username, created_at FROM users WHERE username = (?)"

// ErrUserNotFound indicates no user exists with the requested ID.
var ErrUserNotFound = errors.New("user not found")

type UserService struct {
	DB *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{DB: db}
}

// CreateUser creates a new user identity record and returns its generated ID.
func (s *UserService) CreateUser(username string) (api.CreateUserResponse, error) {
	var resp api.CreateUserResponse

	username = strings.TrimSpace(username)
	if username == "" {
		return resp, fmt.Errorf("username must not be empty")
	}

	result, err := s.DB.Exec(CreateUserQuery, username)
	if err != nil {
		return resp, err
	}
	userID, err := result.LastInsertId()
	if err != nil {
		return resp, err
	}

	resp.UserId = int(userID)
	resp.Username = username
	return resp, nil
}

// GetUser fetches a user by ID.
func (s *UserService) GetUser(userID int) (api.GetUserResponse, error) {
	var resp api.GetUserResponse
	var createdAt string

	err := s.DB.QueryRow(GetUserQuery, userID).Scan(&resp.UserId, &resp.Username, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resp, ErrUserNotFound
		}
		return resp, err
	}

	resp.CreatedAt = createdAt
	return resp, nil
}

// GetUserByUsername fetches a user by username - used for "log back in as
// yourself" without a real auth system: re-entering the same name resolves
// to the same existing user_id instead of creating a new one.
func (s *UserService) GetUserByUsername(username string) (api.GetUserResponse, error) {
	var resp api.GetUserResponse
	var createdAt string

	username = strings.TrimSpace(username)
	err := s.DB.QueryRow(GetUserByUsernameQuery, username).Scan(&resp.UserId, &resp.Username, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resp, ErrUserNotFound
		}
		return resp, err
	}

	resp.CreatedAt = createdAt
	return resp, nil
}
