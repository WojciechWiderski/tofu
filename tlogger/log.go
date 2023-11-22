package tlogger

import (
	"fmt"
	"os"
	"time"
)

const colorRed = "\033[38;5;160m"
const colorGreen = "\033[38;5;40m"
const colorYellow = "\033[38;5;220m"
const colorNone = "\033[0m"

func Error(msg string) {
	writeLog(msg, colorRed)
}

func Success(msg string) {
	writeLog(msg, colorGreen)
}

func Warn(msg string) {
	writeLog(msg, colorYellow)
}

func Info(msg string) {
	writeLog(msg, "")
}

func writeLog(msg string, color string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(os.Stdout, "%s %s %s %s \n", color, now, msg, colorNone)
}
