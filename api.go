package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, struct{ status string }{status: "Up"})
	})

	r.Post("/signup", makeHttpHandleFunc(s.handleSignup))
	r.Post("/login", makeHttpHandleFunc(s.handleLogin))
	r.Route("/user",
		func(r chi.Router) {
			r.Get("/{id}", makeHttpHandleFunc(s.handleGetUserById))
			r.Delete("/{id}", makeHttpHandleFunc(s.handleDeleteUserById))
			// router.Get("{id}/account/{accountId}", withAuth(makeHttpHandleFunc(s.handleAccountByID)))
		})

	r.Mount("/admin", adminRouter())

	r.Route("/admin", func(r chi.Router) {
		r.Get("/user", makeHttpHandleFunc(s.handleGetAllUsers))
		r.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	})

	r.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))
	log.Println("Bank API running on server: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, r)
}

type ApiError struct {
	Err    string `json:"error"`
	Status int    `json:"status"`
}

func (e ApiError) Error() string {
	return e.Err
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if e, ok := err.(ApiError); ok {
				WriteJSON(w, e.Status, e)
			}
			WriteJSON(w, http.StatusBadRequest, ApiError{Err: err.Error(), Status: http.StatusBadRequest})
		}
	}
}

func (s *ApiServer) handleSignup(w http.ResponseWriter, r *http.Request) error {
	createUserReq := new(CreatUserRequest)
	if err := json.NewDecoder(r.Body).Decode(createUserReq); err != nil {
		return &ApiError{Err: "request error RE001", Status: http.StatusBadRequest}
	}
	defer r.Body.Close()

	token, err := s.userService.SignUp(createUserReq)
	if err != nil {
		return &ApiError{Err: "authentication error AE001", Status: http.StatusBadRequest}
	}

	return WriteJSON(w, http.StatusCreated, token)
}

func (s *ApiServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	loginRequest := new(LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(loginRequest); err != nil {
		return err
	}
	defer r.Body.Close()

	token, err := s.userService.Login(loginRequest)
	if err != nil {
		return &ApiError{Err: "authentication error AE002", Status: http.StatusForbidden}
	}

	return WriteJSON(w, http.StatusOK, token)
}

func (s *ApiServer) handleDeleteUserById(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	err := s.userService.DeleteUserById(id)
	if err != nil {
		return &ApiError{Err: "request error AE002", Status: http.StatusBadRequest}
	}

	return WriteJSON(w, http.StatusOK, nil)
}

func (s *ApiServer) handleGetAllUsers(w http.ResponseWriter, r *http.Request) error {
	users, err := s.userService.GetAllUsers()
	if err != nil {
		return &ApiError{Err: "request error AE003", Status: http.StatusBadRequest}
	}

	return WriteJSON(w, http.StatusOK, users)
}

func (s *ApiServer) handleGetUserById(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "id")
	user, err := s.userService.GetUserById(id)
	if err != nil {
		return &ApiError{Err: "Request Error AE004", Status: http.StatusNotFound}
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

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(AdminOnly)
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		perm, ok := ctx.Value("acl.permission").(YourPermissionType)
		if !ok || !perm.IsAdmin() {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// func (s *ApiServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
// 	id := chi.URLParam(r, "id")

// 	if r.Method == "GET" {
// 		return s.handleGetAccountByID(w, r, id)
// 	}

// 	if r.Method == "DELETE" {
// 		return s.handleDeleteAccountByID(w, r, id)
// 	}

// 	return fmt.Errorf("method not allowed: %s", r.Method)
// }

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

// func (s *ApiServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request, id string) error {
// 	acc, err := s.accountStore.GetAccountByID(id)
// 	if err != nil {
// 		return err
// 	}

// 	return WriteJSON(w, http.StatusOK, acc)
// }

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

// func withAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Println("In the Auth middleware")
// 		tokenString := r.Header.Get("Authorization")
// 		_, err := validateJWT(tokenString)
// 		if err != nil {
// 			WriteJSON(w, http.StatusUnauthorized, ApiError{Err: "Authentication Error AE003", Status: http.StatusForbidden})
// 			return
// 		}

// 		handlerFunc(w, r)
// 	}
// }

// func validateJWT(token string) (*jwt.Token, error) {
// 	secret := os.Getenv("JWT_SECRET")

// 	return jwt.Parse(token, func(token *jwt.Token) (any, error) {
// 		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 		}
// 		return []byte(secret), nil
// 	})
// }
