package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func SaveProfilePicture(id int, imageData []byte, fileExtension string) (string, error) {
	if err := os.MkdirAll("./static/profile_img/", 0744); err != nil {
		return "", err
	}

	// 20MB
	maxBytes := 20 << 20 // 20MB in bytes

	// check for image size
	if len(imageData) > maxBytes {
		return "", fmt.Errorf("the image size exceeds the limit of 20MB")
	}

	// set the image file name
	randomNumber := GenerateRandomCodeNumbers(12)
	fileName := fmt.Sprintf("%s-%d%s", randomNumber, id, fileExtension)
	imagePath := "./static/profile_img/" + fileName

	pattern := fmt.Sprintf(`-%d.(jpg|png)$`, id)
	re := regexp.MustCompile(pattern)

	folderPath := "./static/profile_img"

	// Read all files in the specified folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			if re.MatchString(file.Name()) {
				// Full path of the file
				fullPath := filepath.Join(folderPath, file.Name())

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
	if err := os.MkdirAll("./static/payment_proof/", 0744); err != nil {
		return err
	}

	// 20MB
	maxBytes := 20 << 20 // 20MB in bytes

	// check for image size
	if len(imageData) > maxBytes {
		return fmt.Errorf("the image size exceeds the limit of 20MB")
	}

	// set the image file name
	imagePath := "./static/payment_proof/" + fileName

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