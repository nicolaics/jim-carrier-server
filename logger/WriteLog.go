package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nicolaics/jim-carrier-server/utils"
)

func WriteServerLog(errorMessage string) (string, error) {
	logFolder := "static/log/server"
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		log.Printf("error creating folder: %v", err)
		return "", err
	}

	currentDate := time.Now().Format("060102-150405") // YYMMDD-HHmmss

	fileName := fmt.Sprintf("%s-%s", currentDate, utils.GenerateRandomCodeAlphanumeric(6))

	filePath := fmt.Sprintf("%s/%s.log", logFolder, fileName)

	// open the log file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("error open file: %v", err)
		return "", err
	}
	defer file.Close()

	// store the data into the file
	msg := fmt.Sprintf("[Error %s]\n%s\n\n", time.Now().Format("2006/01/02 15:04:05"), errorMessage)
	_, err = file.WriteString(msg)
	if err != nil {
		log.Printf("error write string: %v", err)
		return "", err
	}

	return fileName, nil
}
