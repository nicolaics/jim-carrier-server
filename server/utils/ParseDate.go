package utils

import "time"

// dateStr must include the GMT
func ParseDate(dateStr string) (*time.Time, error) {
	dateFormat := "2006-01-02 -0700MST"
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return nil, err
	}

	return &date, nil
}

func ParseStartDate(dateStr string) (*time.Time, error) {
	dateFormat := "2006-01-02 -0700MST"
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return nil, err
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	return &date, nil
}

func ParseEndDate(dateStr string) (*time.Time, error) {
	dateFormat := "2006-01-02 -0700MST"
	date, err := time.Parse(dateFormat, dateStr)
	if err != nil {
		return nil, err
	}

	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	date = date.AddDate(0, 0, 1)

	return &date, nil
}