package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func WriteServerLog(errorMessage any) error {
	logFolder := "static/log/server"
	if err := os.MkdirAll(logFolder, 0755); err != nil {
		return err
	}

	currentDate := time.Now().Format("060102-150405") // YYMMDD-HHmmss

	fileName := fmt.Sprintf("%s/%s.log", logFolder, currentDate)

	// open the log file
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// change the data into JSON
	jsonData, err := json.Marshal(errorMessage)
	if err != nil {
		return err
	}

	// store the data into the file
	_, err = file.WriteString(fmt.Sprintf("%s\n", jsonData))
	if err != nil {
		return err
	}

	return nil
}
