package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userStore UserStorage
}

func NewUserService(userStore UserStorage) *UserService {
	return &UserService{userStore}
}

func (s *UserService) SignUp(u *CreatUserRequest) (*Token, error) {
	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return nil, fmt.Errorf("password error")
	}
	u.Password = hashedPassword

	userId, err := s.userStore.CreateUser(u)
	if err != nil {
		return nil, fmt.Errorf("user create failed: %d", err)
	}

	token, err := generateJWT(userId, u.Email)
	if err != nil {
		return nil, err
	}

	return &Token{token}, err
}

func (s *UserService) Login(email, password string) (string, error) {
	user, err := s.userStore.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	if match := checkPasswordHash(password, user.HashPassword); !match {
		return "", fmt.Errorf("authentication failure PWD-01")
	}

	return generateJWT(user.ID.Hex(), user.Email)
}

func (s *UserService) GetAllUsers() ([]*User, error) {
	return s.userStore.GetUsers()
}

func (s *UserService) GetUserById(id string) (*User, error) {
	return s.userStore.GetUserById(id)
}

type MyCustomClaims struct {
	UserId           string `json:"userId"`
	UserEmail        string `json:"userEmail"`
	RegisteredClaims jwt.RegisteredClaims
}

func (claims *MyCustomClaims) Valid() error {
	if err := claims.RegisteredClaims.Valid(); err != nil {
		return err
	}
	return nil
}

var mySigningKey = []byte("DFiend")

func generateJWT(userId, userEmail string) (string, error) {
	claims := &MyCustomClaims{
		userId,
		userEmail,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Issuer:    "Bank-API",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(mySigningKey)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
