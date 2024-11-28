package pull

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

func PullLFSContent(repoName, cloneURL, token, workDir string) error {
	repoPath := filepath.Join(workDir, repoName)

	// Create working directory if it doesn't exist
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("❌ Failed to create working directory: %w", err)
	}

	// Check if the repository already exists
	if _, err := os.Stat(repoPath); err == nil {
		pterm.Info.Printf("Repository exists '%s', proceeding with update\n", repoName)

		pullCmd := exec.Command("git", "pull", "--all")
		pullCmd.Dir = repoPath
		if output, err := pullCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("❌ Failed to pull updates: %s, %w", string(output), err)
		}

		lfsPullCmd := exec.Command("git", "lfs", "pull")
		lfsPullCmd.Dir = repoPath
		if output, err := lfsPullCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("❌ Failed to pull LFS content: %s, %w", string(output), err)
		}

		pterm.Success.Printf("Synchronization with '%s' completed successfully\n", repoName)
		return nil
	}

	// Clone the repository with GIT_LFS_SKIP_SMUDGE to avoid large file download during clone
	pterm.Info.Printf("Cloning repository '%s'...\n", repoName)
	cloneCmd := exec.Command("git", "clone", cloneURL)
	cloneCmd.Dir = workDir
	cloneCmd.Env = append(os.Environ(), "GIT_LFS_SKIP_SMUDGE=1")
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		errMsg := strings.ReplaceAll(string(output), token, "****")
		return fmt.Errorf("❌ Failed to clone repository: %s, %w", errMsg, err)
	}

	pterm.Info.Printf("Pulling LFS objects for repository '%s'...\n", repoName)

	// Pull LFS content using the environment token (no URL modification)
	lfsPullCmd := exec.Command("git", "lfs", "pull")
	lfsPullCmd.Dir = repoPath
	if output, err := lfsPullCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("❌ Failed to pull LFS content: %s, %w", string(output), err)
	}

	pterm.Success.Printf("synchronized: %s\n", repoName)
	return nil
}

func PullLFSFromCSV(inputFile, token, workDir string) (int, int, error) {
	spinner, _ := pterm.DefaultSpinner.Start(fmt.Sprintf("pulling repo and LFS data..."))

	// Read CSV file
	file, err := os.Open(inputFile)
	if err != nil {
		return 0, 0, fmt.Errorf("error opening input file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	if _, err := reader.Read(); err != nil {
		return 0, 0, fmt.Errorf("error reading CSV header: %w", err)
	}

	type repoInfo struct {
		name     string
		cloneURL string
	}

	seen := make(map[string]bool)
	records := make([]repoInfo, 0)
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, 0, fmt.Errorf("error reading CSV record: %w", err)
		}
		if len(record) != 3 {
			return 0, 0, fmt.Errorf("invalid CSV record format, expected 3 columns got %d", len(record))
		}
		if seen[record[0]] {
			continue
		}
		seen[record[0]] = true
		records = append(records, repoInfo{
			name:     record[0],
			cloneURL: record[2],
		})
	}

	var (
		processed, failed int32
		wg                sync.WaitGroup
		maxWorkers        = viper.GetInt("WORKERS")
		jobs              = make(chan repoInfo, len(records)) // Make buffered to prevent blocking
	)

	// Send jobs to workers
	processedRepos := sync.Map{}
	for _, repo := range records {
		if _, loaded := processedRepos.LoadOrStore(repo.name, true); loaded {
			continue // Repository already processed, skip it
		}
		jobs <- repo
	}
	close(jobs) // Close the channel to signal workers no more jobs are coming

	// Start worker pool
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repo := range jobs {
				// Insert token into clone URL
				urlParts := strings.SplitN(repo.cloneURL, "://", 2)
				if len(urlParts) != 2 {
					fmt.Printf("\u274c Invalid clone URL format for %s\n", repo.name)
					atomic.AddInt32(&failed, 1)
					continue
				}
				authenticatedURL := fmt.Sprintf("%s://%s@%s", urlParts[0], token, urlParts[1])

				err := PullLFSContent(repo.name, authenticatedURL, token, workDir)
				if err != nil {
					fmt.Printf("\u274c Error processing %s: %v\n", repo.name, err)
					atomic.AddInt32(&failed, 1)
				} else {
					atomic.AddInt32(&processed, 1)
				}
			}
		}()
	}

	// Wait for all workers to complete
	wg.Wait()
	spinner.Success()
	return int(processed), int(failed), nil
}
