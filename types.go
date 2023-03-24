package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Number int64 `json:"number"`
	Token string `json:"token"`
}

type LoginRequest struct {
	Number int64 `json:"number"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount int `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Password string `json:"password"`
}

type Account struct {
	ID        int `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName,omitempty"`
	Number    int64 `json:"number"`
	HashedPassword string `json:"-"`
	Balance   int64 `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}

func (selfAccount *Account) ValidatePassword(passwordToTest string) bool {
	
	err := bcrypt.CompareHashAndPassword([]byte(selfAccount.HashedPassword), []byte(passwordToTest))
	
	if err != nil {
		return false;
	}

	return true;
}

func NewAccount(firstName string, lastName string, password string) (*Account, error){
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost);

	if err != nil {
		return nil, err;
	};

	return &Account{
		FirstName: firstName,
		LastName: lastName,
		HashedPassword: string(hashedPassword),
		Number: int64(rand.Intn(1000)),
		CreatedAt: time.Now().UTC(),
	}, nil;
}