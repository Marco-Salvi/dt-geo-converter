package logger

import (
	"fmt"
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
}

func Error(v ...any) {
	log.Println("[ERROR]", v)
}

func Warning(v ...any) {
	log.Println("[WARNING]", v)
}

// StartCopyLogToFile sets up logging to write to both stdout and a log file.
// It returns the original logger output and the opened log file, so the caller can restore output and close the file.
func StartCopyLogToFile(file string) (originalOutput io.Writer, logFile *os.File, err error) {
	// Create a conversion-specific log file.
	logFile, err = os.Create(file)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to create conversion log file: %w", err)
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
