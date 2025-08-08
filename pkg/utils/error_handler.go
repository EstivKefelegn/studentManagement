package utils

import (
	"fmt"
	"log"
	"os"
)

func ErrorHandler(err error, message string) error {
	error_logger := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	error_logger.Println(message, err)
	return fmt.Errorf("%s",message)
} 