package types

import (
	"time"
)

type ReviewStore interface {
	CreateReview(Review) error

	GetReceivedReviewsByUserID(uid int) ([]ReceivedReviewReturnPayload, error)
	GeSentReviewsByUserID(uid int) ([]SentReviewReturnPayload, error)

	IsReviewDuplicate(reviewerId, revieweeId, orderId int) (bool, error)
}

type RegisterReviewPayload struct {
	OrderID      int    `json:"orderId" validate:"required"`
	RevieweeName string `json:"revieweeName" validate:"required"`
	Content      string `json:"content"`
	Rating       int    `json:"rating" validate:"required"`
}

type ReceivedReviewReturnPayload struct {
	ID                 int       `json:"id"`
	ReviewerID         int       `json:"reviewerId"`
	Content            string    `json:"content"`
	Rating             int       `json:"rating"`
	PackageDestination string    `json:"packageDestination"`
	DepartureDate      time.Time `json:"departureDate"`
	LastModifiedAt     time.Time `json:"lastModifiedAt"`
}

type SentReviewReturnPayload struct {
	ID                 int       `json:"id"`
	RevieweeName       string    `json:"revieweeName"`
	Content            string    `json:"content"`
	Rating             int       `json:"rating"`
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
	Rating         int       `json:"rating"`
	ReviewType     int       `json:"reviewType"`
	CreatedAt      time.Time `json:"createdAt"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}
