package utils

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

func ImportExcelSheet(filePath string, sheetName string) ([][]string, error) {
	// Step 1: Open the Excel file
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	// Step 2: Validate sheet exists
	sheetList := file.GetSheetList()
	found := false
	for _, s := range sheetList {
		if s == sheetName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("sheet %q not found in file %q", sheetName, filePath)
	}

	// Step 3: Get all rows from the selected sheet
	rows, err := file.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("error reading rows from sheet %q: %w", sheetName, err)
	}

	// Step 4: Return rows (first row is header)
	return rows, nil
}
