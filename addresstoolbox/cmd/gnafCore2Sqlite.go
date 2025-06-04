/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// gnafCore2SqliteCmd represents the gnafCore2Sqlite command
var gnafCore2SqliteCmd = &cobra.Command{
	Use:   "gnafCore2Sqlite",
	Short: "GNAF Core psv to sqlite",
	Long: `Given the GNAF Core psv file this command will create a sqlite database.
	This command processes the GNAF Core data and converts it into a SQLite database format, allowing for easier querying and manipulation of address data.`,
	Example: `address-toolbox gnafCore2Sqlite --input gnaf_core.psv --output gnaf_core.db`,

	Run: executeCmd,
}

func executeCmd(cmd *cobra.Command, args []string) {
	fmt.Println("gnafCore2Sqlite called")

	//Read the input flag
	input, err := cmd.Flags().GetString("input")

	if err != nil {
		fmt.Println("Error reading input flag:", err)
		return
	}

	//check that the input file is a psv file
	if input == "" {
		fmt.Println("Input file is required")
		return
	}

	//Read the output flag
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		fmt.Println("Error reading output flag:", err)
		return
	}

}

func init() {
	rootCmd.AddCommand(gnafCore2SqliteCmd)

	//add flags for input and output files
	gnafCore2SqliteCmd.Flags().StringP("input", "i", "", "Input GNAF Core PSV file")
	gnafCore2SqliteCmd.Flags().StringP("output", "o", "", "Output SQLite database file")

	gnafCore2SqliteCmd.MarkFlagRequired("input")
}
