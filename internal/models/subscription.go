package models

import "time"

type Subscription struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Status    string    `db:"status" json:"status"`
}

type SubscriptionRequest struct {
	Email string `json:"email"`
}

type SubscriptionResponse struct {
	Message string `json:"message"`
}
