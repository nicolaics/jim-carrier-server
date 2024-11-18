package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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

func GeneratePaymentProofFileName(fileExtension string) string {
	// set the image file name
	rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNumberOne := GenerateRandomCodeNumbers(6)
	randomNumberTwo := GenerateRandomCodeNumbers(6)

	fileName := fmt.Sprintf("%s-%s%s", randomNumberOne, randomNumberTwo, fileExtension)

	return fileName
}
