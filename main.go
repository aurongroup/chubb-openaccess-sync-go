package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		slog.Error("Error parsing command line arguments", "err", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		slog.Error("Invalid configuration", "err", err)
		os.Exit(1)
	}

	client, err := NewClient(cfg)
	if err != nil {
		slog.Error("Failed to initialize client", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := client.Close(); err != nil {
			slog.Warn("Failed to close client session", "err", err)
		}
	}()

	cache := NewDataCache(client)
	if err := cache.Fill(); err != nil {
		slog.Error("Failed to load API data", "err", err)
		os.Exit(1)
	}

	if err := dispatch(cfg, cache); err != nil {
		slog.Error("Operation failed", "err", err)
		os.Exit(1)
	}
}

func dispatch(cfg AppConfig, cache *DataCache) error {
	switch {
	case cfg.ExportFile != "":
		return PrintCSVReport(cache.records, cfg.ExportFile)

	case cfg.InputFile != "":
		csvRecords, err := ParseCSV(cfg.InputFile)
		if err != nil {
			return err
		}
		result := CompareRecords(csvRecords, cache.records)
		for _, r := range result {
			slog.Info("sync",
				"status", r.SyncStatus.String(),
				"ssno", r.SSNO,
				"badgeId", r.BadgeID,
			)
		}

	case cfg.Cleanup:
		slog.Info("cleanup not yet implemented")

	case cfg.FullExportFile != "":
		return ExportXLSX(cache, cfg.FullExportFile)
	}
	return nil
}
