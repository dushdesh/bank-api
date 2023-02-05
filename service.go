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

	return generateJWT(userId, u.Email)
}

func (s *UserService) Login(r *LoginRequest) (*Token, error) {
	user, err := s.userStore.GetUserByEmail(r.Email)
	if err != nil {
		return nil, err
	}
	if match := checkPasswordHash(r.Password, user.HashPassword); !match {
		return nil, fmt.Errorf("authentication failure PWD-01")
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

func generateJWT(userId, userEmail string) (*Token, error) {
	claims := &MyCustomClaims{
		userId,
		userEmail,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			Issuer:    "Bank-API",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwt, err := token.SignedString(mySigningKey)
	return &Token{jwt}, err
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
