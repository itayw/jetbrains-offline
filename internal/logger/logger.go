package logger

import (
	"fmt"
	"log"
	"os"
)

var debugMode = true

func init() {
	log.SetOutput(os.Stdout)
}

func Info(v ...interface{}) {
	log.Println("[INFO]", fmt.Sprint(v...))
}

func Error(v ...interface{}) {
	log.Println("[ERROR]", fmt.Sprint(v...))
}

func Debug(v ...interface{}) {
	if debugMode {
		log.Println("[DEBUG]", fmt.Sprint(v...))
	}
}
