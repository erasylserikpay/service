package models

type User struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    PhoneNumber string  `json:"phone_number"`
    Password    string  `json:"password,omitempty"`
    Balance     float64 `json:"balance"`
}

type Transaction struct {
    ID              int     `json:"id"`
    SenderID        int     `json:"sender_id"`
    ReceiverID      int     `json:"receiver_id"`
    Amount          float64 `json:"amount"`
    TransactionType string  `json:"transaction_type"`
} 