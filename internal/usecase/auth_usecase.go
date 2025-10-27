package usecase

import (
	"be-parkir/internal/domain/entities"
	"be-parkir/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Register(req *entities.CreateUserRequest) (*entities.LoginResponse, error)
	Login(req *entities.LoginRequest) (*entities.LoginResponse, error)
	LoginJukir(username, password string) (*entities.LoginResponse, error)
	RefreshToken(req *entities.RefreshTokenRequest) (*entities.LoginResponse, error)
	Logout(token string) error
	ValidateToken(tokenString string) (*jwt.Token, error)
	GetUserFromToken(token *jwt.Token) (*entities.User, error)
}

type authUsecase struct {
	userRepo  repository.UserRepository
	redis     *redis.Client
	jwtConfig JWTConfig
}

type JWTConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

func NewAuthUsecase(userRepo repository.UserRepository, redis *redis.Client, jwtConfig JWTConfig) AuthUsecase {
	return &authUsecase{
		userRepo:  userRepo,
		redis:     redis,
		jwtConfig: jwtConfig,
	}
}

func (u *authUsecase) Register(req *entities.CreateUserRequest) (*entities.LoginResponse, error) {
	// Check if user already exists
	existingUser, err := u.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &entities.User{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Role:     req.Role,
		Status:   entities.UserStatusActive,
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Generate tokens
	accessToken, refreshToken, err := u.generateTokens(user)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Store refresh token in Redis
	if err := u.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (u *authUsecase) Login(req *entities.LoginRequest) (*entities.LoginResponse, error) {
	// Get user by email
	user, err := u.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if user.Status != entities.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	// Generate tokens
	accessToken, refreshToken, err := u.generateTokens(user)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Store refresh token in Redis
	if err := u.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	// Remove jukir_profile from response to avoid circular reference
	userWithoutProfile := *user
	userWithoutProfile.JukirProfile = nil

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithoutProfile,
	}, nil
}

func (u *authUsecase) LoginJukir(username, password string) (*entities.LoginResponse, error) {
	// Get user by username (username jukir is stored as email in database)
	user, err := u.userRepo.GetByEmail(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is jukir
	if user.Role != entities.RoleJukir {
		return nil, errors.New("user is not a jukir")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check if user is active
	if user.Status != entities.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	// Generate tokens
	accessToken, refreshToken, err := u.generateTokens(user)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Store refresh token in Redis
	if err := u.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	// Remove jukir_profile from response to avoid circular reference
	userWithoutProfile := *user
	userWithoutProfile.JukirProfile = nil

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithoutProfile,
	}, nil
}

func (u *authUsecase) RefreshToken(req *entities.RefreshTokenRequest) (*entities.LoginResponse, error) {
	// Parse and validate refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.jwtConfig.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	// Check if refresh token exists in Redis
	ctx := context.Background()
	storedToken, err := u.redis.Get(ctx, "refresh_token:"+string(rune(userID))).Result()
	if err != nil || storedToken != req.RefreshToken {
		return nil, errors.New("refresh token not found or expired")
	}

	// Get user
	user, err := u.userRepo.GetByID(uint(userID))
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new tokens
	accessToken, refreshToken, err := u.generateTokens(user)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	// Store new refresh token
	if err := u.storeRefreshToken(user.ID, refreshToken); err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	// Remove jukir_profile from response to avoid circular reference
	userWithoutProfile := *user
	userWithoutProfile.JukirProfile = nil

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userWithoutProfile,
	}, nil
}

func (u *authUsecase) Logout(token string) error {
	// Parse token to get user ID
	parsedToken, err := u.ValidateToken(token)
	if err != nil {
		return err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return errors.New("invalid user ID in token")
	}

	// Remove refresh token from Redis
	ctx := context.Background()
	return u.redis.Del(ctx, "refresh_token:"+string(rune(userID))).Err()
}

func (u *authUsecase) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(u.jwtConfig.SecretKey), nil
	})
}

func (u *authUsecase) GetUserFromToken(token *jwt.Token) (*entities.User, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	return u.userRepo.GetByID(uint(userID))
}

func (u *authUsecase) generateTokens(user *entities.User) (string, string, error) {
	// Generate access token
	accessClaims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(u.jwtConfig.AccessExpiry).Unix(),
		"type":    "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(u.jwtConfig.SecretKey))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(u.jwtConfig.RefreshExpiry).Unix(),
		"type":    "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(u.jwtConfig.SecretKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func (u *authUsecase) storeRefreshToken(userID uint, token string) error {
	ctx := context.Background()
	return u.redis.Set(ctx, "refresh_token:"+string(rune(userID)), token, u.jwtConfig.RefreshExpiry).Err()
}
