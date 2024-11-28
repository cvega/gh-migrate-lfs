package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"gh-migrate-lfs/internal/api"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

// RepoLFSInfo holds information about a repository containing LFS data
type RepoLFSInfo struct {
	Name     string
	Path     string
	CloneURL string
}

func ExportLFSRepos() (stats struct {
	Total      int
	Successful int
	Failed     int
	Found      int
	Depth      int
	OutputFile string
}, duration time.Duration, err error) {
	start := time.Now()
	spinner, _ := pterm.DefaultSpinner.Start("Searching for repositories with LFS content...")

	// Get configuration
	organization := viper.GetString("SOURCE_ORGANIZATION")
	token := viper.GetString("SOURCE_TOKEN")
	outputFile := viper.GetString("OUTPUT_FILE") + "_lfs.csv"
	depth := viper.GetInt("SEARCH_DEPTH")
	hostname := viper.GetString("SOURCE_HOSTNAME")

	if organization == "" || token == "" || outputFile == "" {
		return stats, duration, fmt.Errorf("missing required parameters: organization, token, or outputFile")
	}

	if depth == 0 {
		depth = 1 // Default depth if not specified
	}

	// Fetch repositories
	pterm.Info.Printf("Fetching repository list for %s...", organization)
	repos, err := api.GetRepositories(organization, token, hostname)
	if err != nil {
		return stats, duration, fmt.Errorf("failed to fetch repositories: %w", err)
	}
	pterm.Info.Printf("Found %d repositories\n", len(repos))

	// Process repositories and collect LFS information
	var lfsRepos []RepoLFSInfo
	var successful, failed, found int

	pterm.Info.Printf("Checking repositories for LFS content (searching up to depth %d)...", depth)

	for _, repo := range repos {
		pterm.Info.Printf("Searching repository contents: '%s'...\n", repo)

		hasLFS, path, err := api.CheckGitAttributes(organization, repo, token, depth, hostname)
		if err != nil {
			pterm.Info.Printf("Warning: Failed to determine LFS status for repo %s: %v", repo, err)
			failed++
			continue
		}

		if hasLFS {
			cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", organization, repo)
			if hostname != "" {
				cloneURL = fmt.Sprintf("%s/%s/%s.git", hostname, organization, repo)
			}

			lfsRepos = append(lfsRepos, RepoLFSInfo{
				Name:     repo,
				Path:     path,
				CloneURL: cloneURL,
			})
			found++
			pterm.Success.Printf("LFS filter matched for repository '%s' (path: %s)\n", repo, path)
		}

		successful++
	}

	// Write results to CSV file
	if err := writeToCSV(outputFile, lfsRepos); err != nil {
		return stats, duration, fmt.Errorf("failed to write CSV file: %w", err)
	}

	stats.Total = len(repos)
	stats.Successful = successful
	stats.Failed = failed
	stats.Found = found
	stats.Depth = depth
	stats.OutputFile = outputFile

	spinner.Success()
	return stats, time.Since(start), nil
}

func writeToCSV(filename string, repos []RepoLFSInfo) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Repository", "GitAttributesPaths", "CloneURL"}); err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	// Write data
	for _, repo := range repos {
		if err := writer.Write([]string{
			repo.Name,
			repo.Path,
			repo.CloneURL,
		}); err != nil {
			return fmt.Errorf("error writing repository data: %w", err)
		}
	}

	return nil
}
