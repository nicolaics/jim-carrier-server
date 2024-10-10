package order

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

func (s *Store) CreateOrder(order types.Order) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO order (
					listing_id, giver_id, weight, price,
					payment_status, order_status, notes) 
					VALUES (` + values + `)`

	_, err := s.db.Exec(query, order.ListingID, order.GiverID, order.Weight,
							order.Price, order.PaymentStatus,
							order.OrderStatus, order.Notes)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOrderByID(id int) (*types.Order, error) {
	query := `SELECT * FROM order WHERE id = ? AND deleted_at IS NULL`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order := new(types.Order)

	for rows.Next() {
		order, err = scanRowIntoOrder(rows)
		if err != nil {
			return nil, err
		}
	}

	if order.ID == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

func (s *Store) GetOrderByCarrierID(id int) ([]types.OrderCarrierReturnPayload, error) {
	query := `SELECT l.id, l.destination, l.departure_date, 
					 o.id, 
					 user.name, user.phone_number, 
					 o.weight, o.price, o.payment_status, 
					 o.order_status, o.notes, o.created_at 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = o.giver_id 
					WHERE l.carrier_id = ? 
					AND o.deleted_at IS NULL 
					AND l.deleted_at IS NULL 
					ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]types.OrderCarrierReturnPayload, 0)

	for rows.Next() {
		order, err := scanRowIntoOrderForCarrier(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, *order)
	}

	return orders, nil
}

func (s *Store) GetOrderByGiverID(id int) ([]types.OrderGiverReturnPayload, error) {
	query := `SELECT o.id, o.weight, o.price, o.payment_status, 
						o.order_status, o.notes, o.created_at, 
						l.id, 
						user.name, 
						l.destination, l.departure_date 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = l.carrier_id  
					WHERE o.giver_id = ? 
					AND o.deleted_at IS NULL 
					AND l.deleted_at IS NULL 
					ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]types.OrderGiverReturnPayload, 0)

	for rows.Next() {
		order, err := scanRowIntoOrderForGiver(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, *order)
	}

	return orders, nil
}

func (s *Store) GetCarrierOrderByID(orderId int, userId int) (*types.OrderCarrierReturnPayload, error) {
	query := `SELECT l.id, l.destination, l.departure_date, 
					 o.id, 
					 user.name, user.phone_number, 
					 o.weight, o.price, o.payment_status, 
					 o.order_status, o.notes, o.created_at 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = o.giver_id 
					WHERE o.id = ? 
					AND l.carried_id = ? 
					AND o.deleted_at IS NULL 
					AND l.deleted_at IS NULL 
					ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, orderId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order := new(types.OrderCarrierReturnPayload)

	for rows.Next() {
		order, err = scanRowIntoOrderForCarrier(rows)
		if err != nil {
			return nil, err
		}
	}

	if order.ID == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

func (s *Store) GetGiverOrderByID(orderId int, userId int) (*types.OrderGiverReturnPayload, error) {
	query := `SELECT o.id, o.weight, o.price, o.payment_status, 
						o.order_status, o.notes, o.created_at, 
						l.id, 
						user.name, 
						l.destination, l.departure_date 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = l.carrier_id  
					WHERE o.id = ? 
					AND o.giver_id = ? 
					AND o.deleted_at IS NULL 
					AND l.deleted_at IS NULL 
					ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, orderId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order := new(types.OrderGiverReturnPayload)

	for rows.Next() {
		order, err = scanRowIntoOrderForGiver(rows)
		if err != nil {
			return nil, err
		}
	}

	if order.ID == 0 {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

func (s *Store) DeleteOrder(orderId int, userId int) error {
	query := `UPDATE order SET deleted_at = ? WHERE id = ? AND giver_id = ?`
	_, err := s.db.Exec(query, time.Now(), orderId, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyOrder(id int, order types.Order) error {
	query := `UPDATE order SET weight = ?, price = ?,
					payment_status = ?, order_status = ?, notes = ? 
				WHERE id = ? AND deleted_at IS NULL`

	_, err := s.db.Exec(query, order.Weight, order.Price, order.PaymentStatus,
							order.OrderStatus, order.Notes, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IsOrderDuplicate(userId int, listingId int) (bool, error) {
	query := `SELECT COUNT(*) FROM order WHERE listing_id = ? 
											AND giver_id = ?
											AND deleted_at IS NULL`

	row := s.db.QueryRow(query, listingId, userId)
	if row.Err() != nil {
		return true, row.Err()
	}

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true, err
	}

	return (count > 0), nil
}


func scanRowIntoOrder(rows *sql.Rows) (*types.Order, error) {
	order := new(types.Order)

	err := rows.Scan(
		&order.ID,
		&order.ListingID,
		&order.GiverID,
		&order.Weight,
		&order.Price,
		&order.PaymentStatus,
		&order.OrderStatus,
		&order.Notes,
		&order.CreatedAt,
		&order.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	order.CreatedAt = order.CreatedAt.Local()

	return order, nil
}

func scanRowIntoOrderForCarrier(rows *sql.Rows) (*types.OrderCarrierReturnPayload, error) {
	order := new(types.OrderCarrierReturnPayload)

	err := rows.Scan(
		&order.Listing.ID,
		&order.Listing.Destination,
		&order.Listing.DepartureDate,
		&order.ID,
		&order.GiverName,
		&order.GiverPhoneNumber,
		&order.Weight,
		&order.Price,
		&order.PaymentStatus,
		&order.OrderStatus,
		&order.Notes,
		&order.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	order.Listing.DepartureDate = order.Listing.DepartureDate.Local()
	order.CreatedAt = order.CreatedAt.Local()

	return order, nil
}

func scanRowIntoOrderForGiver(rows *sql.Rows) (*types.OrderGiverReturnPayload, error) {
	order := new(types.OrderGiverReturnPayload)

	err := rows.Scan(
		&order.ID,
		&order.Weight,
		&order.Price,
		&order.PaymentStatus,
		&order.OrderStatus,
		&order.Notes,
		&order.CreatedAt,
		&order.Listing.ID,
		&order.Listing.CarrierName,
		&order.Listing.Destination,
		&order.Listing.DepartureDate,
	)

	if err != nil {
		return nil, err
	}

	order.Listing.DepartureDate = order.Listing.DepartureDate.Local()
	order.CreatedAt = order.CreatedAt.Local()

	return order, nil
}
