package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gh-migrate-lfs/pkg/export"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exports a list of repositories with LFS files to a CSV file",
	Long:  "Exports a list of repositories with LFS files to a CSV file",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get parameters
		organization := cmd.Flag("organization").Value.String()
		token := cmd.Flag("token").Value.String()
		filePrefix := cmd.Flag("file-prefix").Value.String()
		hostname := cmd.Flag("hostname").Value.String()
		searchDepth, _ := cmd.Flags().GetInt("search-depth")

		if hostname != "" {
			// Clean the hostname by removing any protocol and api/v3 if present
			hostname = strings.TrimPrefix(hostname, "http://")
			hostname = strings.TrimPrefix(hostname, "https://")
			hostname = strings.TrimSuffix(hostname, "/api/v3")
			hostname = strings.TrimSuffix(hostname, "/")
			hostname = fmt.Sprintf("https://%s/api/v3", hostname)
		}

		if filePrefix == "" {
			filePrefix = organization
		}

		// Set ENV variables
		os.Setenv("GHMLFS_SOURCE_ORGANIZATION", organization)
		os.Setenv("GHMLFS_SOURCE_TOKEN", token)
		os.Setenv("GHMLFS_OUTPUT_FILE", filePrefix)
		os.Setenv("GHMLFS_SOURCE_HOSTNAME", hostname)
		os.Setenv("GHMLFS_SEARCH_DEPTH", strconv.Itoa(searchDepth))

		// Bind ENV variables in Viper
		viper.BindEnv("SOURCE_ORGANIZATION", "GHMLFS_SOURCE_ORGANIZATION")
		viper.BindEnv("SOURCE_TOKEN", "GHMLFS_SOURCE_TOKEN")
		viper.BindEnv("OUTPUT_FILE", "GHMLFS_OUTPUT_FILE")
		viper.BindEnv("SOURCE_HOSTNAME", "GHMLFS_SOURCE_HOSTNAME")
		viper.BindEnv("SEARCH_DEPTH", "GHMLFS_SEARCH_DEPTH")

		if hostname != "" {
			fmt.Printf("\n🔗 Using: GitHub Enterprise Server: %s\n", hostname)
		} else {
			fmt.Println("\n📡 Using: GitHub.com")
		}

		httpProxy := viper.GetString("HTTP_PROXY")
		httpsProxy := viper.GetString("HTTPS_PROXY")
		if httpProxy != "" || httpsProxy != "" {
			fmt.Printf("🔄 Proxy: ✅ Configured\n\n")
		} else {
			fmt.Printf("🔄 Proxy: ❌ Not configured\n\n")
		}

		stats, duration, err := export.ExportLFSRepos()
		if err != nil {
			return fmt.Errorf("failed to export: %w", err)
		}

		fmt.Printf("\n📊 Export Summary:\n")
		fmt.Printf("Total repositories found: %d\n", stats.Total)
		fmt.Printf("✅ Successfully processed: %d repositories\n", stats.Successful)
		fmt.Printf("❌ Failed to process: %d repositories\n", stats.Failed)
		fmt.Printf("🔍 Maximum search depth: %d\n", stats.Depth)
		fmt.Printf("🔍 Repositories with LFS: %d\n", stats.Found)
		fmt.Printf("📁 Output file: %s\n", stats.OutputFile)
		fmt.Printf("🕐 Total time: %v\n", duration.Round(time.Second))

		return nil
	},
}

func init() {
	ExportCmd.Flags().StringP("file-prefix", "f", "", "Output filenames prefix (optional, defaults to organization name)")

	ExportCmd.Flags().StringP("hostname", "n", "", "GitHub Enterprise Server hostname URL (optional) Ex. https://github.example.com")

	ExportCmd.Flags().StringP("organization", "o", "", "Organization to export (required)")
	ExportCmd.MarkFlagRequired("organization")

	ExportCmd.Flags().StringP("search-depth", "s", "1", "Search depth for .gitattributes file")

	ExportCmd.Flags().StringP("token", "t", "", "GitHub token (required)")
	ExportCmd.MarkFlagRequired("token")
}
