package logger

import (
	"fmt"
	"time"
)

func Info(msg string) {
	fmt.Printf(
		"[%s] %s\n",
		time.Now().Format("15:04:05"),
		msg,
	)
}
