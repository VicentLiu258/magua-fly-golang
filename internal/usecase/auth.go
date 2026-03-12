package usecase

import (
	"context"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/pkg/jwt"
	"magua-fly-golang/internal/pkg/password"
	"magua-fly-golang/internal/repository"
)

// AuthUseCase handles login and token issuance.
type AuthUseCase struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
}

// NewAuthUseCase creates an AuthUseCase.
func NewAuthUseCase(userRepo repository.UserRepository, jwtManager *jwt.Manager) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo, jwtManager: jwtManager}
}

// LoginRequest is the login input.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse returns the token and user info.
type LoginResponse struct {
	Token string      `json:"token"`
	User  *domain.User `json:"user"`
}

// Login validates credentials and returns a JWT.
func (u *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if !password.Check(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}
	user.PasswordHash = ""
	token, err := u.jwtManager.Issue(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}
	return &LoginResponse{Token: token, User: user}, nil
}

// RegisterRequest is the registration input.
type RegisterRequest struct {
	Username string          `json:"username" binding:"required"`
	Password string          `json:"password" binding:"required"`
	Phone    string          `json:"phone"`
	Email    string          `json:"email"`
	Role     domain.UserRole `json:"role" binding:"required"`
}

// Register creates a new user.
func (u *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*domain.User, error) {
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Username:     req.Username,
		PasswordHash: hash,
		Phone:        req.Phone,
		Email:        req.Email,
		Role:         req.Role,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}
