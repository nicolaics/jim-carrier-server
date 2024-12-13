package utils

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nicolaics/jim-carrier-server/config"
)

func GetImage(imageURL string, s3Bucket *s3.S3) ([]byte, error) {
	rawObject, err := s3Bucket.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(config.Envs.BucketName),
			Key:    aws.String(imageURL),
		})
	if err != nil {
		return nil, fmt.Errorf("error get object from s3: %v", err)
	}
	
	imageBytes, err := ioutil.ReadAll(rawObject.Body)
    if err != nil {
        return nil, fmt.Errorf("error read data: %v", err)
    }
	
	return imageBytes, nil
}
