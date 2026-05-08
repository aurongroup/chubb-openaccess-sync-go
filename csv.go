package main

import (
	ecsv "encoding/csv"
	"io"
	"log"
	"openaccess-sync/data/model/csv"
	"openaccess-sync/util/date"
	"os"
	"strings"
)

// ParseCSV reads a pipe-delimited access record CSV from path.
func ParseCSV(path string) ([]*csv.AccessRecord, error) {
	log.Printf("Parsing access records from file: %s", path)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cr := ecsv.NewReader(f)
	cr.Comma = '|'
	cr.LazyQuotes = true

	// Read and index header row
	header, err := cr.Read()
	if err != nil {
		return nil, err
	}

	col := make(map[string]int, len(header))
	for i, h := range header {
		col[strings.TrimSpace(h)] = i
	}

	var records []*csv.AccessRecord
	for {
		row, err := cr.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		r, err := mapRowToAccessRecord(row, col)
		if err != nil {
			return nil, err
		}

		records = append(records, r)
	}

	log.Printf("Total CSV access records: %d", len(records))
	return records, nil
}

func mapRowToAccessRecord(row []string, col map[string]int) (*csv.AccessRecord, error) {
	get := func(name string) string {
		i, ok := col[name]

		if !ok || i >= len(row) {
			return ""
		}

		return row[i]
	}

	return csv.NewAccessRecord(
		get("ssno"),
		get("first"),
		get("last"),
		get("acc_lvl1"),
		get("acc_lvl2"),
		get("acc_lvl3"),
		get("acc_lvl4"),
		get("acc_lvl5"),
		get("acc_lvl6"),
		get("badgeid"),
		date.Parse(get("activate")),
		date.Parse(get("deactivate")),
		get("status"),
		get("badge type"),
	)
}

// PrintCSVReport writes the access records to a pipe-delimited CSV file.
func PrintCSVReport(records []*csv.AccessRecord, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := ecsv.NewWriter(f)
	w.Comma = '|'

	header := []string{
		"ssno", "first", "last",
		"acc_lvl1", "acc_lvl2", "acc_lvl3", "acc_lvl4", "acc_lvl5", "acc_lvl6",
		"badgeid", "activate", "deactivate", "status", "badge type",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range records {
		if err := w.Write(r.ToRow()); err != nil {
			return err
		}
	}

	w.Flush()

	return w.Error()
}
