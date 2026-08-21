package main

import (
	"github.com/mmk31585/workout-tracker/internal/config"
)

func main() {
	_, err := config.Load()
	if err != nil {
		return
	}
}
