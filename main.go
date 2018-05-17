package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > reading config > exit >", err)
		os.Exit(1)
	}
	db, err := createDatabaseConnection(config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > connection to database > exit >", err)
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stdout, "INFO > connection to database > in process >", err)
	}

	router := &router{
		db: db,
	}

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > listening server > exit >", err)
		os.Exit(1)
	}
}
