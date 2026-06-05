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
	"strconv"

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
	// 3. Update cardholders with existing SSNO, moving the SSNO to "Office Phone" (OPHONE)
	// 4. Retrieve all badges and identify cardholders with more than one badge
	//    -> For each additional badge
	//    ----> 4.1 Retrieve access levels for badge
	//    ----> 4.2 Delete existing badge
	//    ----> 4.3 Create new cardholder record
	//    ----> 4.4 Create new badge linked to new cardholder
	//    ----> 4.5 Add access levels to new badge
	// 5. Update all cardholder records with new SSNO based on attached badges

	// 1. Identify and remove inactive badges
	log.Printf("# 1/5 Identify and remove inactive badges")
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

	badgeCache = lenel.NewBadgeCache()
	if err := badgeCache.Fill(cl); err != nil {
		log.Fatalf("Badge cache fill failed: %v", err)
	}

	// 2. Identify and remove cardholders without badges
	log.Printf("# 2/5 Identify and remove cardholders without badges")
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

	// 3. Find cardholders with an SSNO and move the value to OPHONE (office phone),
	// unless the SSNO matches one of the cardholder's attached badge IDs.
	// --- Perhaps move remove duplicates first?
	log.Printf("# 3/5 Find cardholders with an SSNO and move to OPHONE")
	cardholderCache = lenel.NewCardholderCache()
	if err := cardholderCache.Fill(cl); err != nil {
		log.Fatalf("Cardholder cache fill failed: %v", err)
	}

	badgeCache = lenel.NewBadgeCache()
	if err := badgeCache.Fill(cl); err != nil {
		log.Fatalf("Badge cache fill failed: %v", err)
	}

	for _, c := range cardholderCache.GetItems() {
		if c.SSNO != "" {
			ssnoID, err := strconv.ParseInt(c.SSNO, 10, 64)
			isBadgeID := err == nil && func() bool {
				for _, b := range badgeCache.GetByCardholder(c.ID) {
					log.Printf("Checking badge ID %d for cardholder %s %s (%d)", b.ID, c.FirstName, c.LastName, c.ID)
					if b.ID == ssnoID {
						log.Printf("SSNO %s matches attached badge ID %d for cardholder %s %s (%d)", c.SSNO, b.ID, c.FirstName, c.LastName, c.ID)
						return true
					}
				}
				return false
			}()

			if isBadgeID {
				log.Printf("Skipping SSNO→OPHONE for cardholder %s %s (%d): SSNO %s matches attached badge", c.FirstName, c.LastName, c.ID, c.SSNO)
				continue
			}

			log.Printf("Moving SSNO to OPHONE for cardholder %s %s (%d) [SSNO: %s]", c.FirstName, c.LastName, c.ID, c.SSNO)
			c.OfficePhone = c.SSNO
			c.SSNO = ""
			if err := cardholderCache.Update(cl, c); err != nil {
				log.Printf("Failed to update cardholder %s %s (%d) %v", c.FirstName, c.LastName, c.ID, err)
			}
		}
	}

	// 4. Retrieve all badges and identify cardholders with more than one badge
	log.Printf("# 4/5 Identify cardholders with more than one badge")
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
		log.Printf("Found %d duplicate cardholder IDs", len(duplicates))

		for cardholderID, badgeKeys := range duplicates {
			log.Printf("--> Duplicate cardholder ID %d found on badges: %v", cardholderID, badgeKeys)

			for _, badgeKey := range badgeKeys {
				log.Printf("----> Badge ID %d has duplicate cardholder ID %d", badgeKey, cardholderID)

				existingBadge := badgeCache.GetByKey(badgeKey)
				if existingBadge == nil {
					log.Printf("***** Badge ID %d not found in cache - skipping", badgeKey)
					continue
				}

				assignments := assignmentCache.GetItemsByBadgeKey(badgeKey)

				newBadge, err := model.NewBadge(
					existingBadge.ID,
					0,
					existingBadge.Activate,
					existingBadge.Deactivate,
					existingBadge.Status,
					existingBadge.Type,
					cardholderID,
				)

				if err != nil {
					log.Printf("***** Error creating new badge for badge ID %d: %v", badgeKey, err)
					continue
				}

				_, err = badgeCache.Create(cl, newBadge)
				if err != nil {
					log.Printf("***** Error creating new badge for badge ID %d: %v", badgeKey, err)
					continue
				}

				for _, assignment := range assignments {
					// TODO - logging
					a := alc.GetByID(assignment.AccessLevel)
					if a == nil {
						log.Printf("Access level %d not found for badge ID %d", assignment.AccessLevel, badgeKey)
						continue
					}
					log.Printf("------> Assignment for badge ID %d: %s (%d)", badgeKey, a.Name, a.ID)

					newAssignment, err := model.NewAccessLevelAssignment(assignment.AccessLevel, newBadge.Key)
					if err != nil {
						log.Printf("******* Error creating new assignment object for badge ID %d: %v", badgeKey, err)
						continue
					}

					err = assignmentCache.Create(cl, newAssignment)
					if err != nil {
						log.Printf("******* Error creating new assignment for badge ID %d: %v", badgeKey, err)
						continue
					}

					// Delete existing assignment
					err = assignmentCache.Delete(cl, assignment)
					if err != nil {
						log.Printf("******* Error deleting assignment for existing badge ID %d: %v", badgeKey, err)
						continue
					}
				}

				// Delete badge
				err = badgeCache.Delete(cl, existingBadge)
				if err != nil {
					log.Printf("***** Error deleting badge ID %d: %v", badgeKey, err)
					continue
				}
			}
		}
	}

	//// 5. Update all cardholder records with new SSNO based on attached badges
	log.Printf("# 5/5 Update all cardholder records with new SSNO based on attached badges")
	badgeCache = lenel.NewBadgeCache()
	if err := badgeCache.Fill(cl); err != nil {
		log.Fatalf("Badge cache fill failed: %v", err)
	}

	for _, cardholder := range cardholderCache.GetItems() {
		badges := badgeCache.GetByCardholder(cardholder.ID)
		if len(badges) > 0 {
			if len(badges) > 1 {
				log.Printf("Cardholder %s %s (%d) has multiple badges: %d", cardholder.FirstName, cardholder.LastName, cardholder.ID, len(badges))
				continue
			}

			badgeIDStr := strconv.FormatInt(badges[0].ID, 10)
			if cardholder.SSNO == badgeIDStr {
				log.Printf("Skipping cardholder %s %s (%d): SSNO %s already matches attached badge", cardholder.FirstName, cardholder.LastName, cardholder.ID, cardholder.SSNO)
				continue
			}

			cardholder.SSNO = badgeIDStr
			err := cardholderCache.Update(cl, cardholder)
			if err != nil {
				log.Printf("Failed to update cardholder %d: %v", cardholder.ID, err)
				continue
			}

			log.Printf("Updated cardholder %s %s (%d) SSNO to badge ID %d", cardholder.FirstName, cardholder.LastName, cardholder.ID, badges[0].ID)
		}
	}
}
