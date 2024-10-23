package review

import (
	"database/sql"
	// "fmt"
	// "time"

	"github.com/nicolaics/jim-carrier/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateReview(review types.Review) error {
	query := `INSERT INTO review 
				(order_id, reviewer_id, reviewee_id, content, rating, review_type) 
				VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, review.OrderID, review.ReviewerID, review.RevieweeID,
						review.Content, review.Rating, review.ReviewType)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IsReviewDuplicate(reviewerId, revieweeId, orderId int) (bool, error) {
	query := `SELECT COUNT(*) FROM review WHERE reviewer_id = ? AND reviewee_id = ? AND order_id = ?`
	row := s.db.QueryRow(query, reviewerId, revieweeId, orderId)
	if row.Err() != nil {
		return true, row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return count > 0, nil
}