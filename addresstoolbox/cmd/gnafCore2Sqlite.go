package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	constants "github.com/jacobsonjn/goaddress/addresstoolbox/internal"
	sqlite "github.com/jacobsonjn/goaddress/addresstoolbox/internal/db"
	bytes "github.com/jacobsonjn/goaddress/addresstoolbox/internal/helpers"
	"github.com/spf13/cobra"
)

// gnafCore2SqliteCmd represents the gnafCore2Sqlite command
var gnafCore2SqliteCmd = &cobra.Command{
	Use:   "gnafCore2Sqlite",
	Short: "GNAF Core psv to sqlite",
	Long: `Given the GNAF Core psv file this command will create a sqlite database.
	This command processes the GNAF Core data and converts it into a SQLite database format, allowing for easier querying and manipulation of address data.`,
	Example: `address-toolbox gnafCore2Sqlite --input gnaf_core.psv --output gnaf_core.db`,

	Run: runGnafCore2Sqlite,
}

func runGnafCore2Sqlite(cmd *cobra.Command, args []string) {
	fmt.Println("gnafCore2Sqlite called")

	// Read the input flag
	inputFilePath, _ := cmd.Flags().GetString("input")

	// Check that the file extension is .psv
	fi := validateInputFilenamePath(inputFilePath, true)
	if fi == nil {
		return
	}

	// Get the directory of the input file
	dir := filepath.Dir(inputFilePath)
	// Get the file name with extension
	baseNameWithExt := filepath.Base(fi.Name())
	baseNameWithoutExt := baseNameWithExt[:len(baseNameWithExt)-len(filepath.Ext(baseNameWithExt))]

	// Read the input table name flag if set
	tableName, _ := cmd.Flags().GetString("tableName")
	if tableName == "" {
		tableName = filepath.Base(baseNameWithoutExt)
	}

	// Read the output flag
	outputFilePath, _ := cmd.Flags().GetString("output")
	// If the output filename is not specified, use the input filename with .sqlite extension
	if outputFilePath == "" {
		fileName := baseNameWithoutExt + ".sqlite"    // Strip extension and append .sqlite
		outputFilePath = filepath.Join(dir, fileName) // Join directory and new file name
	}

	// Initialize database
	gnafDB, err := sqlite.NewGnafDB(outputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer gnafDB.Close()

	// Drop the GNAF table if it exists
	if err := gnafDB.DropGnafTable(baseNameWithoutExt); err != nil {
		log.Fatal(err)
	}

	// Create the GNAF table
	if err := gnafDB.CreateGnafTable(baseNameWithoutExt); err != nil {
		log.Fatal(err)
	}

	// Optimize database for bulk inserts
	if err := gnafDB.OptimizeForBulkInsert(); err != nil {
		fmt.Printf("%sERROR: applying PRAGMA optimizations failed %v%s\n", constants.Red, err, constants.Reset)
	}

	if err := processGnafCoreFile(inputFilePath, gnafDB); err != nil {
		fmt.Printf("%sERROR: processing GNAF core file failed: %v%s\n", constants.Red, err, constants.Reset)
	}
}

func processGnafCoreFile(inputFilePath string, gnafDB *sqlite.GnafDB) error {
	const batchSize = 10_000 // Adjust based on memory and performance
	var batch []sqlite.GnafRecord
	var recordCount int64   // Track total number of records inserted
	startTime := time.Now() // Start timing

	// Callback to process headers
	processHeaders := func(headers []string) error {
		fmt.Printf("Headers (immediately available): %v\n", headers)
		return nil
	}

	// Callback to process EOF
	eof := func() error {
		// Insert any remaining records
		if len(batch) > 0 {
			if err := gnafDB.InsertGnafRecordsBatch(batch); err != nil {
				return fmt.Errorf("failed to insert final batch: %w", err)
			}
			recordCount += int64(len(batch))
			fmt.Printf("%sInserted final batch of %d records%s\n", constants.Cyan, len(batch), constants.Reset)
		}

		// Reset PRAGMAs to safer defaults
		if err := gnafDB.ResetPragmas(); err != nil {
			return fmt.Errorf("failed to reset PRAGMAs: %w", err)
		}

		// Calculate and log inserts per second
		elapsed := time.Since(startTime).Seconds()
		if elapsed > 0 {
			ips := float64(recordCount) / elapsed
			fmt.Printf("%sProcessed %d records in %.2f seconds (%.2f inserts/second)%s\n",
				constants.Green, recordCount, elapsed, ips, constants.Reset)
		} else {
			fmt.Printf("%sProcessed %d records (time too short to calculate IPS)%s\n",
				constants.Green, recordCount, constants.Reset)
		}

		return nil
	}

	// Callback to process records
	processRecord := func(record []string) error {
		if len(record) < 27 {
			fmt.Printf("%sSkipping invalid record with %d fields (PID: %s)%s\n",
				constants.Yellow, len(record), record[0], constants.Reset)
			return nil
		}

		// Convert latitude and longitude to float64
		longitude, err := strconv.ParseFloat(record[25], 64)
		if err != nil {
			fmt.Printf("%sInvalid longitude in record %s: %v%s\n", constants.Yellow, record[0], err, constants.Reset)
			return nil
		}
		latitude, err := strconv.ParseFloat(record[26], 64)
		if err != nil {
			fmt.Printf("%sInvalid latitude in record %s: %v%s\n", constants.Yellow, record[0], err, constants.Reset)
			return nil
		}

		gnafRecord := sqlite.GnafRecord{
			AddressDetailPID: record[0],
			DateCreated:      record[1],
			AddressLabel:     record[2],
			AddressSiteName:  record[3],
			BuildingName:     record[4],
			FlatType:         record[5],
			FlatNumber:       record[6],
			LevelType:        record[7],
			LevelNumber:      record[8],
			NumberFirst:      record[9],
			NumberLast:       record[10],
			LotNumber:        record[11],
			StreetName:       record[12],
			StreetType:       record[13],
			StreetSuffix:     record[14],
			LocalityName:     record[15],
			State:            record[16],
			Postcode:         record[17],
			LegalParcelID:    record[18],
			MBCode:           record[19],
			AliasPrincipal:   record[20],
			PrincipalPID:     record[21],
			PrimarySecondary: record[22],
			PrimaryPID:       record[23],
			GeocodeType:      record[24],
			Longitude:        longitude,
			Latitude:         latitude,
		}
		batch = append(batch, gnafRecord)

		// Insert batch when it reaches batchSize
		if len(batch) >= batchSize {
			if err := gnafDB.InsertGnafRecordsBatch(batch); err != nil {
				return fmt.Errorf("failed to insert batch: %w", err)
			}
			recordCount += int64(len(batch))
			fmt.Printf("%sInserted batch of %d records (total: %d)%s\n",
				constants.Cyan, len(batch), recordCount, constants.Reset)
			batch = nil // Reset batch
		}

		return nil
	}

	// Call the streaming function
	err := bytes.ReadSeparatedFileStream(inputFilePath, "|", true, processHeaders, processRecord, eof)
	return err
}

// isValidInputFilename checks if the input file is a valid .psv file and exists on disk
func validateInputFilenamePath(input string, printFileSizeToConsole bool) os.FileInfo {
	// Check that the input file is a .psv file
	if filepath.Ext(input) != ".psv" {
		fmt.Printf("%sERROR: Input file must be a .psv file%s\n", constants.Red, constants.Reset)
		return nil
	}

	// Check that the input file exists on disk
	fi, err := os.Stat(input)
	if os.IsNotExist(err) {
		fmt.Printf("%sERROR: Input file does not exist: %s%s\n", constants.Red, input, constants.Reset)
		return nil
	} else if err != nil {
		fmt.Printf("%sERROR: unable to check input file: %v%s\n", constants.Red, err, constants.Reset)
		return nil
	}

	// Print the file size in human-readable format
	if printFileSizeToConsole {
		fmt.Printf("%sFile size: %s%s\n", constants.Cyan, bytes.FormatBytesAsHumanReadable(fi.Size()), constants.Reset)
	}

	return fi
}

func init() {
	rootCmd.AddCommand(gnafCore2SqliteCmd)

	// Add flags for input and output files
	gnafCore2SqliteCmd.Flags().StringP("input", "i", "", "Input GNAF Core PSV file")
	gnafCore2SqliteCmd.Flags().StringP("tableName", "t", "", "Table name to insert data into")
	gnafCore2SqliteCmd.Flags().StringP("output", "o", "", "Output SQLite database file")

	gnafCore2SqliteCmd.MarkFlagRequired("input")
}
