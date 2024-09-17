package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Balance float64 `json:"balance"`
}

var (
	users = make(map[int]User)
	mu    sync.Mutex
)

func initUsers() {
	// Инициализируем пользователей
	users[1] = User{ID: 1, Name: "Alice", Balance: 100.0}
	users[2] = User{ID: 2, Name: "Bob", Balance: 50.0}
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr) // Преобразование строки в int
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if user, ok := users[userID]; ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

func updateBalance(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if existingUser, ok := users[user.ID]; ok {
		existingUser.Balance += user.Balance
		users[user.ID] = existingUser
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
	}
}

func transferBalance(w http.ResponseWriter, r *http.Request) {
	var transfer struct {
		FromID int     `json:"from_id"`
		ToID   int     `json:"to_id"`
		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&transfer); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	fromUser, fromExists := users[transfer.FromID]
	toUser, toExists := users[transfer.ToID]

	if !fromExists || !toExists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if fromUser.Balance < transfer.Amount {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	// Выполняем перевод
	fromUser.Balance -= transfer.Amount
	toUser.Balance += transfer.Amount
	users[transfer.FromID] = fromUser
	users[transfer.ToID] = toUser

	w.WriteHeader(http.StatusOK)
}

func main() {
	initUsers()

	http.HandleFunc("/getBalance", getBalance)
	http.HandleFunc("/updateBalance", updateBalance)
	http.HandleFunc("/transferBalance", transferBalance)

	log.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
