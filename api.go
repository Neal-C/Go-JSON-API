package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwtV5 "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store Storage
}

type ApiFunc func(http.ResponseWriter, *http.Request) error;

type APIError struct {
	Error string `json:"error"`
}



func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store: store,
	};
}

//Gorilla mux: to get url parsing, parameters and such
//Gorilla is a minimal package, for http "multiplexing"
func (selfAPIServer *APIServer) Run() {

	router := mux.NewRouter();

	router.HandleFunc("/login", makeHTTPHandlerFunc(selfAPIServer.handleLogin)).Methods("POST");

	router.HandleFunc("/account", makeHTTPHandlerFunc(selfAPIServer.handleCreateAccount)).Methods("POST");

	router.HandleFunc("/account", withJWTAuth(makeHTTPHandlerFunc(selfAPIServer.handleGetAccountById), selfAPIServer.store)).Methods("GET");

	router.HandleFunc("/accounts", makeHTTPHandlerFunc(selfAPIServer.handleGetAccounts)).Methods("GET");

	router.HandleFunc("/transfer", makeHTTPHandlerFunc(selfAPIServer.handleTransfer)).Methods("PATCH");

	router.HandleFunc("/account", makeHTTPHandlerFunc(selfAPIServer.handleDeleteAccount)).Methods("DELETE");

	log.Println("JSON API server running on port: ", selfAPIServer.listenAddr)

	log.Fatal(http.ListenAndServe(selfAPIServer.listenAddr, router));

}

//Never forget to close your connections, it's not done automatically

//617 fart
func (selfAPIServer *APIServer) handleLogin(responseWriter http.ResponseWriter, request *http.Request) error{

	defer request.Body.Close();

	var loginRequest LoginRequest;

	if err := json.NewDecoder(request.Body).Decode(&loginRequest); err != nil {
		return err;
	};

	account , err := selfAPIServer.store.GetAccountByNumber(int(loginRequest.Number));
		
	if err != nil {
		return err;
	}

	if !account.ValidatePassword(loginRequest.Password){
		return fmt.Errorf("😑😑😑😑😑😑😑");
	};

	tokenString, err := createJWT(account);

	if err != nil {
		return err;
	}

	loginResponse := LoginResponse{
		Number: loginRequest.Number,
		Token: tokenString,
	};




	return WriteJSON(responseWriter, http.StatusAccepted, loginResponse);
}

func (selfAPIServer *APIServer) handleCreateAccount(responseWriter http.ResponseWriter, request *http.Request) error {

	defer request.Body.Close();

	createAccountRequest := CreateAccountRequest{};

	if err := json.NewDecoder(request.Body).Decode(&createAccountRequest); err != nil {
		return err
	};

	account, err := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName, createAccountRequest.Password);

	if err != nil {
		return err;
	}

	if err := selfAPIServer.store.CreateAccount(account); err != nil {
		return err;
	};

	return WriteJSON(responseWriter, http.StatusCreated, account);
}

func (selfAPIServer *APIServer) handleGetAccountById(responseWriter http.ResponseWriter, request *http.Request) error {

	defer request.Body.Close();

	id, err := getID(request);

	if err != nil {
		return err;
	}

	account, err := selfAPIServer.store.GetAccountById(id);
	if err != nil {
		return err;
	}

	return WriteJSON(responseWriter, http.StatusOK, account);
}

func (selfAPIServer *APIServer) handleGetAccounts(responseWriter http.ResponseWriter, request *http.Request) error {

	defer request.Body.Close();
	
	accounts, err := selfAPIServer.store.GetAccounts();
	if err != nil {
		return err;
	}

	return WriteJSON(responseWriter, http.StatusOK, accounts);

};

func (selfAPIServer *APIServer) handleTransfer(responseWriter http.ResponseWriter, request *http.Request) error {

	defer request.Body.Close();

	transferRequest := TransferRequest{};

	if err := json.NewDecoder(request.Body).Decode(&transferRequest); err != nil {
		return err;
	}

	return WriteJSON(responseWriter, http.StatusAccepted, transferRequest);
}

func (selfAPIServer *APIServer) handleDeleteAccount(responseWriter http.ResponseWriter, request *http.Request) error {
	defer request.Body.Close();

	id, err := getID(request);

	if err != nil {
		return err;
	}

	if err := selfAPIServer.store.DeleteAccount(id); err != nil {
		return err;
	}
	return WriteJSON(responseWriter, http.StatusAccepted, map[string]int{"deleted": id} );
}



func WriteJSON(responseWriter http.ResponseWriter, status int, v any) error {
	responseWriter.Header().Add("Content-Type", "application/json");
	responseWriter.WriteHeader(status);
	return json.NewEncoder(responseWriter).Encode(v);
}

func createJWT(account *Account)(string, error){
	
	secret := os.Getenv("JWT_SECRET");

	// Create the Claims
	claims := &jwtV5.MapClaims{
		"ExpiresAt": jwtV5.NewNumericDate(time.Unix(1516239022, 0)),
		"accountNumber": account.Number,
	}

	token := jwtV5.NewWithClaims(jwtV5.SigningMethodHS256, claims);

	signedString, err := token.SignedString([]byte(secret));

	fmt.Printf("%v %v \n", signedString, err);

	return signedString, err;
}

func permissionDenied(responseWriter http.ResponseWriter){
	WriteJSON(responseWriter, http.StatusForbidden, APIError{Error:"permission denied"});
}

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFeHBpcmVzQXQiOjE1MTYyMzkwMjIsImFjY291bnROdW1iZXIiOjMyMn0.3h6-mcJZoUEXz1MjdYb20layoFcIZ1ejgbCVut0Ap9U

func withJWTAuth(handlerFn http.HandlerFunc, store Storage) http.HandlerFunc {



	return func(responseWriter http.ResponseWriter, request *http.Request){
		fmt.Println("went through 'withJWTAuth' middleware");

		jwtString := request.Header.Get("x-jwt-token");

		token, err := validateJWT(jwtString);

		if err != nil {
			permissionDenied(responseWriter)
			return;
		}

		if !token.Valid {
			permissionDenied(responseWriter);
			return;
		}

		userID, err := getID(request);

		if err != nil {
			permissionDenied(responseWriter);
			return;
		}

		account , err := store.GetAccountById(userID);

		if err != nil {
			permissionDenied(responseWriter);
			return;
		};
		//so far ok, check below

		//typecasting to jwtV5.MapClaims
		claims := token.Claims.(jwtV5.MapClaims);

		//typecasting to float64, then converting to int64
		if account.Number != int64((claims["accountNumber"]).(float64)) {
			permissionDenied(responseWriter);
			return;
		}

		
		handlerFn(responseWriter, request);
	}
}

func validateJWT(jwtString string)(*jwtV5.Token, error){

	secret := os.Getenv("JWT_SECRET");

	token, err := jwtV5.Parse(jwtString, func(token *jwtV5.Token) (interface{}, error){
		if _, ok := token.Method.(*jwtV5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]);
		}
		return []byte(secret), nil;
	});


	return token, err;
}

func makeHTTPHandlerFunc(fn ApiFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request){
		if err := fn(responseWriter, request); err != nil {
			WriteJSON(responseWriter, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func getID(request *http.Request) (int, error){
	
	idStr := request.URL.Query().Get("id");

	id, err := strconv.Atoi(idStr);

	if err != nil {
		return id, fmt.Errorf("😭😭 invalid id 😭😭");
	}

	return id, nil;
}