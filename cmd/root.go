// Package cmd defines the harness CLI using cobra.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/poorlyordered/writers_harness/internal/anthropic"
	"github.com/poorlyordered/writers_harness/internal/box"
	"github.com/poorlyordered/writers_harness/internal/config"
	"github.com/poorlyordered/writers_harness/internal/session"
	"github.com/poorlyordered/writers_harness/internal/storage"
	"github.com/poorlyordered/writers_harness/internal/utils"
)

var (
	cfgFile string

	// shared across commands, set in PersistentPreRunE
	cfg        *config.Config
	aiClient   *anthropic.Client
	boxClient  *box.Client
	localStore *storage.Local
	sessMgr    *session.Manager
	prompter   *utils.Loader
)

var rootCmd = &cobra.Command{
	Use:   "harness",
	Short: "Writing Harness — AI-assisted prose writing from idea to draft",
	Long: `Writing Harness guides you through four phases of story development:
  Phase 1 — Idea Generation (seed prompt)
  Phase 2 — Idea Expansion (Snowflake Method)
  Phase 3 — Control Card Completion
  Phase 4 — Prose Generation

All state is stored in Box. Run 'harness new-series' to begin.`,
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config.json)")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initClients()
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath(".")
	}
	viper.AutomaticEnv()
}

func initClients() error {
	path := cfgFile
	if path == "" {
		path = "config.json"
	}

	var err error
	cfg, err = config.Load(path)
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	aiClient = anthropic.New("", cfg.Anthropic.Model, cfg.Anthropic.MaxTokens)
	boxClient = box.New(cfg.Box.MCPServerURL, cfg.Box.MCPAuthToken)
	aiClient.SetMCPServer(boxClient.MCPServerParam())

	localStore = storage.NewLocal(cfg.Local.SyncFolder)
	sessMgr = session.NewManager(cfg.Local.ConfigFolder)
	prompter = utils.NewLoader("prompts")
	return nil
}
