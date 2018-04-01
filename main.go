package main

import (
	"os"

	"github.com/elastic/beats/libbeat/beat"

	"github.com/josephlewis42/githubbeat/beater"
)

func main() {
	err := beat.Run("githubbeat", "", beater.New)
	if err != nil {
		os.Exit(1)
	}
}
