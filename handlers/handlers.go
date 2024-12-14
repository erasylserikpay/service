package handlers

import (
	"banking-app/models"
	"database/sql"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec("INSERT INTO users (name, phone_number, password_hash) VALUES ($1, $2, $3)",
		user.Name, user.PhoneNumber, string(hashedPassword))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		PhoneNumber string `json:"phone_number"`
		Password    string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user models.User
	var hashedPassword string
	err := h.db.QueryRow("SELECT id, name, phone_number, password_hash, balance FROM users WHERE phone_number = $1",
		credentials.PhoneNumber).Scan(&user.ID, &user.Name, &user.PhoneNumber, &hashedPassword, &user.Balance)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func (h *Handler) TopUpBalance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2",
		req.Amount, req.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SenderID    int     `json:"sender_id"`
		PhoneNumber string  `json:"receiver_phone"`
		Amount      float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var receiverID int
	err = tx.QueryRow("SELECT id FROM users WHERE phone_number = $1", req.PhoneNumber).Scan(&receiverID)
	if err != nil {
		http.Error(w, "Receiver not found", http.StatusNotFound)
		return
	}

	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1",
		req.Amount, req.SenderID)
	if err != nil {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2",
		req.Amount, receiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO transactions (sender_id, receiver_id, amount, transaction_type) VALUES ($1, $2, $3, 'transfer')",
		req.SenderID, receiverID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	var balance float64
	err := h.db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
}
