package main

import (
	"banking-app/config"
	"banking-app/handlers"
	"log"
	"net/http"
)

func main() {
	db := config.InitDB()
	defer db.Close()

	h := handlers.NewHandler(db)

	http.HandleFunc("/api/register", h.Register)
	http.HandleFunc("/api/login", h.Login)
	http.HandleFunc("/api/topup", h.TopUpBalance)
	http.HandleFunc("/api/transfer", h.Transfer)
	http.HandleFunc("/api/balance/", h.GetBalance)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
