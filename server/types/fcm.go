package types

import "time"

type FCMHistoryStore interface {
	CreateFCMHistory(fcmHistory FCMHistory) error
}

type FCMHistory struct {
	ID        int       `json:"id"`
	ToUserID  int       `json:"toUserId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"imageUrl"`
	Link      string    `json:"link"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"createdAt"`
}