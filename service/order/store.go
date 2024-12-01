package order

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nicolaics/jim-carrier/constants"
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

	query := `INSERT INTO order_list (
					listing_id, giver_id, weight, price,
					currency_id, package_content, package_img_url, notes, 
					order_confirmation_deadline) 
					VALUES (` + values + `)`

	deadline := time.Date(time.Now().Local().Year(), time.Now().Local().Month(), time.Now().Local().Day(), 0, 0, 0, 0, time.Now().Local().Location())
	deadline = deadline.AddDate(0, 0, 2)

	_, err := s.db.Exec(query, order.ListingID, order.GiverID, order.Weight,
		order.Price, order.CurrencyID, order.PackageContent, order.PackageImageURL,
		order.Notes, deadline)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOrderByID(id int) (*types.Order, error) {
	query := `SELECT * FROM order_list WHERE id = ? AND deleted_at IS NULL`
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

func (s *Store) GetOrdersByCarrierID(id int) ([]types.OrderCarrierReturnFromDB, error) {
	err := s.UpdateOrderStatusByDeadline()
	if err != nil {
		return nil, err
	}

	query := `SELECT l.id, l.destination, l.departure_date, 
					 o.id, 
					 user.name, user.phone_number, 
					 o.weight, o.price, 
					 c.name, 
					 o.package_content, o.package_img_url, 
					 o.payment_status, o.paid_at, 
					 o.payment_proof_url, 
					 o.order_status, o.package_location, 
					 o.notes, o.created_at, o.last_modified_at  
					FROM order_list AS o 
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

func (s *Store) GetOrdersByGiverID(id int) ([]types.OrderGiverReturnFromDB, error) {
	query := `SELECT o.id, o.weight, o.price, 
						c.name, 
						o.package_content, o.package_img_url, 
						o.payment_status, o.paid_at,
						o.payment_proof_url, 
						o.order_status, o.package_location, 
						o.notes, o.created_at, o.last_modified_at, 
						l.id, 
						user.name, 
						l.destination, l.departure_date 
					FROM order_list AS o 
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
					 o.package_content, o.package_img_url, 
					 o.payment_status, o.paid_at, 
					 o.payment_proof_url, 
					 o.order_status, o.package_location, 
					 o.notes, o.created_at, o.last_modified_at 
					FROM order_list AS o 
					JOIN listing AS l ON l.id = o.listing_id 
					JOIN user ON user.id = o.giver_id 
					JOIN currency AS c ON c.id = o.currency_id 
					WHERE o.id = ? 
					AND l.carrier_id = ? 
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
						o.package_content, o.package_img_url, 
						o.payment_status, o.paid_at, 
						o.payment_proof_url, 
						o.order_status, o.package_location, 
						o.notes, o.created_at, o.last_modified_at, 
						l.id, 
						user.name, 
						l.destination, l.departure_date 
					FROM order_list AS o 
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
	query := `UPDATE order_list SET deleted_at = ? WHERE id = ? AND giver_id = ?`
	_, err := s.db.Exec(query, time.Now(), orderId, userId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyOrder(id int, order types.Order) error {
	query := `UPDATE order_list SET weight = ?, price = ?, 
					currency_id = ?, package_content = ?, package_img_url = ?, 
					payment_status = ?, package_location = ?, order_confirmation_deadline = ?, 
					order_status = ?, notes = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

	deadline := time.Date(time.Now().Local().Year(), time.Now().Local().Month(), time.Now().Local().Day(), 0, 0, 0, 0, time.Now().Local().Location())
	deadline = deadline.AddDate(0, 0, 2)

	_, err := s.db.Exec(query, order.Weight, order.Price, order.CurrencyID,
		order.PackageContent, order.PackageImageURL, order.PaymentStatus,
		order.PackageLocation, deadline, order.OrderStatus, order.Notes,
		time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdatePackageLocation(id int, orderStatus int, packageLocation string) error {
	if orderStatus == -1 {
		query := `UPDATE order_list SET package_location = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, packageLocation, time.Now(), id)
		if err != nil {
			return err
		}
	} else {
		query := `UPDATE order_list SET order_status = ?, package_location = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, orderStatus, packageLocation, time.Now(), id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) UpdatePaymentStatus(id int, paymentStatus int, paymentProofUrl string) error {
	var err error

	if paymentStatus == constants.PAYMENT_STATUS_COMPLETED {
		query := `UPDATE order_list SET payment_status = ?, paid_at = ?, payment_proof_url = ?, 
				last_modified_at = ? WHERE id = ? AND deleted_at IS NULL`
		_, err = s.db.Exec(query, paymentStatus, time.Now(), paymentProofUrl, time.Now(), id)
	} else {
		query := `UPDATE order_list SET payment_status = ?, last_modified_at = ? 
					WHERE id = ? AND deleted_at IS NULL`
		_, err = s.db.Exec(query, paymentStatus, time.Now(), id)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateOrderStatus(id int, orderStatus int, packageLocation string) error {
	if packageLocation != "" {
		query := `UPDATE order_list SET order_status = ?, package_location = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, orderStatus, packageLocation, time.Now(), id)
		if err != nil {
			return err
		}
	} else {
		query := `UPDATE order_list SET order_status = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`

		_, err := s.db.Exec(query, orderStatus, time.Now(), id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) IsOrderDuplicate(userId int, listingId int) (bool, error) {
	query := `SELECT COUNT(*) FROM order_list WHERE listing_id = ? AND giver_id = ? AND deleted_at IS NULL`

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

func (s *Store) IsPaymentProofURLExist(paymentProofUrl string) bool {
	query := `SELECT COUNT(*) FROM order_list WHERE payment_proof_url = ? 
											AND deleted_at IS NULL`

	row := s.db.QueryRow(query, paymentProofUrl)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true
	}

	return (count > 0)
}

func (s *Store) IsPackageImageURLExist(packageImgUrl string) bool {
	query := `SELECT COUNT(*) FROM order_list WHERE package_img_url = ? 
											AND deleted_at IS NULL`

	row := s.db.QueryRow(query, packageImgUrl)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return true
	}

	return (count > 0)
}

func (s *Store) UpdateOrderStatusByDeadline() error {
	query := `UPDATE order_list SET order_status = ?, last_modified_at = ? 
				WHERE order_confirmation_deadline < ? 
				AND order_status = ? 
				AND deleted_at IS NULL`

	_, err := s.db.Exec(query, constants.ORDER_STATUS_CANCELLED, time.Now(),
		time.Now().UTC().Format("2006-01-02 15:04:05"), constants.ORDER_STATUS_WAITING)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetOrdersByListingID(listingId int) ([]types.OrderBulk, error) {
	query := `SELECT o.id, user.email 
				FROM order_list AS o 
				JOIN listing AS l ON l.id = o.listing_id 
				JOIN user ON o.giver_id = user.id 
				WHERE l.id = ? 
				AND l.deleted_at IS NULL
				AND o.deleted_at IS NULL`
	rows, err := s.db.Query(query, listingId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderBulks := make([]types.OrderBulk, 0)

	for rows.Next() {
		orderBulk := new(types.OrderBulk)

		err = rows.Scan(&orderBulk.ID, &orderBulk.GiverEmail)
		if err != nil {
			return nil, err
		}

		orderBulks = append(orderBulks, *orderBulk)
	}

	return orderBulks, nil
}

func (s *Store) GetOrderID(order types.Order) (int, error) {
	query := `SELECT id  
				FROM order_list 
				WHERE listing_id = ? 
				AND giver_id = ? 
				AND weight = ? 
				AND price = ? 
				AND currency_id = ? 
				AND package_content = ? 
				AND package_img_url = ? 
				AND notes = ? 
				AND deleted_at IS NULL`
	row := s.db.QueryRow(query, order.ListingID, order.GiverID, order.Weight,
		order.Price, order.CurrencyID, order.PackageContent, order.PackageImageURL,
		order.Notes)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var id int

	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Store) GetOrderCountByCarrierID(id int) (int, error) {
	err := s.UpdateOrderStatusByDeadline()
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*)
				FROM order_list AS o 
				JOIN listing AS l ON l.id = o.listing_id 
				WHERE l.carrier_id = ? 
				AND o.deleted_at IS NULL 
				AND l.deleted_at IS NULL`
	row := s.db.QueryRow(query, id)
	if row.Err() != nil {
		return 0, row.Err()
	}

	var count int
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func scanRowIntoOrder(rows *sql.Rows) (*types.Order, error) {
	temp := new(struct {
		ID                        int            `json:"id"`
		ListingID                 int            `json:"listingId"`
		GiverID                   int            `json:"giverId"`
		Weight                    float64        `json:"weight"`
		Price                     float64        `json:"price"`
		CurrencyID                int            `json:"currencyId"`
		PackageContent            string         `json:"packageContent"`
		PackageImageURL           sql.NullString `json:"packageImageUrl"`
		PaymentStatus             int            `json:"paymentStatus"`
		PaidAt                    sql.NullTime   `json:"paidAt"`
		PaymentProofURL           sql.NullString `json:"paymentProofUrl"`
		OrderConfirmationDeadline time.Time      `json:"orderConfirmationDeadline"`
		OrderStatus               int            `json:"orderStatus"`
		PackageLocation           string         `json:"packageLocation"`
		Notes                     sql.NullString `json:"notes"`
		CreatedAt                 time.Time      `json:"createdAt"`
		LastModifiedAt            time.Time      `json:"lastModifiedAt"`
		DeletedAt                 sql.NullTime   `json:"deletedAt"`
	})

	err := rows.Scan(
		&temp.ID,
		&temp.ListingID,
		&temp.GiverID,
		&temp.Weight,
		&temp.Price,
		&temp.CurrencyID,
		&temp.PackageContent,
		&temp.PackageImageURL,
		&temp.PaymentStatus,
		&temp.PaidAt,
		&temp.PaymentProofURL,
		&temp.OrderConfirmationDeadline,
		&temp.OrderStatus,
		&temp.PackageLocation,
		&temp.Notes,
		&temp.CreatedAt,
		&temp.LastModifiedAt,
		&temp.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	order := &types.Order{
		ID:                        temp.ID,
		ListingID:                 temp.ListingID,
		GiverID:                   temp.GiverID,
		Weight:                    temp.Weight,
		Price:                     temp.Price,
		CurrencyID:                temp.CurrencyID,
		PackageContent:            temp.PackageContent,
		PackageImageURL:           temp.PackageImageURL.String,
		PaymentStatus:             temp.PaymentStatus,
		PaidAt:                    temp.PaidAt.Time,
		PaymentProofURL:           temp.PaymentProofURL.String,
		OrderConfirmationDeadline: temp.OrderConfirmationDeadline,
		OrderStatus:               temp.OrderStatus,
		PackageLocation:           temp.PackageLocation,
		Notes:                     temp.Notes.String,
		CreatedAt:                 temp.CreatedAt,
		LastModifiedAt:            temp.LastModifiedAt,
		DeletedAt:                 temp.DeletedAt,
	}

	order.CreatedAt = order.CreatedAt.Local()
	order.LastModifiedAt = order.LastModifiedAt.Local()
	order.OrderConfirmationDeadline = order.OrderConfirmationDeadline.Local()

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
		&order.PackageContent,
		&order.PackageImageURL,
		&order.PaymentStatus,
		&order.PaidAt,
		&order.PaymentProofURL,
		&order.OrderStatus,
		&order.PackageLocation,
		&order.Notes,
		&order.CreatedAt,
		&order.LastModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	if order.PaidAt.Valid {
		order.PaidAt.Time = order.PaidAt.Time.Local()
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
		&order.PackageContent,
		&order.PackageImageURL,
		&order.PaymentStatus,
		&order.PaidAt,
		&order.PaymentProofURL,
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

	if order.PaidAt.Valid {
		order.PaidAt.Time = order.PaidAt.Time.Local()
	}

	order.Listing.DepartureDate = order.Listing.DepartureDate.Local()
	order.CreatedAt = order.CreatedAt.Local()

	return order, nil
}
