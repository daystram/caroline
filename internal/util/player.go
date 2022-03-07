package util

import (
	"fmt"
	"log"
)

func NewPlayerWorkerLogger(guildID string) func(v ...interface{}) {
	prefix := []interface{}{fmt.Sprintf("[%s] player:", guildID)}
	return func(v ...interface{}) {
		log.Println(append(prefix, v...)...)
	}
}
