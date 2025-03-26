package logger

import (
	"log"
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
