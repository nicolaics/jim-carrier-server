package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
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

func GeneratePictureFileName(fileExtension string) string {
	// set the image file name
	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumberOne := GenerateRandomCodeNumbers(6)
	randomNumberTwo := GenerateRandomCodeNumbers(6)

	fileName := fmt.Sprintf("%s-%s%s", randomNumberOne, randomNumberTwo, fileExtension)

	return fileName
}

func WrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text) // Split the text into words
	line := ""

	for _, word := range words {
		if (len(line) + len(word) + 1) > width { // Check if adding the word exceeds the width
			lines = append(lines, line) // Append the current line to the list
			line = word                 // Start a new line with the current word
		} else {
			if line != "" {
				line += " " // Add a space before appending the word
			}

			line += word
		}
	}

	if line != "" {
		lines = append(lines, line) // Append the last line
	}

	return lines
}
