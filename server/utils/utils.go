package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func GenerateRandomCodeNumbers(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "0123456789"

	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

func GenerateRandomCodeAlphanumeric(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

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

	date = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	return &date, nil
}

func SaveProfilePicture(id int, imageData []byte, fileExtension string) (string, error) {
	// directory, err := filepath.Abs("static/profile_img/")
	// if err != nil {
	// 	return "", err
	// }

	// if err := os.MkdirAll(directory, 0744); err != nil {
	// 	return "", err
	// }

	// 20MB
	maxBytes := 20 << 20 // 20MB in bytes

	// check for image size
	if len(imageData) > maxBytes {
		return "", fmt.Errorf("the image size exceeds the limit of 20MB")
	}

	// set the image file name
	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumber := GenerateRandomCodeNumbers(12)
	imagePath := fmt.Sprintf("static/profile_img/%s-%d%s", randomNumber, id, fileExtension)

	pattern := fmt.Sprintf(`-%d.(jpg|png)$`, id)
	re := regexp.MustCompile(pattern)

	folderPath := "./static/profile_img"

	// Read all files in the specified folder
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", nil
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

	// err = ioutil.WriteFile(fileName, data.Image, 0644)
	// if err != nil {
	//     http.Error(w, "Could not save image", http.StatusInternalServerError)
	//     return
	// }

	return imagePath, nil
}
