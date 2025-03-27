package logger

import (
	"io"
	"log"
	"os"
)

var DebugEnabled bool

func Debug(v ...any) {
	if DebugEnabled {
		log.Println("[DEBUG]", v)
	}
}

func Info(v ...any) {
	log.Println("[INFO]", v)
}

func Fatal(v ...any) {
	log.Println("[FATAL ERROR]", v)
	os.Exit(1)
}

func Error(v ...any) {
	log.Println("[ERROR]", v)
}

func Warning(v ...any) {
	log.Println("[WARNING]", v)
}

// StartCopyLogToFile sets up logging to write to both stdout and a log file.
// It returns the original logger output and the opened log file, so the caller can restore output and close the file.
func StartCopyLogToFile(fileName, path string) (originalOutput io.Writer, logFile *os.File, err error) {
	// Create a conversion-specific log file.
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		Error("Error creating directory", path, ":", err)
		return nil, nil, err
	}
	filePath := path + "/" + fileName
	logFile, err = os.Create(filePath)
	if err != nil {
		Error("Error creating README file at", filePath, ":", err)
		return nil, nil, err
	}

	// Save the original logger output.
	originalOutput = log.Writer()

	// Set the logger to write to both stdout and the log file.
	mw := io.MultiWriter(originalOutput, logFile)
	log.SetOutput(mw)

	return originalOutput, logFile, nil
}

// StopCopyLogToFile restores the logger's output and closes the log file.
func StopCopyLogToFile(originalOutput io.Writer, logFile *os.File) {
	log.SetOutput(originalOutput)
	logFile.Close()
}
