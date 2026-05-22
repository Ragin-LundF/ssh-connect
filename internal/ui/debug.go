package ui

import (
	"log"
	"os"
)

var debugLogger *log.Logger

// SetDebug toggles verbose UI interaction logs.
func SetDebug(enabled bool) {
	if enabled {
		debugLogger = log.New(os.Stderr, "[ui] ", log.LstdFlags|log.Lmicroseconds)
		debugf("debug logging enabled")
		return
	}
	debugLogger = nil
}

func debugf(format string, args ...interface{}) {
	if debugLogger == nil {
		return
	}
	debugLogger.Printf(format, args...)
}

