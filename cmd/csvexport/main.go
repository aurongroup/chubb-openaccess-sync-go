package main

import (
	"errors"
	"fmt"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/config"
	"openaccess-sync/pkg/csv"
	"openaccess-sync/pkg/data/lenel"
	"openaccess-sync/pkg/data/model"
	"os"

	"github.com/spf13/pflag"
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

	// 1. Retrieve status, type, and access levels from Lenel
	bsc := lenel.NewBadgeStatusCache()
	if err := bsc.Fill(cl); err != nil {
		log.Fatalf("Badge status cache fill failed: %v", err)
	}

	btc := lenel.NewBadgeTypeCache()
	if err := btc.Fill(cl); err != nil {
		log.Fatalf("Badge type cache fill failed: %v", err)
	}

	alc := lenel.NewAccessLevelCache()
	if err := alc.Fill(cl); err != nil {
		log.Fatalf("Access level cache fill failed: %v", err)
	}

	bc := lenel.NewBadgeCache()
	if err := bc.Fill(cl); err != nil {
		log.Fatalf("Badge cache fill failed: %v", err)
	}

	records := make([]*model.AccessRecord, 0, len(bc.GetItems()))
	for _, badge := range bc.GetItems() {
		r, err := model.NewAccessRecord(
			"ssno",
			"first",
			"last",
			"accLvl1",
			"accLvl2",
			"accLvl3",
			"accLvl4",
			"accLvl5",
			"accLvl6",
			fmt.Sprintf("%d", badge.ID),
			nil,        // activate,
			nil,        //deactivate *time.Time,
			"active",   // status,
			"employee", //badgeType string,
		)

		if err != nil {
			log.Printf("Failed to create access record for badge ID %d: %v", badge.ID, err)
			continue
		}

		records = append(records, r)
	}

	err = csv.Write(records, cfg.File)
	if err != nil {
		log.Fatalf("Operation failed: %v", err)
	}
}
