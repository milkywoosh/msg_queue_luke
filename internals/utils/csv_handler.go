package utils

import (
	"encoding/csv"
	"fmt"
	"io"
)

type CSVData struct {
	Columns []string
	Rows    [][]string
}

// accept any object that implement io.Reader
func CSVReader(fileWithoutHeader io.Reader) (*CSVData, error) {
	reader := csv.NewReader(fileWithoutHeader)
	reader.Comma = ';'

	columns, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV rows: %w", err)
	}

	return &CSVData{
		Columns: columns,
		Rows:    rows,
	}, nil
}
