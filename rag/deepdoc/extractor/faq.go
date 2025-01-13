package extractor

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"strings"

	"github.com/xuri/excelize/v2"
)

// FaqExtractor ...
type FaqExtractor struct {
}

// NewFaqExtractor ...
func NewFaqExtractor() *FaqExtractor {
	return &FaqExtractor{}
}

// Extract ...
func (e *FaqExtractor) Extract(m *Meta, binary []byte) ([]string, error) {
	if len(binary) == 0 {
		// get again
	}

	if reCsv.MatchString(m.Filename) {
		return Csv().Extract(m.Filename, binary)
	}
	if reXlsx.MatchString(m.Filename) {
		return Excel().Extract(m.Filename, binary)
	}

	return nil, fmt.Errorf("invalid file format %v for faq", m.Filename)
}

// XlsxDocument ...
type XlsxDocument struct {
}

// Excel ...
func Excel() *XlsxDocument {
	return &XlsxDocument{}
}

// Extract extract QA pairs from excel
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
	if len(rows) <= 1 {
		return nil, fmt.Errorf("empty sheet")
	}

	m, n := 0, 0
	for i := range rows[0] {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1) // Q
		value, _ := doc.GetCellValue(sheet, cell)
		if value == "问题" {
			m = i + 1
		}
		if value == "答案" {
			n = i + 1
		}
	}
	if m == 0 || n == 0 {
		// return nil, fmt.Errorf("invalid faq file")
		for i, row := range rows {
			for j := range row {
				cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
				value, _ := doc.GetCellValue(sheet, cell)
				value = strings.ReplaceAll(value, "\n", "")

				_, link, _ := doc.GetCellHyperLink(sheet, cell)
				value += link
				sections = append(sections, value)
			}
		}
	} else {
		for i := 1; i < len(rows); i++ {
			cellQ, _ := excelize.CoordinatesToCellName(m, i+1) // Q
			valueQ, _ := doc.GetCellValue(sheet, cellQ)

			cellA, _ := excelize.CoordinatesToCellName(n, i+1) // A
			_, link, _ := doc.GetCellHyperLink(sheet, cellA)
			valueA, _ := doc.GetCellValue(sheet, cellA)
			valueA += link

			sections = append(sections, valueQ+"|||"+valueA)
		}
	}

	return sections, nil
}

// CsvDocument ...
type CsvDocument struct {
}

// Csv ...
func Csv() *CsvDocument {
	return &CsvDocument{}
}

// Extract ...
func (d *CsvDocument) Extract(filename string, binary []byte) ([]string, error) {
	// safe 校验
	// read csv values using csv.Reader
	csvReader := csv.NewReader(bytes.NewReader(binary))
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	sections := []string{}
	for i := 0; i < len(records); i++ {
		if len(records[i]) != 2 {
			return nil, fmt.Errorf("csv file format error")
		}
		q := records[i][0]
		a := records[i][1]
		sections = append(sections, q+"|||"+a)
	}

	return sections, nil
}
