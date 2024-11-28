package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"gh-migrate-lfs/pkg/sync"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync LFS content to migrated repositories",
	Long:  "Sync LFS content to migrated repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := cmd.Flag("input").Value.String()
		workDir := cmd.Flag("work-dir").Value.String()
		newOrg := cmd.Flag("new-org").Value.String()
		token := cmd.Flag("token").Value.String()

		// Read CSV file
		file, err := os.Open(inputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer file.Close()

		reader := csv.NewReader(file)
		// Skip header
		if _, err := reader.Read(); err != nil {
			return fmt.Errorf("error reading CSV header: %w", err)
		}

		var processed, failed int
		// Process each repository
		for {
			record, err := reader.Read()
			if err != nil {
				break
			}

			repoName := record[0]
			if err := sync.SyncLFSContent(repoName, workDir, newOrg, token); err != nil {
				fmt.Printf("❌ Error syncing %s: %v\n", repoName, err)
				failed++
				continue
			}
			processed++
		}

		fmt.Printf("\n📊 Sync Summary:\n")
		fmt.Printf("✅ Successfully processed: %d repositories\n", processed)
		fmt.Printf("❌ Failed: %d repositories\n", failed)

		if failed > 0 {
			return fmt.Errorf("failed to sync %d repositories", failed)
		}
		return nil
	},
}

func init() {
	SyncCmd.Flags().StringP("input", "i", "lfs-repos.csv", "Input CSV file")
	SyncCmd.Flags().StringP("work-dir", "w", "", "Working directory with cloned repositories (required)")
	SyncCmd.Flags().StringP("new-org", "o", "", "New organization name (required)")
	SyncCmd.Flags().StringP("token", "t", "", "GitHub token with repo scope (required)")

	SyncCmd.MarkFlagRequired("work-dir")
	SyncCmd.MarkFlagRequired("new-org")
	SyncCmd.MarkFlagRequired("token")
}
