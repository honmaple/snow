package utils

import (
	"fmt"
	"time"
)

// DateFormats ..
var DateFormats = []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02"}

// parseTime ..
func ParseTime(value string) (time.Time, error) {
	for _, f := range DateFormats {
		if date, err := time.Parse(f, value); err == nil {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("date format error")
}
