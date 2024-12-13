package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nicolaics/jim-carrier-server/config"
	"github.com/nicolaics/jim-carrier-server/constants"
)

func SaveProfilePicture(id int, imageData []byte, fileExtension, oldFileName string, bucket *s3.S3) (string, error) {
	// set the image file name
	randomNumber := GenerateRandomCodeNumbers(12)
	fileName := fmt.Sprintf("%s-%d%s", randomNumber, id, fileExtension)
	imagePath := constants.PROFILE_IMG_DIR_PATH + fileName

	bucket.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.Envs.BucketName),
		Key:    aws.String(oldFileName),
	})

	_, err := bucket.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.Envs.BucketName),
		Key:    aws.String(imagePath),
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		return "", fmt.Errorf("error uploading file: %v", err)
	}

	return imagePath, nil
}

func SaveImage(imageData []byte, filePath string, bucket *s3.S3) error {
	_, err := bucket.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(config.Envs.BucketName),
		Key:    aws.String(filePath),
		Body:   bytes.NewReader(imageData),
	})
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
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
