package main

import (
	"fmt"
	"log"
	"os"

	"project-3/db"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run cmd/migrate/main.go <command>\nAvailable commands: up, down, status, reset")
	}
	cmd := os.Args[1]

	opts := db.DefaultMigrationOptions()

	var err error
	switch cmd {
	case "up":
		err = db.MigrateUp(opts)
	case "down":
		err = db.MigrateDown(opts)
	case "status":
		err = db.MigrateStatus(opts)
	case "reset":
		err = db.MigrateReset(opts)
	default:
		log.Fatalf("Unknown command: %s\nAvailable: up, down, status, reset", cmd)
	}

	if err != nil {
		log.Fatalf("command %s failed: %v", cmd, err)
	}
	fmt.Println("Migration command completed:", cmd)
}
