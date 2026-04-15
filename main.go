package main

import (
	"errors"
	"io"
	"log"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("Error parsing command line arguments: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	client, err := NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close client session: %v", err)
		}
	}()

	cache := NewDataCache(client)
	if err := cache.Fill(); err != nil {
		log.Fatalf("Failed to load API data: %v", err)
	}

	switch {
	case cfg.ExportFile != "":
		err = PrintCSVReport(cache.records, cfg.ExportFile)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}
		break

	case cfg.InputFile != "":
		csvValues, err := ParseCSV(cfg.InputFile)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}

		csvRecords := make([]*AccessRecord, len(csvValues))
		for i := range csvValues {
			csvRecords[i] = &csvValues[i]
		}

		var diffWriter io.Writer
		if cfg.DiffFile != "" {
			f, err := os.Create(cfg.DiffFile)
			if err != nil {
				log.Fatalf("Failed to open diff file: %v", err)
			}
			defer f.Close()
			diffWriter = f
		}
		result := CompareRecords(csvRecords, cache.records, diffWriter)
		for _, r := range result {
			log.Printf("status=%s ssno=%s badgeId=%s", r.SyncStatus.String(), r.SSNO, r.BadgeID)
		}
		break

	case cfg.Cleanup:
		log.Println("cleanup not yet implemented")
		break

	case cfg.FullExportFile != "":
		err = ExportXLSX(cache, cfg.FullExportFile)
		if err != nil {
			log.Fatalf("Operation failed: %v", err)
		}
		break
	}
}
