package main

import (
	json "encoding/json"
	"net/http"
)

type User struct {
	Username string `json:"username"`
	Age      int    `json:"age"`
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users := []User{
		{Username: "John", Age: 30},
		{Username: "Jane", Age: 25},
	}

	response, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	if user.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Username is required"))
		return
	}

	response, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = enableCors(handler)

	mux.HandleFunc("GET /api/users", GetUsers)
	mux.HandleFunc("POST /api/user", CreateUser)

	println("server start at:8080")
	http.ListenAndServe(":8080", mux)
}
