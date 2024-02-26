package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "0000"
	DB_NAME     = "postgres"
	HOST        = "localhost"
)
const (
	host     = HOST
	port     = 5432
	user     = DB_USER
	password = DB_PASSWORD
	dbname   = DB_NAME
)

var db *sql.DB

type GeneratedValue struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

func main() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := mux.NewRouter()
	router.HandleFunc("/generate", GenerateValue).Methods("POST")
	router.HandleFunc("/retrieve/{id}", GetValueByID).Methods("GET")

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func GenerateValue(w http.ResponseWriter, r *http.Request) {
	// Генерируем случайное значение и уникальный ID
	value := generateRandomValue()
	id := rand.Intn(1000) // Генерируем случайное число для ID

	// Сохраняем значение и ID в базе данных
	_, err := db.Exec("INSERT INTO generated_values (id, value) VALUES ($1, $2)", id, value)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := GeneratedValue{ID: id, Value: value}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateRandomValue() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 10)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}
func GetValueByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "ID not provided", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var value string
	err = db.QueryRow("SELECT value FROM generated_values WHERE id = $1", id).Scan(&value)
	if err != nil {
		http.Error(w, "Value not found", http.StatusNotFound)
		return
	}

	response := map[string]string{"value": value}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
