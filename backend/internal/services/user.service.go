package services

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/miloszbo/meals-finder/internal/models"
	repository "github.com/miloszbo/meals-finder/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var key []byte = []byte(os.Getenv("APP_JWT_KEY"))

type UserService interface {
	LoginUser(ctx context.Context, loginData *models.LoginUserRequest) (string, error)
	CreateUser(ctx context.Context, req *models.CreateUserRequest) error
	GetUser(ctx context.Context, username string) (repository.GetUserDataRow, error)
	UpdateUserSettings(ctx context.Context, req *models.UpdateUserSettingsRequest, username string) error
	AddUserTag(ctx context.Context, username string, req *models.UserTag) error
	DisplayUserTag(ctx context.Context, username string) ([]repository.DisplayUserTagRow, error)
	DeleteUserTag(ctx context.Context, username string, tagName string) error
}

type BaseUserService struct {
	DbConn *pgx.Conn
	Repo   *repository.Queries
}

func NewBaseUserService(conn *pgx.Conn) BaseUserService {
	return BaseUserService{
		DbConn: conn,
		Repo:   repository.New(conn),
	}
}

func (s *BaseUserService) LoginUser(ctx context.Context, loginData *models.LoginUserRequest) (string, error) {
	user, err := s.Repo.LoginUserWithUsername(ctx, loginData.Login)
	if err != nil {
		return "", ErrUnauthorizedUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Passwdhash), []byte(loginData.Password)); err != nil {
		return "", ErrUnauthorizedUser
	}

	token, err := s.generateJWT(user.Username)
	if err != nil {
		log.Println(err.Error())
		return "", ErrInternalFailure
	}

	return token, nil
}

func (s *BaseUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) error {
	if err := req.Validate(); err != nil {
		return ErrInternalFailure
	}

	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(req.Passwdhash), bcrypt.DefaultCost)
	if err != nil {
		log.Println("password hashing failed:", err)
		return ErrInternalFailure
	}

	err = s.Repo.CreateUser(ctx, repository.CreateUserParams{
		Username:    req.Username,
		Passwdhash:  string(hashedPasswd),
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Age:         req.Age,
		Sex:         req.Sex,
	})
	if err != nil {
		log.Println("create user failed:", err)
		return ErrInternalFailure
	}

	return nil
}

func (s *BaseUserService) GetUser(ctx context.Context, username string) (repository.GetUserDataRow, error) {
	data, err := s.Repo.GetUserData(ctx, username)

	return data, err
}

func (s *BaseUserService) generateJWT(username string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": username,
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})
	return t.SignedString(key)
}

func (s *BaseUserService) UpdateUserSettings(ctx context.Context, req *models.UpdateUserSettingsRequest, username string) error {
	// Update user settings
	err := s.Repo.UpdateUserSettings(ctx, repository.UpdateUserSettingsParams{
		Username:    username,
		Email:       req.Email,
		Name:        req.Name,
		Surname:     req.Surname,
		PhoneNumber: req.PhoneNumber,
		Age:         req.Age,
		Sex:         req.Sex,
		Weight:      req.Weight,
		Height:      req.Height,
		Bmi:         req.Bmi,
	})
	if err != nil {
		log.Println("update user settings failed:", err)
		return ErrInternalFailure
	}

	return nil
}

func (s *BaseUserService) AddUserTag(ctx context.Context, username string, userTag *models.UserTag) error {
	err := s.Repo.InsertUserTag(ctx, repository.InsertUserTagParams{
		Username:    username,
		TagName:     userTag.Name,
		TagTypeName: userTag.TagType,
	})

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (s *BaseUserService) DeleteUserTag(ctx context.Context, username string, tagName string) error {
	err := s.Repo.DeleteUserTag(ctx, repository.DeleteUserTagParams{
		Username: username,
		TagName:  tagName,
	})

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (s *BaseUserService) DisplayUserTag(ctx context.Context, username string) ([]repository.DisplayUserTagRow, error) {
	data, err := s.Repo.DisplayUserTag(ctx, username)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return data, nil
}

// For testing
type MockUserService struct{}

func (s *MockUserService) LoginUser(ctx context.Context, loginData *models.LoginUserRequest) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": "testUser",
			"exp": time.Now().Add(24 * time.Hour).Unix(),
			"iat": time.Now().Unix(),
		})
	return t.SignedString(key)
}

func (s *MockUserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) error {
	return nil
}
