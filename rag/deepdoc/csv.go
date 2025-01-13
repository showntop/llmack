package deepdoc

import (
	"bytes"
	"encoding/csv"
)

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
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data[0], nil
}
