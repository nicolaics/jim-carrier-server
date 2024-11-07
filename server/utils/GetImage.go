package utils

import (
	"os"
)

func GetImage(imageURL string) ([]byte, error) {
	imageBytes, err := os.ReadFile(imageURL)
	if err != nil {
		return nil, err
	}

	return imageBytes, nil
}