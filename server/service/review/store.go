package review

import (
	"database/sql"
	"fmt"
	"time"

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

func (s *Store) GetReviewByID(id int) (*types.Review, error) {
	query := `SELECT * FROM review WHERE id = ?`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	review := new(types.Review)

	for rows.Next() {
		review, err = scanRowIntoReview(rows)
		if err != nil {
			return nil, err
		}
	}

	if review.ID == 0 {
		return nil, fmt.Errorf("review not found")
	}

	return review, nil
}

func (s *Store) GetReceivedReviewsByUserID(uid int) ([]types.ReceivedReviewReturnPayload, error) {
	query := `SELECT r.id, r.reviewer_id, r.content, r.rating, 
					l.destination, l.departure_date, 
					r.last_modified_at 
				FROM review AS r 
				JOIN order_list AS o ON r.order_id = o.id 
				JOIN listing AS l ON l.id = o.listing_id 
				WHERE r.reviewee_id = ? 
				ORDER BY o.created_at DESC`
	rows, err := s.db.Query(query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]types.ReceivedReviewReturnPayload, 0)

	for rows.Next() {
		review, err := scanRowIntoReceivedReview(rows)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, *review)
	}

	return reviews, nil
}

func (s *Store) GetSentReviewsByUserID(uid int) ([]types.SentReviewReturnPayload, error) {
	query := `SELECT r.id, 
					user.name, 
					r.content, r.rating, 
					l.destination, l.departure_date, 
					r.last_modified_at 
				FROM review AS r 
				JOIN order_list AS o ON r.order_id = o.id 
				JOIN listing AS l ON l.id = o.listing_id 
				JOIN user ON user.id = r.reviewee_id 
				WHERE r.reviewer_id = ? 
				ORDER BY o.created_at DESC`
	rows, err := s.db.Query(query, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]types.SentReviewReturnPayload, 0)

	for rows.Next() {
		review, err := scanRowIntoSentReview(rows)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, *review)
	}

	return reviews, nil
}

func (s *Store) DeleteReview(id int) error {
	query := `DELETE FROM review WHERE id = ?`
	_, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyReview(id int, content string, rating int) error {
	query := `UPDATE review SET content = ?, rating = ? WHERE id = ?`
	_, err := s.db.Exec(query, content, rating, id)
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

func (s *Store) GetAverageRating(userId int, reviewType int) (float64, error) {
	query := `SELECT AVG(r.rating) 
				FROM review AS r 
				JOIN order_list AS o ON r.order_id = o.id 
				JOIN listing AS l ON l.id = o.listing_id 
				WHERE r.reviewee_id = ? 
				AND r.review_type = ?
				AND o.deleted_at IS NULL 
				AND l.deleted_at IS NULL`
	row := s.db.QueryRow(query, userId, reviewType)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var avgRating sql.NullFloat64
	err := row.Scan(&avgRating)
	if err != nil {
		return 0, err
	}

	if !avgRating.Valid {
		return 0.0, nil
	}

	return avgRating.Float64, nil
}

func scanRowIntoReceivedReview(rows *sql.Rows) (*types.ReceivedReviewReturnPayload, error) {
	temp := new(struct {
		ID                 int
		ReviewerID         int
		Content            sql.NullString
		Rating             int
		PackageDestination string
		DepartureDate      time.Time
		LastModifiedAt     time.Time
	})

	err := rows.Scan(
		&temp.ID,
		&temp.ReviewerID,
		&temp.Content,
		&temp.Rating,
		&temp.PackageDestination,
		&temp.DepartureDate,
		&temp.LastModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	temp.LastModifiedAt = temp.LastModifiedAt.Local()
	temp.DepartureDate = temp.DepartureDate.Local()

	content := ""
	if temp.Content.Valid {
		content = temp.Content.String
	}

	receivedReview := types.ReceivedReviewReturnPayload{
		ID:                 temp.ID,
		ReviewerID:         temp.ReviewerID,
		Content:            content,
		Rating:             temp.Rating,
		PackageDestination: temp.PackageDestination,
		DepartureDate:      temp.DepartureDate,
		LastModifiedAt:     temp.LastModifiedAt,
	}

	return &receivedReview, nil
}

func scanRowIntoSentReview(rows *sql.Rows) (*types.SentReviewReturnPayload, error) {
	temp := new(struct {
		ID                 int
		RevieweeName       string
		Content            sql.NullString
		Rating             int
		PackageDestination string
		DepartureDate      time.Time
		LastModifiedAt     time.Time
	})

	err := rows.Scan(
		&temp.ID,
		&temp.RevieweeName,
		&temp.Content,
		&temp.Rating,
		&temp.PackageDestination,
		&temp.DepartureDate,
		&temp.LastModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	temp.LastModifiedAt = temp.LastModifiedAt.Local()
	temp.DepartureDate = temp.DepartureDate.Local()

	content := ""
	if temp.Content.Valid {
		content = temp.Content.String
	}

	sentReview := types.SentReviewReturnPayload{
		ID:                 temp.ID,
		RevieweeName:       temp.RevieweeName,
		Content:            content,
		Rating:             temp.Rating,
		PackageDestination: temp.PackageDestination,
		DepartureDate:      temp.DepartureDate,
		LastModifiedAt:     temp.LastModifiedAt,
	}

	return &sentReview, nil
}

func scanRowIntoReview(rows *sql.Rows) (*types.Review, error) {
	temp := new(struct {
		ID             int
		OrderID        int
		ReviewerID     int
		RevieweeID     int
		Content        sql.NullString
		Rating         int
		ReviewType     int
		CreatedAt      time.Time
		LastModifiedAt time.Time
	})

	err := rows.Scan(
		&temp.ID,
		&temp.OrderID,
		&temp.ReviewerID,
		&temp.RevieweeID,
		&temp.Content,
		&temp.Rating,
		&temp.ReviewType,
		&temp.CreatedAt,
		&temp.LastModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	temp.CreatedAt = temp.CreatedAt.Local()
	temp.LastModifiedAt = temp.LastModifiedAt.Local()

	content := ""
	if temp.Content.Valid {
		content = temp.Content.String
	}

	review := types.Review{
		ID:             temp.ID,
		OrderID:        temp.OrderID,
		ReviewerID:     temp.ReviewerID,
		RevieweeID:     temp.RevieweeID,
		Content:        content,
		Rating:         temp.Rating,
		ReviewType:     temp.ReviewType,
		CreatedAt:      temp.CreatedAt,
		LastModifiedAt: temp.LastModifiedAt,
	}

	return &review, nil
}
