package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/nicolaics/jim-carrier-server/constants"
)

func SaveProfilePicture(id int, imageData []byte, fileExtension string) (string, error) {
	if err := os.MkdirAll(constants.PROFILE_IMG_DIR_PATH, 0744); err != nil {
		return "", err
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

func SavePaymentProof(imageData []byte, filePath string) error {
	if err := os.MkdirAll(constants.PAYMENT_PROOF_DIR_PATH, 0744); err != nil {
		return err
	}

	// create the empty file for the image
	file, err := os.Create(filePath)
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

func SavePackageImage(imageData []byte, filePath string) error {
	if err := os.MkdirAll(constants.PACKAGE_IMG_DIR_PATH, 0744); err != nil {
		return err
	}

	// create the empty file for the image
	file, err := os.Create(filePath)
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

func DownloadImage(srcURL string) ([]byte, string, error) {
	resp, err := http.Head(srcURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var fileExt string
	contentType := resp.Header.Get("Content-Type")
	switch contentType {
	case "image/jpeg":
		fileExt = ".jpg"
	case "image/png":
		fileExt = ".png"
	default:
		return nil, "", fmt.Errorf("failed to identify picture extension: %s", contentType)
	}

	// Step 3: Download the image with a GET request
	resp, err = http.Get(srcURL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", err
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return imageData, fileExt, nil
}
