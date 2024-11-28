package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "migrate-lfs",
	Short: "gh cli extension to migrate LFS files between git repositories",
	Long:  "gh cli extension to migrate LFS files between git repositories",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add root command flags (not persistent)
	rootCmd.PersistentFlags().String("http-proxy", "", "HTTP proxy (can also use HTTP_PROXY env var)")
	rootCmd.PersistentFlags().String("https-proxy", "", "HTTPS proxy (can also use HTTPS_PROXY env var)")
	rootCmd.PersistentFlags().String("no-proxy", "", "No proxy list (can also use NO_PROXY env var)")
	rootCmd.PersistentFlags().Int("max-retries", 3, "Maximum retry attempts")
	rootCmd.PersistentFlags().String("retry-delay", "1s", "Delay between retries")

	// Bind flags to viper
	viper.BindPFlag("HTTP_PROXY", rootCmd.PersistentFlags().Lookup("http-proxy"))
	viper.BindPFlag("HTTPS_PROXY", rootCmd.PersistentFlags().Lookup("https-proxy"))
	viper.BindPFlag("NO_PROXY", rootCmd.PersistentFlags().Lookup("no-proxy"))
	viper.BindPFlag("MAX_RETRIES", rootCmd.PersistentFlags().Lookup("max-retries"))
	viper.BindPFlag("RETRY_DELAY", rootCmd.PersistentFlags().Lookup("retry-delay"))

	// Add subcommands
	rootCmd.AddCommand(ExportCmd)
	rootCmd.AddCommand(PullCmd)
	rootCmd.AddCommand(SyncCmd)

	// hide -h, --help from global/proxy flags
	rootCmd.Flags().BoolP("help", "h", false, "")
	rootCmd.Flags().Lookup("help").Hidden = true
}

func initConfig() {
	viper.SetEnvPrefix("GHMLFS")
	viper.BindEnv("HTTP_PROXY")
	viper.BindEnv("HTTPS_PROXY")
	viper.BindEnv("NO_PROXY")
	viper.BindEnv("MAX_RETRIES")
	viper.AutomaticEnv()
}
