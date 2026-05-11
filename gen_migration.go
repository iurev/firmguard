package main

import (
	"context"
	"fmt"
	"log"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

func main() {
	migrator, err := rivermigrate.New[riverpgxv5.Driver](riverpgxv5.New(nil), nil)
	if err != nil {
		log.Fatal(err)
	}
	migrations, err := migrator.GetAvailableMigrations(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("-- +goose Up")
	for _, m := range migrations {
		fmt.Printf("-- Migration version %d\n", m.Version)
		fmt.Println(m.StatementsUp)
		fmt.Println()
	}

	fmt.Println("-- +goose Down")
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		fmt.Printf("-- Migration version %d\n", m.Version)
		fmt.Println(m.StatementsDown)
		fmt.Println()
	}
}
