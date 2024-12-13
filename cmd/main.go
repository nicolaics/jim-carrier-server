package main

import (
	"database/sql"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier-server/cmd/api"
	"github.com/nicolaics/jim-carrier-server/config"
	"github.com/nicolaics/jim-carrier-server/db"
)

func main() {
	db, err := db.NewMySQLStorage(mysql.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})
	if err != nil {
		log.Fatal(err)
	}

	initStorage(db)
	s3, err := initS3Bucket()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()

	server := api.NewAPIServer((":" + config.Envs.Port), db, router, s3)

	// check the error, if error is not nill
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}

func initS3Bucket() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Envs.AWSRegion),
	})
	if err != nil {
		return nil, err
	}
	
	svc := s3.New(sess)

	return svc, nil
	// filePath := "Donuts.png"

	// file, err := os.Open(filePath)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error opening file:", err)
	// 	return
	// }
	// defer file.Close()

	// key := "Donuts.png"

	// Read the contents of the file into a buffer
	// var buf bytes.Buffer
	// if _, err := io.Copy(&buf, file); err != nil {
	// 	fmt.Fprintln(os.Stderr, "Error reading file:", err)
	// 	return
	// }

	// This uploads the contents of the buffer to S3
	// _, err = svc.PutObject(&s3.PutObjectInput{
	// 	Bucket: aws.String(bucket),
	// 	Key:    aws.String(key),
	// 	Body:   bytes.NewReader(buf.Bytes()),
	// })
	// if err != nil {
	// 	fmt.Println("Error uploading file:", err)
	// 	return
	// }

	// fmt.Println("File uploaded successfully!!!")
}
