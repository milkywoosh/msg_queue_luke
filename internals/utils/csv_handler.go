package utils

import (
	"encoding/csv"
	"io"
)

type CSVData struct {
	Columns []string
	Rows    [][]string
}

// accept any object that implement io.Reader
func CSVReader(fileMultipart io.Reader) (*CSVData, error) {
	reader := csv.NewReader(fileMultipart)

	columns, err := reader.Read()
	if err != nil {
		return nil, err
	}

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return &CSVData{
		Columns: columns,
		Rows:    rows,
	}, nil
}
