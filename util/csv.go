package util

import (
	"encoding/csv"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
)

func ReadCsvFile(fileHeader *multipart.FileHeader, csvHeader []string) ([][]string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while opening CSV file : %v", err))
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error while reading CSV file : %v", err))
	}

	bodyRecord := make([][]string, 0)
	for i, line := range records {
		if i == 0 {
			for k, column := range line {
				if column != csvHeader[k] {
					return nil, errors.New(fmt.Sprintf("CSV header doesn't matches with validator : %v", strings.Join(csvHeader, ", ")))
				}
			}
		} else {
			bodyRecord = append(bodyRecord, line)
		}
	}
	return bodyRecord, nil
}
