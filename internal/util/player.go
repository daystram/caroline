package util

import (
	"fmt"
	"log"
)

func NewWorkerLogger(guildID, name string) func(v ...interface{}) {
	prefix := []interface{}{fmt.Sprintf("[%s] %s:", guildID, name)}
	return func(v ...interface{}) {
		log.Println(append(prefix, v...)...)
	}
}
