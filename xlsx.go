package main

import (
	lenel2 "openaccess-sync/data/lenel"
	"openaccess-sync/data/model"
	"openaccess-sync/util/date"

	"github.com/xuri/excelize/v2"
)

// ExportXLSX writes the cache contents to an XLSX file with 5 sheets.
func ExportXLSX(cache *lenel2.DataCache, path string) error {
	f := excelize.NewFile()
	defer f.Close()

	bold, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	if err != nil {
		return err
	}

	if err := writeBadgesSheet(f, cache, bold); err != nil {
		return err
	}

	if err := writeCardholdersSheet(f, cache, bold); err != nil {
		return err
	}

	if err := writeAccessLevelsSheet(f, cache, bold); err != nil {
		return err
	}

	if err := writeBadgeTypesSheet(f, cache, bold); err != nil {
		return err
	}

	if err := writeBadgeStatusesSheet(f, cache, bold); err != nil {
		return err
	}

	_ = f.DeleteSheet("Sheet1")
	return f.SaveAs(path)
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

func writeBadgesSheet(f *excelize.File, cache *lenel2.DataCache, style int) error {
	const sheet = "badges"
	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	headers := []string{
		"ID", "Badge Key", "Activate", "Deactivate", "Status", "Type", "Cardholder SSNO",
		"Access Level 1", "Access Level 2", "Access Level 3",
		"Access Level 4", "Access Level 5", "Access Level 6",
	}

	if err := writeHeader(f, sheet, headers, style); err != nil {
		return err
	}

	for i, badge := range cache.GetBadges() {
		row := i + 2
		vals := []any{
			badge.ID,
			badge.Key,
			date.Format(badge.Activate),
			date.Format(badge.Deactivate),
			badgeStatusName(badge),
			badgeTypeName(badge),
			cardholderSSNO(badge),
		}

		levels := cache.GetAccessLevelsByBadge(badge.Key) // FIXME - changed ID to Key to enable compilation

		for j := 0; j < 6; j++ {
			if j < len(levels) {
				vals = append(vals, levels[j].Name)
			} else {
				vals = append(vals, "")
			}
		}

		for col, v := range vals {
			if err := f.SetCellValue(sheet, cell(col+1, row), v); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeCardholdersSheet(f *excelize.File, cache *lenel2.DataCache, style int) error {
	const sheet = "cardholders"

	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	if err := writeHeader(f, sheet, []string{"ID", "SSNO", "First Name", "Last Name"}, style); err != nil {
		return err
	}

	for i, ch := range cache.GetCardholders() {
		row := i + 2
		for col, v := range []any{ch.ID, ch.SSNO, ch.FirstName, ch.LastName} {
			if err := f.SetCellValue(sheet, cell(col+1, row), v); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeAccessLevelsSheet(f *excelize.File, cache *lenel2.DataCache, style int) error {
	const sheet = "access levels"

	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	if err := writeHeader(f, sheet, []string{"ID", "Name"}, style); err != nil {
		return err
	}

	for i, al := range cache.GetAccessLevels() {
		row := i + 2
		if err := f.SetCellValue(sheet, cell(1, row), al.ID); err != nil {
			return err
		}

		if err := f.SetCellValue(sheet, cell(2, row), al.Name); err != nil {
			return err
		}
	}

	return nil
}

func writeBadgeTypesSheet(f *excelize.File, cache *lenel2.DataCache, style int) error {
	const sheet = "badge types"

	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	if err := writeHeader(f, sheet, []string{"ID", "Name"}, style); err != nil {
		return err
	}

	for i, bt := range cache.GetBadgeTypes() {
		row := i + 2
		if err := f.SetCellValue(sheet, cell(1, row), bt.ID); err != nil {
			return err
		}
		if err := f.SetCellValue(sheet, cell(2, row), bt.Name); err != nil {
			return err
		}
	}

	return nil
}

func writeBadgeStatusesSheet(f *excelize.File, cache *lenel2.DataCache, style int) error {
	const sheet = "badge status"

	if _, err := f.NewSheet(sheet); err != nil {
		return err
	}

	if err := writeHeader(f, sheet, []string{"ID", "Name"}, style); err != nil {
		return err
	}

	for i, bs := range cache.GetBadgeStatuses() {
		row := i + 2
		if err := f.SetCellValue(sheet, cell(1, row), bs.ID); err != nil {
			return err
		}

		if err := f.SetCellValue(sheet, cell(2, row), bs.Name); err != nil {
			return err
		}
	}

	return nil
}

func badgeStatusName(b *model.Badge) string {
	//if b.Status != nil { // FIXME
	//	return b.Status.Name
	//}

	return ""
}

func badgeTypeName(b *model.Badge) string {
	//if b.Type != nil { // FIXME
	//	return b.Type.Name
	//}

	return ""
}

func cardholderSSNO(b *model.Badge) string {
	//if b.Cardholder != nil { // FIXME
	//	return b.Cardholder.SSNO
	//}

	return ""
}
