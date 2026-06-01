package main

import (
	"errors"
	"fmt"
	"log"
	"openaccess-sync/pkg/client"
	"openaccess-sync/pkg/config"
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

	// 1. Identify and remove inactive badges
	// 2. Identify and remove cardholders without badges
	// 3. Retrieve all badges and identify cardholders with more than one badge
	//    -> For each additional badge
	//    ----> 3.1 Retrieve access levels for badge
	//    ----> 3.2 Delete existing badge
	//    ----> 3.3 Create new cardholder record
	//    ----> 3.4 Create new badge linked to new cardholder
	//    ----> 3.5 Add access levels to new badge
	// 4. Update cardholders with existing SSNO, moving the SSNO to "Office Phone" (OPHONE)
	// 5. Update all cardholder records with new SSNO based on attached badges

	// 1. Identify and remove inactive badges
	activeStatus := bsc.GetByName("Active")
	if activeStatus == nil {
		log.Fatalf("Badge status 'Active' not found in cache")
	}

	badgeCache := lenel.NewBadgeCache()
	if err := badgeCache.FillWithFilter(cl, fmt.Sprintf("STATUS!=%d", activeStatus.ID)); err != nil {
		log.Fatalf("Badge cache fill failed for non-active badges: %v", err)
	}

	assignmentCache := lenel.NewAssignmentCache()
	if err := assignmentCache.Fill(cl); err != nil {
		log.Fatalf("Assignment cache fill failed: %v", err)
	}

	for _, b := range badgeCache.GetItems() {
		s := bsc.GetByID(b.Status)
		if s == nil {
			log.Printf("Badge status ID %d not found in cache", b.Status)
			continue
		}
		log.Printf("Deleting badge %d (%s)", b.Key, s.Name)
		if err := badgeCache.Delete(cl, b); err != nil {
			log.Printf("Failed to delete badge %d: %v", b.Key, err)
		}
	}

	// 2. Identify and remove cardholders without badges
	cardholderCache := lenel.NewCardholderCache()
	if err := cardholderCache.FillDetached(cl); err != nil {
		log.Fatalf("Detached cardholder cache fill failed: %v", err)
	}

	for _, c := range cardholderCache.GetItems() {
		log.Printf("Deleting detached cardholder %s %s (%d)", c.FirstName, c.LastName, c.ID)

		if err := cardholderCache.Delete(cl, c); err != nil {
			log.Printf("Failed to delete cardholder %s %s (%d) %v", c.FirstName, c.LastName, c.ID, err)
		}
	}

	// 3. Find cardholders with an SSNO and move the value to OPHONE (office phone)
	cardholderCache = lenel.NewCardholderCache()
	if err := cardholderCache.Fill(cl); err != nil {
		log.Fatalf("Cardholder cache fill failed: %v", err)
	}

	for _, c := range cardholderCache.GetItems() {
		if c.SSNO != "" {
			log.Printf("Moving SSNO to OPHONE for cardholder %s %s (%d) [SSNO: %s]", c.FirstName, c.LastName, c.ID, c.SSNO)
			c.OfficePhone = c.SSNO
			c.SSNO = ""
			if err := cardholderCache.Update(cl, c); err != nil {
				log.Printf("Failed to update cardholder %s %s (%d) %v", c.FirstName, c.LastName, c.ID, err)
			}
		}
	}

	// 4. Retrieve all badges and identify cardholders with more than one badge
	cardholderCache = lenel.NewCardholderCache()
	if err := cardholderCache.Fill(cl); err != nil {
		log.Fatalf("Cardholder cache fill failed: %v", err)
	}

	badgeCache = lenel.NewBadgeCache()
	if err := badgeCache.Fill(cl); err != nil {
		log.Fatalf("Badge cache fill failed: %v", err)
	}

	cardholderMap := make(map[int32]*model.Cardholder)
	duplicates := make(map[int32][]int32)

	for _, b := range badgeCache.GetItems() {
		c := cardholderCache.GetByID(b.Cardholder)

		if _, ok := cardholderMap[b.Cardholder]; ok {
			if _, exists := duplicates[b.Cardholder]; !exists {
				duplicates[b.Cardholder] = []int32{}
			}
			duplicates[b.Cardholder] = append(duplicates[b.Cardholder], b.Key)
			continue
		} else {
			cardholderMap[b.Cardholder] = c
		}
	}

	if len(duplicates) > 0 {
		log.Printf("Found %d duplicate cardholder IDs: %v", len(duplicates), duplicates) // FIXME

		for cardholderID, badgeIDs := range duplicates {
			log.Printf("Duplicate cardholder ID %d found on badges: %v", cardholderID, badgeIDs)

			for _, badgeID := range badgeIDs {
				log.Printf("Badge ID %d has duplicate cardholder ID %d", badgeID, cardholderID)

				assignments := assignmentCache.GetItemsByBadgeKey(badgeID)

				for _, assignment := range assignments {
					a := alc.GetByID(assignment.AccessLevel)
					if a == nil {
						log.Printf("Access level %d not found for badge ID %d", assignment.AccessLevel, badgeID)
						continue
					}
					log.Printf("Assignment for badge ID %d: %s (%d)", badgeID, a.Name, a.ID)
				}

				// Delete badge

				// Create new badge

				// Add access levels

				//ch := model.NewCardholder(0, badgeID, ch.Firstname, ch.Lastname)
				//cardholderCache.Create(cl, ch)
				//
				//b.Cardholder = ch.ID
				//badgeCache.Update(cl, b)
			}
		}
	}

	// 4. Update cardholders with existing SSNO, moving the SSNO to "Office Phone" (OPHONE)

	//for _, b := range badgeCache.GetItems() {
	//	ch := cardholderCache.GetByID(b.Cardholder)
	//	if ch == nil {
	//		log.Printf("Cardholder ID %d not found in cache", b.Cardholder)
	//		continue
	//	}
	//
	//	if ch.SSNO != "" && ch.OfficePhone == "" {
	//		ch.OfficePhone = ch.SSNO
	//	}
	//
	//	ch.SSNO = strconv.FormatInt(b.ID, 10)
	//
	//	log.Printf("Setting cardholder %s %s (%d) SSNO to badge ID %d", ch.FirstName, ch.LastName, ch.ID, b.ID)
	//	if err := cardholderCache.Update(cl, ch); err != nil {
	//		log.Printf("Failed to update cardholder %d: %v", b.Key, err)
	//	}
	//}

	//for _, c := range cardholderCache.GetItems() {
	//	log.Printf("Updating cardholder %s %s (%d)", c.FirstName, c.LastName, c.ID)
	//
	//	if err := cardholderCache.Update(cl, c); err != nil {
	//		log.Printf("Failed to delete cardholder %s %s (%d) %v", c.FirstName, c.LastName, c.ID, err)
	//	}
	//}
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
