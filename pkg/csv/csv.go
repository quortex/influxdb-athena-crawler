package csv

import (
	"encoding/csv"
	"io"
	"strings"
)

// parseString parses a CSV string to a map[string]interface{} slice
func ParseString(strCSV string) ([]map[string]interface{}, error) {
	// Read CSV object
	var header []string
	res := []map[string]interface{}{}

	reader := csv.NewReader(strings.NewReader(strCSV))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// First line contains header fields
		if header == nil {
			header = line
			continue
		}

		// Other lines contains rows
		row := make(map[string]interface{}, len(header))
		for i, e := range line {
			row[header[i]] = e
		}
		res = append(res, row)
	}

	return res, nil
}
