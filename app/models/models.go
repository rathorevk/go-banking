package models

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username" validate:"required"`
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type Account struct {
	ID       int64   `json:"id"`
	UserID   string  `json:"user_id" validate:"required"`
	Balance  float64 `json:"balance" validate:"required"`
	Currency string  `json:"currency" default:"EUR" validate:"required,oneof=USD EUR GBP"`
	Status   string  `json:"status" default:"active"`
}

type Transaction struct {
	ID              string `json:"transactionId" validate:"required" db:"id,pk"`
	AccountID       int64  `json:"account_id" validate:"required" db:"account_id,index"`
	Amount          string `json:"amount" validate:"required" db:"amount"`
	AmountFloat     float64
	Source          string `json:"source" validate:"required,oneof=game server payment" db:"source"`
	TransactionType string `json:"state" validate:"required,oneof=win lose" db:"transaction_type"`
	InsertedAt      string `json:"inserted_at" db:"inserted_at"`
}

type UserBalance struct {
	UserID  int64  `json:"userId"`
	Balance string `json:"balance"`
}
