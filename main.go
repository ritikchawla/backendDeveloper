package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type User struct {
	Username string
	Password string
	Role     string
}

type Book struct {
	Name            string
	Author          string
	PublicationYear int
}

var secretKey = []byte("secret")

var users = []User{
	{Username: "admin", Password: "admin123", Role: "admin"},
	{Username: "user", Password: "user123", Role: "regular"},
}

// Sample book data
var regularUserBooks = []Book{
	{Name: "Book 1", Author: "Author 1", PublicationYear: 2020},
	{Name: "Book 2", Author: "Author 2", PublicationYear: 2019},
}

var adminUserBooks = []Book{
	{Name: "Book 3", Author: "Author 3", PublicationYear: 2018},
	{Name: "Book 4", Author: "Author 4", PublicationYear: 2017},
}

func main() {
	r := mux.NewRouter()

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	r.Use(corsMiddleware)

	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/home", homeHandler).Methods("GET")
	r.HandleFunc("/addBook", addBookHandler).Methods("POST")
	r.HandleFunc("/deleteBook", deleteBookHandler).Methods("DELETE")

	fmt.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", r)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user exists and credentials are correct
	validUser := false
	for _, u := range users {
		if u.Username == user.Username && u.Password == user.Password {
			validUser = true
			break
		}
	}
	if !validUser {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Get user role from JWT token
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	role, ok := claims["role"].(string)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	var books []Book

	if role == "admin" {
		books = append(books, regularUserBooks...)
		books = append(books, adminUserBooks...)
	} else {
		books = append(books, regularUserBooks...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func addBookHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var book Book
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate parameters
	if book.Name == "" || book.Author == "" || book.PublicationYear == 0 {
		http.Error(w, "Invalid book data", http.StatusBadRequest)
		return
	}

	// Add book to regularUserBooks
	regularUserBooks = append(regularUserBooks, book)

	w.WriteHeader(http.StatusCreated)
}

func deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	// Get user role from JWT token
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse book name from query parameter
	bookName := strings.ToLower(r.URL.Query().Get("name"))
	if bookName == "" {
		http.Error(w, "Missing book name parameter", http.StatusBadRequest)
		return
	}

	// Find and delete the book from regularUserBooks
	for i, book := range regularUserBooks {
		if strings.ToLower(book.Name) == bookName {
			regularUserBooks = append(regularUserBooks[:i], regularUserBooks[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
