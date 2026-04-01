// Stockyard Cipher — Secrets manager for CI/CD.
// Store env vars encrypted, expose to pipelines via short-lived tokens.
// Single binary, embedded SQLite, AES-256-GCM encryption at rest.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/stockyard-dev/stockyard-cipher/internal/license"
	"github.com/stockyard-dev/stockyard-cipher/internal/server"
	"github.com/stockyard-dev/stockyard-cipher/internal/store"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Printf("cipher %s\n", version)
		os.Exit(0)
	}
	if len(os.Args) > 1 && (os.Args[1] == "--health" || os.Args[1] == "health") {
		fmt.Println("ok")
		os.Exit(0)
	}

	log.SetFlags(log.Ltime | log.Lshortfile)

	retentionDays := 30
	if r := os.Getenv("RETENTION_DAYS"); r != "" {
		if n, err := strconv.Atoi(r); err == nil && n > 0 {
			retentionDays = n
		}
	}

	port := 8870
	if p := os.Getenv("PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	encKey := os.Getenv("CIPHER_ENCRYPTION_KEY")

	licenseKey := os.Getenv("CIPHER_LICENSE_KEY")
	licInfo, licErr := license.Validate(licenseKey, "cipher")
	if licenseKey != "" && licErr != nil {
		log.Printf("[license] WARNING: %v — running in free tier", licErr)
		licInfo = nil
	}
	limits := server.LimitsFor(licInfo)
	if licInfo != nil && licInfo.IsPro() {
		log.Printf("  License:   Pro (%s)", licInfo.CustomerID)
	} else {
		log.Printf("  License:   Free tier (set CIPHER_LICENSE_KEY to unlock Pro)")
	}

	if limits.RetentionDays > retentionDays {
		retentionDays = limits.RetentionDays
	}

	db, err := store.Open(dataDir, encKey)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	log.Printf("")
	log.Printf("  Stockyard Cipher %s", version)
	log.Printf("  API:            http://localhost:%d/api/projects", port)
	log.Printf("  Secrets:        http://localhost:%d/api/secrets (token auth)", port)
	log.Printf("  Encryption:     AES-256-GCM")
	log.Printf("  Retention:      %d days", retentionDays)
	log.Printf("  Dashboard:      http://localhost:%d/ui", port)
	log.Printf("")

	go func() {
		for {
			time.Sleep(6 * time.Hour)
			n, err := db.Cleanup(retentionDays)
			if err != nil {
				log.Printf("[cleanup] error: %v", err)
			} else if n > 0 {
				log.Printf("[cleanup] deleted %d old log entries, expired tokens", n)
			}
		}
	}()

	srv := server.New(db, port, limits)
	if err := srv.Start(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
