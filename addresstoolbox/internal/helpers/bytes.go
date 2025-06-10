package helpers

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// humanReadableSize converts bytes to a human-readable format
func FormatBytesAsHumanReadable(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// ReadPipeSeparatedFileStream reads a pipe-separated file, calling processHeaders when headers are read
// and processRecord for each data record.
// Example usage:
/*
    //Callback to process headers
	processHeaders := func(headers []string) error {
		fmt.Printf("Headers (immediately available): %v\n", headers)
		return nil
	}

	// Callback to process records
	processRecord := func(record []string) error {
		fmt.Printf("Processing record: %v\n", record)
		return nil
	}

 	//Call the streaming function
	err := ReadPipeSeparatedFileStream(filepath, separator, firstLineHeaders, processHeaders, processRecord)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
*/
func ReadSeparatedFileStream(
	filepath, separator string,
	firstLineHeaders bool,
	processHeaders func(headers []string) error,
	processRecord func(record []string) error,
	eof func() error, // Optional EOF callback, can be nil
) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Wrap file with BOMStrippingReader
	bomReader := NewBOMStrippingReader(file, true)

	reader := csv.NewReader(bufio.NewReader(bomReader))
	reader.Comma = rune(separator[0])
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	lineCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			// Call EOF function if provided
			if eof != nil {
				if err := eof(); err != nil {
					return fmt.Errorf("failed to execute EOF function: %w", err)
				}
			}
			break
		}
		if err != nil {
			return fmt.Errorf("error parsing file at line %d: %w", lineCount+1, err)
		}
		if len(record) == 0 || (len(record) == 1 && record[0] == "") {
			continue
		}

		if firstLineHeaders && lineCount == 0 {
			// Call processHeaders immediately when headers are read
			if err := processHeaders(record); err != nil {
				return fmt.Errorf("error processing headers: %w", err)
			}
		} else {
			// Process data records
			if err := processRecord(record); err != nil {
				return fmt.Errorf("error processing record at line %d: %w", lineCount+1, err)
			}
		}
		lineCount++
	}

	return nil
}
