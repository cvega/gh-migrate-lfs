package sync

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

func buildGitURL(org, repo, token string) string {
	return fmt.Sprintf("https://%s@github.com/%s/%s.git", token, org, repo)
}

func SyncFromCSV() error {
	spinner, _ := pterm.DefaultSpinner.Start("Syncing LFS content to target repositories...")

	// Get configuration from viper
	inputFile := viper.GetString("GHMLFS_FILE")
	workDir := viper.GetString("GHMLFS_WORK_DIR")
	targetOrg := viper.GetString("GHMLFS_TARGET_ORGANIZATION")
	token := viper.GetString("GHMLFS_TARGET_TOKEN")

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
		if err := SyncLFSContent(repoName, workDir, targetOrg, token); err != nil {
			fmt.Printf("Error syncing %s: %v\n", repoName, err)
			failed++
			continue
		}
		processed++
	}

	spinner.Success()

	fmt.Printf("\n📊 Sync Summary:\n")
	fmt.Printf("✅ Successfully processed: %d repositories\n", processed)
	fmt.Printf("❌ Failed: %d repositories\n", failed)

	if failed > 0 {
		return fmt.Errorf("failed to sync %d repositories", failed)
	}
	return nil
}

func SyncLFSContent(repoName, workDir, targetOrg, token string) error {
	repoPath := filepath.Join(workDir, repoName)
	newRemote := buildGitURL(targetOrg, repoName, token)

	fmt.Printf("Updating remote for %s...\n", repoName)

	// Remove existing origin
	removeCmd := exec.Command("git", "remote", "remove", "origin")
	removeCmd.Dir = repoPath
	if err := removeCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove origin: %w", err)
	}

	// Add new origin
	addCmd := exec.Command("git", "remote", "add", "origin", newRemote)
	addCmd.Dir = repoPath
	if err := addCmd.Run(); err != nil {
		// Mask token in error message
		errMsg := strings.ReplaceAll(err.Error(), token, "****")
		return fmt.Errorf("failed to add new origin: %s", errMsg)
	}

	fmt.Printf("Pushing LFS content for %s to new organization...\n", repoName)

	// Push LFS content to new remote
	pushCmd := exec.Command("git", "lfs", "push", "--all", "origin")
	pushCmd.Dir = repoPath
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push LFS content: %s, %w", string(output), err)
	}

	// Push all branches
	pushAllCmd := exec.Command("git", "push", "--all", "origin")
	pushAllCmd.Dir = repoPath
	if output, err := pushAllCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push all branches: %s, %w", string(output), err)
	}

	fmt.Printf("Successfully synced LFS content for %s\n", repoName)
	return nil
}
