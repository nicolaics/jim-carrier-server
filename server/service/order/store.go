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
	for i := 0; i < 8; i++ {
		values += ", ?"
	}

	query := `INSERT INTO order (
					listing_id, giver_id, weight, price,
					currency_id, payment_status, order_status, 
					package_location, notes) 
					VALUES (` + values + `)`

	_, err := s.db.Exec(query, order.ListingID, order.GiverID, order.Weight,
		order.Price, order.CurrencyID, order.PaymentStatus,
		order.OrderStatus, order.PackageLocation, order.Notes)
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

func (s *Store) GetOrderByCarrierID(id int) ([]types.OrderCarrierReturnFromDB, error) {
	query := `SELECT l.id, l.destination, l.departure_date, 
					 o.id, 
					 user.name, user.phone_number, 
					 o.weight, o.price, 
					 c.name, 
					 o.payment_status, 
					 o.order_status, o.package_location, 
					 o.notes, o.created_at, o.last_modified_at  
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = o.giver_id 
					JOIN currency AS c ON c.id = o.currency_id 
					WHERE l.carrier_id = ? 
					AND o.deleted_at IS NULL 
					AND l.deleted_at IS NULL 
					ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]types.OrderCarrierReturnFromDB, 0)

	for rows.Next() {
		order, err := scanRowIntoOrderForCarrier(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, *order)
	}

	return orders, nil
}

func (s *Store) GetOrderByGiverID(id int) ([]types.OrderGiverReturnFromDB, error) {
	query := `SELECT o.id, o.weight, o.price, 
						c.name, 
						o.payment_status, 
						o.order_status, o.package_location, 
						o.notes, o.created_at, o.last_modified_at, 
						l.id, 
						user.name, 
						l.destination, l.departure_date 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = l.carrier_id  
					JOIN currency AS c ON c.id = o.currency_id 
					WHERE o.giver_id = ? 
					AND o.deleted_at IS NULL 
					AND l.deleted_at IS NULL 
					ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]types.OrderGiverReturnFromDB, 0)

	for rows.Next() {
		order, err := scanRowIntoOrderForGiver(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, *order)
	}

	return orders, nil
}

func (s *Store) GetCarrierOrderByID(orderId int, userId int) (*types.OrderCarrierReturnFromDB, error) {
	query := `SELECT l.id, l.destination, l.departure_date, 
					 o.id, 
					 user.name, user.phone_number, 
					 o.weight, o.price, 
					 c.name, 
					 o.payment_status, 
					 o.order_status, o.package_location, 
					 o.notes, o.created_at, o.last_modified_at 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = o.giver_id 
					JOIN currency AS c ON c.id = o.currency_id 
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

	order := new(types.OrderCarrierReturnFromDB)

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

func (s *Store) GetGiverOrderByID(orderId int, userId int) (*types.OrderGiverReturnFromDB, error) {
	query := `SELECT o.id, o.weight, o.price, 
						c.name, 
						o.payment_status, 
						o.order_status, o.package_location, 
						o.notes, o.created_at, o.last_modified_at, 
						l.id, 
						user.name, 
						l.destination, l.departure_date 
					FROM order AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = l.carrier_id  
					JOIN currency AS c ON c.id = o.currency_id 
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

	order := new(types.OrderGiverReturnFromDB)

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
					currency_id = ?, 
					payment_status = ?, package_location = ?, 
					order_status = ?, notes = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

	_, err := s.db.Exec(query, order.Weight, order.Price, order.CurrencyID, order.PaymentStatus,
		order.PackageLocation, order.OrderStatus, order.Notes,
		time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePackageLocation(id int, orderStatus int, packageLocation string) error {
	if orderStatus == -1 {
		query := `UPDATE order SET package_location = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, packageLocation, time.Now(), id)
		if err != nil {
			return err
		}
	} else {
		query := `UPDATE order SET order_status = ?, package_location = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, orderStatus, packageLocation, time.Now(), id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UpdatePaymentStatus(id int, paymentStatus int) error {
	query := `UPDATE order SET payment_status = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

	_, err := s.db.Exec(query, paymentStatus, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateOrderStatus(id int, orderStatus int, packageLocation string) error {
	if packageLocation != "" {
		query := `UPDATE order SET order_status = ?, package_location = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, orderStatus, packageLocation, time.Now(), id)
		if err != nil {
			return err
		}
	} else {
		query := `UPDATE order SET order_status = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, orderStatus, time.Now(), id)
		if err != nil {
			return err
		}
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
		&order.CurrencyID,
		&order.PaymentStatus,
		&order.OrderStatus,
		&order.PackageLocation,
		&order.Notes,
		&order.CreatedAt,
		&order.LastModifiedAt,
		&order.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	order.CreatedAt = order.CreatedAt.Local()
	order.LastModifiedAt = order.LastModifiedAt.Local()

	return order, nil
}

func scanRowIntoOrderForCarrier(rows *sql.Rows) (*types.OrderCarrierReturnFromDB, error) {
	order := new(types.OrderCarrierReturnFromDB)

	err := rows.Scan(
		&order.Listing.ID,
		&order.Listing.Destination,
		&order.Listing.DepartureDate,
		&order.ID,
		&order.GiverName,
		&order.GiverPhoneNumber,
		&order.Weight,
		&order.Price,
		&order.Currency,
		&order.PaymentStatus,
		&order.OrderStatus,
		&order.PackageLocation,
		&order.Notes,
		&order.CreatedAt,
		&order.LastModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	order.Listing.DepartureDate = order.Listing.DepartureDate.Local()
	order.CreatedAt = order.CreatedAt.Local()

	return order, nil
}

func scanRowIntoOrderForGiver(rows *sql.Rows) (*types.OrderGiverReturnFromDB, error) {
	order := new(types.OrderGiverReturnFromDB)

	err := rows.Scan(
		&order.ID,
		&order.Weight,
		&order.Price,
		&order.Currency,
		&order.PaymentStatus,
		&order.OrderStatus,
		&order.PackageLocation,
		&order.Notes,
		&order.CreatedAt,
		&order.LastModifiedAt,
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
