package main

import (
	"errors"
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

	if err := dispatch(cfg, cache); err != nil {
		log.Fatalf("Operation failed: %v", err)
	}
}

func dispatch(cfg AppConfig, cache *DataCache) error {
	switch {
	case cfg.ExportFile != "":
		return PrintCSVReport(cache.records, cfg.ExportFile)

	case cfg.InputFile != "":
		csvValues, err := ParseCSV(cfg.InputFile)
		if err != nil {
			return err
		}
		csvRecords := make([]*AccessRecord, len(csvValues))
		for i := range csvValues {
			csvRecords[i] = &csvValues[i]
		}
		result := CompareRecords(csvRecords, cache.records)
		for _, r := range result {
			log.Printf("sync status=%s ssno=%s badgeId=%s", r.SyncStatus.String(), r.SSNO, r.BadgeID)
		}

	case cfg.Cleanup:
		log.Println("cleanup not yet implemented")

	case cfg.FullExportFile != "":
		return ExportXLSX(cache, cfg.FullExportFile)
	}
	return nil
}
