//go:build ignore

package main

import (
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

func main() {
	migrator, err := rivermigrate.New[pgx.Tx](riverpgxv5.New(nil), nil)
	if err != nil {
		log.Fatal(err)
	}
	migrations := migrator.AllVersions()

	fmt.Println("-- +goose Up")
	for _, m := range migrations {
		fmt.Printf("-- Migration version %d\n", m.Version)
		fmt.Println(m.SQLUp)
		fmt.Println()
	}

	fmt.Println("-- +goose Down")
	for i := len(migrations) - 1; i >= 0; i-- {
		m := migrations[i]
		fmt.Printf("-- Migration version %d\n", m.Version)
		fmt.Println(m.SQLDown)
		fmt.Println()
	}
}
