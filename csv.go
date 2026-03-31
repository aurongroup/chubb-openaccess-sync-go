package main

import (
	"encoding/csv"
	"io"
	"log/slog"
	"os"
	"strings"
)

// ParseCSV reads a pipe-delimited access record CSV from path.
func ParseCSV(path string) ([]AccessRecord, error) {
	slog.Info("Parsing access records from file", "path", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseCSVReader(f)
}

func parseCSVReader(r io.Reader) ([]AccessRecord, error) {
	cr := csv.NewReader(r)
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

	var records []AccessRecord
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, mapRowToAccessRecord(row, col))
	}

	slog.Info("Parsed access records", "count", len(records))
	return records, nil
}

func mapRowToAccessRecord(row []string, col map[string]int) AccessRecord {
	get := func(name string) string {
		i, ok := col[name]
		if !ok || i >= len(row) {
			return ""
		}
		return row[i]
	}
	return AccessRecord{
		SSNO:       get("ssno"),
		First:      get("first"),
		Last:       get("last"),
		AccLvl1:    get("acc_lvl1"),
		AccLvl2:    get("acc_lvl2"),
		AccLvl3:    get("acc_lvl3"),
		AccLvl4:    get("acc_lvl4"),
		AccLvl5:    get("acc_lvl5"),
		AccLvl6:    get("acc_lvl6"),
		BadgeID:    get("badgeid"),
		Activate:   parseDate(get("activate")),
		Deactivate: parseDate(get("deactivate")),
		Status:     get("status"),
		BadgeType:  get("badge type"),
	}
}

// PrintCSVReport writes the access records to a pipe-delimited CSV file.
func PrintCSVReport(records []AccessRecord, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
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
