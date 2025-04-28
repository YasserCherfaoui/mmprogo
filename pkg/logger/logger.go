package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLogger() {
	// Create logs directory if it doesn't exist
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Open log files
	infoFile, err := os.OpenFile(filepath.Join(logDir, "info.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open info log file:", err)
	}

	warningFile, err := os.OpenFile(filepath.Join(logDir, "warning.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open warning log file:", err)
	}

	errorFile, err := os.OpenFile(filepath.Join(logDir, "error.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open error log file:", err)
	}

	// Initialize loggers
	Info = log.New(io.MultiWriter(os.Stdout, infoFile),
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(io.MultiWriter(os.Stdout, warningFile),
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(os.Stderr, errorFile),
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
