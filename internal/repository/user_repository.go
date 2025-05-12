package repository

import (
	"Bank/internal/model"
	"database/sql"
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user with given email or username already exists")
)

type UserRepository interface {
	Create(u *model.User) error
	GetByID(id int) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(u *model.User) error {
	// Выполняем INSERT и возвращаем id, created_at
	query := `
        INSERT INTO users(username, email, password_hash)
        VALUES($1, $2, $3)
        RETURNING id, created_at
    `
	return r.db.QueryRow(query, u.Username, u.Email, u.PasswordHash).
		Scan(&u.ID, &u.CreatedAt)
}

func (r *userRepo) GetByID(id int) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (r *userRepo) GetByEmail(email string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (r *userRepo) GetByUsername(username string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).
		Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}
