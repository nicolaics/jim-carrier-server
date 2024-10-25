package listing

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

func (s *Store) CreateListing(listing types.Listing) error {
	values := "?"
	for i := 0; i < 6; i++ {
		values += ", ?"
	}

	query := `INSERT INTO listing (
					carrier_id, destination, weight_available, 
					price_per_kg, departure_date, exp_status, description) 
					VALUES (` + values + `)`

	_, err := s.db.Exec(query, listing.CarrierID, listing.Destination, listing.WeightAvailable,
		listing.PricePerKg, listing.DepartureDate, listing.ExpStatus,
		listing.Description)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetAllListings() ([]types.ListingReturnPayload, error) {
	err := s.UpdateListingExpStatus()
	if err != nil {
		return nil, fmt.Errorf("error updating listing status: %v", err)
	}

	query := `SELECT l.id, l.carrier_id, user.name, l.destination, 
					l.weight_available, l.price_per_kg, 
					l.departure_date, l.description, l.last_modified_at  
				FROM listing AS l 
				JOIN user ON user.id = l.carrier_id 
				WHERE l.exp_status = ? 
				AND l.deleted_at IS NULL 
				ORDER BY l.departure_date DESC`
	rows, err := s.db.Query(query, constants.EXP_STATUS_AVAILABLE)
	if err != nil {
		return nil, err
	}

	listings := make([]types.ListingReturnPayload, 0)

	for rows.Next() {
		listing, err := scanRowIntoListingReturn(rows)
		if err != nil {
			return nil, err
		}

		listings = append(listings, *listing)
	}

	return listings, nil
}

func (s *Store) UpdateListingExpStatus() error {
	query := `UPDATE listing SET exp_status = ? 
				WHERE (departure_date > ? OR weight_available <= 0) AND deleted_at IS NULL`
	_, err := s.db.Exec(query, constants.EXP_STATUS_EXPIRED, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) IsListingDuplicate(carrierId int, destination string, weightAvailable float64, departureDate time.Time) (bool, error) {
	query := `SELECT COUNT(*) FROM listing 
				WHERE carrier_id = ? AND destination = ? 
				AND weight_available = ?  
				AND departure_date = ? AND exp_stauts = ? 
				AND deleted_at IS NULL`
	row := s.db.QueryRow(query, carrierId, destination, weightAvailable,
		departureDate, constants.EXP_STATUS_AVAILABLE)
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

func (s *Store) GetListingByPayload(carrierName string, destination string, weightAvailable float64, pricePerKg float64, departureDate time.Time) (*types.ListingReturnPayload, error) {
	query := `SELECT l.id, l.carrier_id, user.name, l.destination, l.weight_available, 
					l.price_per_kg, l.departure_date, l.description, l.last_modified_at 
				FROM listing AS l 
				JOIN user ON user.id = l.carrier_id 
				WHERE user.name = ? AND l.destination = ? 
				AND l.weight_available = ? AND l.price_per_kg = ? 
				AND l.departure_date = ? AND l.exp_stauts = ? 
				AND l.deleted_at IS NULL`
	rows, err := s.db.Query(query, carrierName, destination, weightAvailable,
		pricePerKg, departureDate, constants.EXP_STATUS_AVAILABLE)
	if err != nil {
		return nil, err
	}

	listing := new(types.ListingReturnPayload)

	for rows.Next() {
		listing, err = scanRowIntoListingReturn(rows)
		if err != nil {
			return nil, err
		}
	}

	if listing.ID == 0 {
		return nil, fmt.Errorf("listing not found")
	}

	return listing, nil
}

func (s *Store) GetListingByID(id int) (*types.ListingReturnPayload, error) {
	query := `SELECT l.id, l.carrier_id, user.name, l.destination, l.weight_available, 
					l.price_per_kg, l.departure_date, l.description, l.last_modified_at 
				FROM listing AS l 
				JOIN user ON user.id = l.carrier_id 
				WHERE id = ? AND exp_stauts = ? 
				AND deleted_at IS NULL`
	rows, err := s.db.Query(query, id, constants.EXP_STATUS_AVAILABLE)
	if err != nil {
		return nil, err
	}

	listing := new(types.ListingReturnPayload)

	for rows.Next() {
		listing, err = scanRowIntoListingReturn(rows)
		if err != nil {
			return nil, err
		}
	}

	if listing.ID == 0 {
		return nil, fmt.Errorf("listing not found")
	}

	return listing, nil
}

func (s *Store) DeleteListing(id int) error {
	query := `UPDATE listing SET deleted_at = ? 
				WHERE id = ? AND deleted_at IS NULL`
	_, err := s.db.Exec(query, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ModifyListing(id int, listing types.Listing) error {
	query := `UPDATE listing 
				SET destination = ?, weight_available = ?, 
					price_per_kg = ?, departure_date = ?, exp_status = ?, 
					description = ?, last_modified_at = ? 
				WHERE id = ? AND delete_at IS NULL`

	_, err := s.db.Exec(query, listing.Destination, listing.WeightAvailable,
		listing.PricePerKg, listing.DepartureDate, listing.ExpStatus,
		listing.Description, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SubtractWeightAvailable(listingId int, minusValue float64) error {
	query := `SELECT weight_available FROM lising WHERE id = ? AND deleted_at IS NULL`
	row := s.db.QueryRow(query, listingId)
	if row.Err() != nil {
		return row.Err()
	}

	var oldWeightAvail float64
	err := row.Scan(&oldWeightAvail)
	if err != nil {
		return err
	}

	if (oldWeightAvail - minusValue) < 0.0 {
		return fmt.Errorf("weight available is not enough")
	}

	query = `UPDATE listing SET weight_available = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`
	_, err = s.db.Exec(query, (oldWeightAvail - minusValue), time.Now(), listingId)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) AddWeightAvailable(listingId int, addValue float64) error {
	query := `SELECT weight_available FROM lising WHERE id = ? AND deleted_at IS NULL`
	row := s.db.QueryRow(query, listingId)
	if row.Err() != nil {
		return row.Err()
	}

	var oldWeightAvail float64
	err := row.Scan(&oldWeightAvail)
	if err != nil {
		return err
	}

	query = `UPDATE listing SET weight_available = ?, last_modified_at = ? 
				WHERE id = ? AND deleted_at IS NULL`
	_, err = s.db.Exec(query, (oldWeightAvail + addValue), time.Now(), listingId)
	if err != nil {
		return err
	}

	return nil
}

// func scanRowIntoListing(rows *sql.Rows) (*types.Listing, error) {
// 	listing := new(types.Listing)

// 	err := rows.Scan(
// 		&listing.ID,
// 		&listing.CarrierID,
// 		&listing.Destination,
// 		&listing.WeightAvailable,
// 		&listing.PricePerKg,
// 		&listing.DepartureDate,
// 		&listing.ExpStatus,
// 		&listing.Description,
// 		&listing.CreatedAt,
// 		&listing.DeletedAt,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	listing.DepartureDate = listing.DepartureDate.Local()
// 	listing.CreatedAt = listing.CreatedAt.Local()

// 	return listing, nil
// }

func scanRowIntoListingReturn(rows *sql.Rows) (*types.ListingReturnPayload, error) {
	listing := new(types.ListingReturnPayload)

	err := rows.Scan(
		&listing.ID,
		&listing.CarrierID,
		&listing.CarrierName,
		&listing.Destination,
		&listing.WeightAvailable,
		&listing.PricePerKg,
		&listing.DepartureDate,
		&listing.Description,
		&listing.LastModifiedAt,
	)

	if err != nil {
		return nil, err
	}

	listing.LastModifiedAt = listing.LastModifiedAt.Local()

	return listing, nil
}
