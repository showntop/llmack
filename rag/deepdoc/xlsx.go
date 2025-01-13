package deepdoc

import (
	"bytes"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// XlsxDocument ...
type XlsxDocument struct {
}

// Excel ...
func Excel() *XlsxDocument {
	return &XlsxDocument{}
}

// Extract ...
func (d *XlsxDocument) Extract(filename string, binary []byte) ([]string, error) {
	doc, err := excelize.OpenReader(bytes.NewReader(binary))
	if err != nil {
		return nil, err
	}
	if len(doc.GetSheetList()) <= 0 {
		return nil, fmt.Errorf("empty excel")
	}
	sheet := doc.GetSheetList()[0]

	sections := []string{}
	rows, err := doc.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	for i, row := range rows {
		for j := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			value, _ := doc.GetCellValue(sheet, cell)

			_, link, _ := doc.GetCellHyperLink(sheet, cell)
			value += link
			sections = append(sections, value)
		}
	}

	return sections, nil
}
