package sync

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func buildGitURL(org, repo, token string) string {
	return fmt.Sprintf("https://%s@github.com/%s/%s.git", token, org, repo)
}

func SyncLFSContent(repoName, workDir, newOrg, token string) error {
	repoPath := filepath.Join(workDir, repoName)
	newRemote := buildGitURL(newOrg, repoName, token)

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
