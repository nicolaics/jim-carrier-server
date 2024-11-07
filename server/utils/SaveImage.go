package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nicolaics/jim-carrier/constants"
)

func SaveProfilePicture(id int, imageData []byte, fileExtension string) (string, error) {
	if err := os.MkdirAll(constants.PROFILE_IMG_DIR_PATH, 0744); err != nil {
		return "", err
	}

	// 5MB
	maxBytes := 5 << 20 // 5MB in bytes

	// check for image size
	if len(imageData) > maxBytes {
		return "", fmt.Errorf("the image size exceeds the limit of 5MB")
	}

	// set the image file name
	randomNumber := GenerateRandomCodeNumbers(12)
	fileName := fmt.Sprintf("%s-%d%s", randomNumber, id, fileExtension)
	imagePath := constants.PROFILE_IMG_DIR_PATH + fileName

	pattern := fmt.Sprintf(`-%d.(jpg|png)$`, id)
	re := regexp.MustCompile(pattern)

	// Read all files in the specified folder
	files, err := os.ReadDir(constants.PROFILE_IMG_DIR_PATH)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			if re.MatchString(file.Name()) {
				// Full path of the file
				fullPath := filepath.Join(constants.PROFILE_IMG_DIR_PATH, file.Name())

				// Delete the previous profile picture
				err := os.Remove(fullPath)
				if err != nil {
					return "", fmt.Errorf("error delete old profile picture: %v", err)
				}

				break
			}
		}
	}

	// create the empty file for the image
	file, err := os.Create(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// save the image data
	_, err = file.Write(imageData)
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

func SavePaymentProof(imageData []byte, fileName string) error {
	if err := os.MkdirAll(constants.PAYMENT_PROOF_DIR_PATH, 0744); err != nil {
		return err
	}

	// 10MB
	maxBytes := 10 << 20 // 10MB in bytes

	// check for image size
	if len(imageData) > maxBytes {
		return fmt.Errorf("the image size exceeds the limit of 10MB")
	}

	// set the image file name
	imagePath := constants.PAYMENT_PROOF_DIR_PATH + fileName

	// create the empty file for the image
	file, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// save the image data
	_, err = file.Write(imageData)
	if err != nil {
		return err
	}

	return nil
}