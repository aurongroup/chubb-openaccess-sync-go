package main

import (
	"errors"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/config"
	"openaccess-sync/pkg/data/lenel"
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

	//cache := lenel.NewDataCache(cl)

	//if err := cache.Fill(); err != nil {
	//	log.Fatalf("Failed to load API data: %v", err)
	//}

	//arc := csv.BuildAccessRecordCache(cache)
	//
	//err = csv.Write(arc.Records(), cfg.File)
	//if err != nil {
	//	log.Fatalf("Operation failed: %v", err)
	//}
}
