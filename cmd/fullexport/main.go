package main

import (
	"errors"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/config"
	"openaccess-sync/pkg/data/lenel"
	"openaccess-sync/pkg/data/model"
	"os"

	"github.com/spf13/pflag"
	"github.com/xuri/excelize/v2"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			os.Exit(0)
		}
		log.Fatalf("Error parsing command line arguments: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	cl, err := client.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}
	defer func() {
		if err := cl.Close(); err != nil {
			log.Printf("Failed to close client session: %v", err)
		}
	}()

	// Load caches
	statusCache := lenel.NewBadgeStatusCache()
	if err := statusCache.Fill(cl); err != nil {
		log.Fatalf("Failed to load badge status cache: %s", err)
	}

	typeCache := lenel.NewBadgeTypeCache()
	if err := typeCache.Fill(cl); err != nil {
		log.Fatalf("Failed to load badge type cache: %s", err)
	}

	levelCache := lenel.NewAccessLevelCache()
	if err := levelCache.Fill(cl); err != nil {
		log.Fatalf("Failed to load access level cache: %s", err)
	}

	cardholderCache := lenel.NewCardholderCache()
	if err := cardholderCache.Fill(cl); err != nil {
		log.Fatalf("Failed to load cardholder level cache: %s", err)
	}

	assignmentCache := lenel.NewAssignmentCache()
	if err := assignmentCache.Fill(cl); err != nil {
		log.Fatalf("Failed to load assignment cache: %s", err)
	}

	badgeCache := lenel.NewBadgeCache()
	if err := badgeCache.Fill(cl); err != nil {
		log.Fatalf("Failed to load badge cache: %s", err)
	}
	badgeCache.Resolve(statusCache, typeCache, cardholderCache, assignmentCache, levelCache)

	f := excelize.NewFile()
	defer f.Close()

	bold, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		log.Fatalf("Failed to create bold style: %v", err)
	}

	if err := writeSheet(f, bold, "Badge Status", statusCache); err != nil {
		log.Fatalf("Failed to write badge status sheet: %v", err)
	}

	if err := writeSheet(f, bold, "Badge Type", typeCache); err != nil {
		log.Fatalf("Failed to write badge type sheet: %v", err)
	}

	if err := writeSheet(f, bold, "Access Level", levelCache); err != nil {
		log.Fatalf("Failed to write access level sheet: %v", err)
	}

	if err := writeSheet(f, bold, "Cardholder", cardholderCache); err != nil {
		log.Fatalf("Failed to write cardholder level sheet: %v", err)
	}

	if err := writeSheet(f, bold, "Badge", badgeCache); err != nil {
		log.Fatalf("Failed to write badge level sheet: %v", err)
	}

	_ = f.DeleteSheet("Sheet1")
	f.SaveAs(cfg.File)
}

// cell converts 1-indexed (col, row) to an Excel cell name like "A1".
func cell(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}

func writeHeader(f *excelize.File, sheet string, headers []string, style int) error {
	for i, h := range headers {
		c := cell(i+1, 1)

		if err := f.SetCellValue(sheet, c, h); err != nil {
			return err
		}

		if err := f.SetCellStyle(sheet, c, c, style); err != nil {
			return err
		}
	}
	return nil
}

func writeSheet(f *excelize.File, style int, sheet string, cache model.RowObjectCache) error {
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	if err := writeHeader(f, sheet, cache.RowHeader(), style); err != nil {
		return err
	}

	for i, it := range cache.GetRowItems() {
		for j, jt := range it.ToRow() {
			if err := f.SetCellValue(sheet, cell(j+1, i+2), jt); err != nil {
				return err
			}
		}
	}

	return nil
}
