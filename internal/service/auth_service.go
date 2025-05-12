package service

import (
	"Bank/internal/model"
	"Bank/internal/repository"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(u repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  u,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(reg *model.UserRegistration) (*model.User, error) {
	if _, err := s.userRepo.GetByEmail(reg.Email); err == nil {
		return nil, repository.ErrUserExists
	}
	if _, err := s.userRepo.GetByUsername(reg.Username); err == nil {
		return nil, repository.ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     reg.Username,
		Email:        reg.Email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(login *model.UserLogin) (string, error) {
	u, err := s.userRepo.GetByEmail(login.Email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(login.Password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return s.generateToken(u.ID)
}

func (s *AuthService) generateToken(userID int) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
