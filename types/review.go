package types

import (
	"time"
)

type ReviewStore interface {
	CreateReview(Review) error

	GetReviewByID(id int) (*Review, error)

	GetReceivedReviewsByUserID(uid int) ([]ReceivedReviewReturnPayload, error)
	GetSentReviewsByUserID(uid int) ([]SentReviewReturnPayload, error)

	DeleteReview(id int) error

	ModifyReview(id int, content string, rating float64) error

	IsReviewDuplicate(reviewerId, revieweeId, orderId int) (bool, error)

	GetAverageRating(userId int, reviewType int) (float64, error)
}

type RegisterReviewPayload struct {
	OrderID      int     `json:"orderId" validate:"required"`
	RevieweeName string  `json:"revieweeName" validate:"required"`
	Content      string  `json:"content"`
	Rating       float64 `json:"rating" validate:"required"`
}

type DeleteReviewPayload struct {
	ID int `json:"id" validate:"required"`
}

type ModifyReviewPayload struct {
	ID      int     `json:"id" validate:"required"`
	Content string  `json:"content"`
	Rating  float64 `json:"rating" validate:"required"`
}

type ReceivedReviewPayload struct {
	CarrierID int `json:"carrierId" validate:"required"`
}

type ReceivedReviewReturnPayload struct {
	ID                 int       `json:"id"`
	ReviewerID         int       `json:"reviewerId"`
	Content            string    `json:"content"`
	Rating             float64   `json:"rating"`
	PackageDestination string    `json:"packageDestination"`
	DepartureDate      time.Time `json:"departureDate"`
	LastModifiedAt     time.Time `json:"lastModifiedAt"`
}

type SentReviewReturnPayload struct {
	ID                 int       `json:"id"`
	RevieweeName       string    `json:"revieweeName"`
	Content            string    `json:"content"`
	Rating             float64   `json:"rating"`
	PackageDestination string    `json:"packageDestination"`
	DepartureDate      time.Time `json:"departureDate"`
	LastModifiedAt     time.Time `json:"lastModifiedAt"`
}

type Review struct {
	ID             int       `json:"id"`
	OrderID        int       `json:"orderId"`
	ReviewerID     int       `json:"reviewerId"`
	RevieweeID     int       `json:"revieweeId"`
	Content        string    `json:"content"`
	Rating         float64   `json:"rating"`
	ReviewType     int       `json:"reviewType"`
	CreatedAt      time.Time `json:"createdAt"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}
