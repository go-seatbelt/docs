package main

import (
	"log"

	"github.com/go-seatbelt/seatbelt"
)

func logger(fn func(c *seatbelt.Context) error) func(*seatbelt.Context) error {
	return func(c *seatbelt.Context) error {
		log.Printf("received request: %s\n", c.Request().URL.Path)
		return fn(c)
	}
}

func main() {
	app := seatbelt.New()
	app.Use(logger)
}
