package types

import "time"

type FCMHistoryStore interface {
	CreateFCMHistory(fcmHistory FCMHistory) error
}

type FCMData struct {
	TitleLoc string `json:"title_loc_key,omitempty"`
	BodyLoc  string `json:"body_loc_key,omitempty"`
	Type     string `json:"type,omitempty"`
}

type FCMHistory struct {
	ID        int       `json:"id"`
	ToUserID  int       `json:"toUserId"`
	ToToken   string    `json:"toToken"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	ImageURL  string    `json:"image"`
	Link      string    `json:"link"`
	Data      FCMData   `json:"data"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"createdAt"`
}
