package cmd

import (
	"fmt"
	"strings"
	"time"

	"gh-migrate-lfs/pkg/pull"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Does a git clone and lfs pull on exported repositories",
	Long:  "Does a git clone and lfs pull on exported repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFile := cmd.Flag("file").Value.String()
		token := cmd.Flag("token").Value.String()
		hostname := cmd.Flag("hostname").Value.String()
		workDir := cmd.Flag("work-dir").Value.String()
		workers := cmd.Flag("workers").Value.String()

		if workers == "" {
			workers = "1"
		}

		if hostname != "" {
			hostname = strings.TrimPrefix(hostname, "http://")
			hostname = strings.TrimPrefix(hostname, "https://")
			hostname = strings.TrimSuffix(hostname, "/api/v3")
			hostname = strings.TrimSuffix(hostname, "/")
			hostname = fmt.Sprintf("https://%s/api/v3", hostname)
			fmt.Printf("\n🔗 Using: GitHub Enterprise Server: %s\n", hostname)
		} else {
			fmt.Println("\n📡 Using: GitHub.com")
		}

		httpProxy := viper.GetString("HTTP_PROXY")
		httpsProxy := viper.GetString("HTTPS_PROXY")
		if httpProxy != "" || httpsProxy != "" {
			fmt.Println("🔄 Proxy: ✅ Configured")
		} else {
			fmt.Printf("🔄 Proxy: ❌ Not configured\n\n")
		}

		start := time.Now()
		processed, failed, err := pull.PullLFSFromCSV(inputFile, token, workDir)
		if err != nil {
			return err
		}
		duration := time.Since(start)

		fmt.Printf("\n📊 Pull Summary:\n")
		fmt.Printf("✅ Successfully processed: %d repositories\n", processed)
		fmt.Printf("❌ Failed: %d repositories\n", failed)
		fmt.Printf("🕐 Total time: %v\n", duration.Round(time.Second))

		if failed > 0 {
			return fmt.Errorf("failed to process %d repositories", failed)
		}
		return nil
	},
}

func init() {
	PullCmd.Flags().StringP("file", "f", "", "Exported LFS repos file path, csv format (required)")
	PullCmd.MarkFlagRequired("file")

	PullCmd.Flags().StringP("hostname", "n", "", "GitHub Enterprise Server hostname URL (optional) Ex. https://github.example.com")

	PullCmd.Flags().StringP("token", "t", "", "GitHub token with repo scope (required)")
	PullCmd.MarkFlagRequired("token")

	PullCmd.Flags().StringP("work-dir", "d", "", "Working directory for cloned repositories (required)")
	PullCmd.MarkFlagRequired("work-dir")

	PullCmd.Flags().IntP("workers", "w", 1, "Number of concurrent workers to use")
	viper.BindPFlag("workers", PullCmd.Flags().Lookup("workers"))
}
