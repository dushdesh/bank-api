package main

import (
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Email        string             `json:"email" bson:"email"`
	FirstName    string             `json:"firstName" bson:"first_name"`
	LastName     string             `json:"lastName" bson:"last_name"`
	HashPassword string             `json:"password" bson:"hash_password"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
}

type CreatUserRequest struct {
	Email     string `json:"email,omitempty" bson:"email,omitempty"`
	FirstName string `json:"firstName,omitempty" bson:"first_name,omitempty"`
	LastName  string `json:"lastName,omitempty" bson:"last_name,omitempty"`
	Password  string `json:"password,omitempty" bson:"hash_password,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token,omitempty"`
}

type TransferRequest struct {
	TransferTo int     `json:"transferTo"`
	Amount     float64 `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Account struct {
	ID        int       `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Number    int64     `json:"number"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    int64(rand.Intn(100000)),
		CreatedAt: time.Now().UTC(),
	}
}
