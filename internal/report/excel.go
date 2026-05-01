package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/facundo/splitwise-reconcilier/internal/mapping"
	"github.com/xuri/excelize/v2"
)

func Generate(expenses []mapping.MappedExpense, cfg *mapping.Config, month time.Time, outputPath, myName, partnerName string) error {
	f := excelize.NewFile()
	defer f.Close()

	boldStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E1F2"}, Pattern: 1},
	})
	if err != nil {
		return fmt.Errorf("creating bold style: %w", err)
	}

	currencyStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 7, // #,##0.00
	})
	if err != nil {
		return fmt.Errorf("creating currency style: %w", err)
	}

	if err := buildDetailSheet(f, expenses, cfg, boldStyle, currencyStyle, myName, partnerName); err != nil {
		return err
	}
	if err := buildSummarySheet(f, expenses, cfg, boldStyle, currencyStyle, myName); err != nil {
		return err
	}
	if err := buildUnassignedSheet(f, expenses, cfg, boldStyle, currencyStyle, myName); err != nil {
		return err
	}

	// Remove default "Sheet1" if it still exists
	f.DeleteSheet("Sheet1")

	return f.SaveAs(outputPath)
}

func buildDetailSheet(f *excelize.File, expenses []mapping.MappedExpense, cfg *mapping.Config, boldStyle, currencyStyle int, myName, partnerName string) error {
	sheet := "Expense Detail"
	f.NewSheet(sheet)

	headers := []string{"Date", "Description", "Category", "Total Amount", myName, partnerName, "Card", "Expense ID", "Notes"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, boldStyle)
	}

	for i, e := range expenses {
		row := i + 2
		date, _ := time.Parse(time.RFC3339, e.Expense.Date)

		f.SetCellValue(sheet, cellName(1, row), date.Format("2006-01-02"))
		f.SetCellValue(sheet, cellName(2, row), e.Expense.Description)
		f.SetCellValue(sheet, cellName(3, row), e.Expense.Category.Name)
		f.SetCellValue(sheet, cellName(4, row), mapping.ParseCost(e.Expense.Cost))
		f.SetCellValue(sheet, cellName(5, row), e.MyShare)
		f.SetCellValue(sheet, cellName(6, row), e.TheirShare)
		f.SetCellValue(sheet, cellName(7, row), mapping.FormatCardName(cfg, e.Card))
		f.SetCellValue(sheet, cellName(8, row), e.Expense.ID)
		f.SetCellValue(sheet, cellName(9, row), e.Expense.Details)

		for _, col := range []int{4, 5, 6} {
			f.SetCellStyle(sheet, cellName(col, row), cellName(col, row), currencyStyle)
		}
	}

	widths := map[int]float64{1: 12, 2: 35, 3: 18, 4: 14, 5: 12, 6: 14, 7: 20, 8: 12, 9: 30}
	for col, w := range widths {
		colName, _ := excelize.ColumnNumberToName(col)
		f.SetColWidth(sheet, colName, colName, w)
	}

	return nil
}

func buildSummarySheet(f *excelize.File, expenses []mapping.MappedExpense, cfg *mapping.Config, boldStyle, currencyStyle int, myName string) error {
	sheet := "Summary by Card"
	f.NewSheet(sheet)

	totals := map[string]float64{}
	for _, e := range expenses {
		totals[e.Card] += e.MyShare
	}

	headers := []string{"Card", myName}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, boldStyle)
	}

	row := 2
	for _, card := range cardOrder(cfg, totals) {
		name := mapping.FormatCardName(cfg, card)
		f.SetCellValue(sheet, cellName(1, row), name)
		f.SetCellValue(sheet, cellName(2, row), totals[card])
		f.SetCellStyle(sheet, cellName(2, row), cellName(2, row), currencyStyle)
		row++
	}

	f.SetColWidth(sheet, "A", "A", 25)
	f.SetColWidth(sheet, "B", "B", 16)
	return nil
}

func buildUnassignedSheet(f *excelize.File, expenses []mapping.MappedExpense, cfg *mapping.Config, boldStyle, currencyStyle int, myName string) error {
	sheet := "Unassigned"
	f.NewSheet(sheet)

	headers := []string{"Date", "Description", "Category", "Total Amount", myName, "Expense ID", "Notes"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, boldStyle)
	}

	row := 2
	for _, e := range expenses {
		if e.Card != mapping.Unassigned {
			continue
		}
		date, _ := time.Parse(time.RFC3339, e.Expense.Date)
		f.SetCellValue(sheet, cellName(1, row), date.Format("2006-01-02"))
		f.SetCellValue(sheet, cellName(2, row), e.Expense.Description)
		f.SetCellValue(sheet, cellName(3, row), e.Expense.Category.Name)
		f.SetCellValue(sheet, cellName(4, row), mapping.ParseCost(e.Expense.Cost))
		f.SetCellValue(sheet, cellName(5, row), e.MyShare)
		f.SetCellValue(sheet, cellName(6, row), e.Expense.ID)
		f.SetCellValue(sheet, cellName(7, row), e.Expense.Details)

		for _, col := range []int{4, 5} {
			f.SetCellStyle(sheet, cellName(col, row), cellName(col, row), currencyStyle)
		}
		row++
	}

	widths := map[int]float64{1: 12, 2: 35, 3: 18, 4: 14, 5: 12, 6: 12, 7: 30}
	for col, w := range widths {
		colName, _ := excelize.ColumnNumberToName(col)
		f.SetColWidth(sheet, colName, colName, w)
	}

	return nil
}

// cardOrder returns card IDs ordered: configured cards first, then UNASSIGNED.
func cardOrder(cfg *mapping.Config, totals map[string]float64) []string {
	seen := map[string]bool{}
	var order []string
	for _, c := range cfg.Cards {
		if _, ok := totals[c.ID]; ok {
			order = append(order, c.ID)
			seen[c.ID] = true
		}
	}
	for k := range totals {
		if !seen[k] && !strings.EqualFold(k, mapping.Unassigned) {
			order = append(order, k)
		}
	}
	if _, ok := totals[mapping.Unassigned]; ok {
		order = append(order, mapping.Unassigned)
	}
	return order
}

func cellName(col, row int) string {
	name, _ := excelize.CoordinatesToCellName(col, row)
	return name
}
