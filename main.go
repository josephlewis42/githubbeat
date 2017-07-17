package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/jlevesy/githubbeat/beater"
)

func main() {
	err := beat.Run("githubbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
