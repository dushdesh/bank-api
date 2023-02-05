package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	jwt "github.com/golang-jwt/jwt/v4"
)

type ApiServer struct {
	listenAddr   string
	accountStore AccountStorage
	userService  *UserService
}

func NewApiServer(listenAddr string, accountStore AccountStorage, userService *UserService) *ApiServer {
	return &ApiServer{
		listenAddr:   listenAddr,
		accountStore: accountStore,
		userService:  userService,
	}
}

func (s *ApiServer) Run() {
	router := chi.NewRouter()
	router.HandleFunc("/signup", makeHttpHandleFunc(s.handleUser))
	router.HandleFunc("/login", makeHttpHandleFunc(s.handleLogin))
	router.HandleFunc("/user", makeHttpHandleFunc(s.handleUser))
	router.HandleFunc("/user/{id}", makeHttpHandleFunc(s.handleGetUserById))

	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withAuth(makeHttpHandleFunc(s.handleAccountByID)))

	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))
	log.Println("Bank API running on server: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *ApiServer) handleUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		return s.handleSignup(w, r)
	}
	if r.Method == "GET" {
		return s.handleGetAllUsers(w, r)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}


func (s *ApiServer) handleSignup(w http.ResponseWriter, r *http.Request) error {
	createUserReq := new(CreatUserRequest)
	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
		return err
	}
	defer r.Body.Close()

	token, err := s.userService.SignUp(createUserReq)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusCreated, token)
}

func (s *ApiServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}

	loginRequest := new(LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		return err
	}
	defer r.Body.Close()


	token, err := s.userService.Login(loginRequest)
	if err != nil {
		return fmt.Errorf("authentication failure")
	}
	
	return WriteJSON(w, http.StatusOK, token)
}

func (s *ApiServer) handleGetAllUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, users)
}

func (s *ApiServer) handleGetUserById(w http.ResponseWriter, r *http.Request) error {
	id, err := extractId(r)
	if err != nil {
		return err
	}
	user, err := s.userService.GetUserById(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, user)
}

func (s *ApiServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}

func (s *ApiServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := extractId(r)
	if err != nil {
		return err
	}

	if r.Method == "GET" {
		return s.handleGetAccountByID(w, r, id)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccountByID(w, r, id)
	}

	return fmt.Errorf("method not allowed: %s", r.Method)
}

func (s *ApiServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.accountStore.GetAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (s *ApiServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}
	defer r.Body.Close()

	acc := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)
	if err := s.accountStore.CreateAccount(acc); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusCreated, acc)
}

func (s *ApiServer) handleDeleteAccountByID(w http.ResponseWriter, r *http.Request, id string) error {
	if err := s.accountStore.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, nil)
}

func (s *ApiServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request, id string) error {
	acc, err := s.accountStore.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, acc)
}

func (s *ApiServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		transferReq := new(TransferRequest)
		if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
			return err
		}
		defer r.Body.Close()

		return WriteJSON(w, http.StatusOK, transferReq)
	}
	return fmt.Errorf("method not allowed: %s", r.Method)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func extractId(r *http.Request) (string, error) {
	idStr := chi.URLParam(r, "id")
	return idStr, nil
}

func withAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("In the Auth middleware")
		tokenString := r.Header.Get("Authorization")
		_, err := validateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(token string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}
