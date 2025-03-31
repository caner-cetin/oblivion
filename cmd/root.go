package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/caner-cetin/oblivion/internal"
	"github.com/caner-cetin/oblivion/internal/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cfg = &config.Config
var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "oblivion",
	Short: "deployment setup for cansu.dev",
	Long:  `A DEAD ROAD, A DARK SUN, NOW WAITS BEYOND OBLIIVIIOOOOOOOOOOOOOOOON`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	initConfig()
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVar(&cfgPath, "config", "", "toml path (default $HOME/.oblivion.toml)")
	rootCmd.AddCommand(getPostgresCmd())
	rootCmd.AddCommand(getVersionCmd())
	rootCmd.AddCommand(getKumaCmd())
	rootCmd.AddCommand(getStaticCmd())
	rootCmd.AddCommand(getNetworkCmd())
	rootCmd.AddCommand(getObserverCmd())
}

func initConfig() {
	if cfgPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		cfgPath = filepath.Join(home, ".oblivion.toml")
	}
	cfg.SetDefaults()
	contents, err := internal.ReadFile(cfgPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			var cfgBytes []byte
			if cfgBytes, err = toml.Marshal(cfg); err != nil {
				log.Fatal().Err(err).Msg("failed to marshal default config")
			}
			cfgFile, err := os.Create(cfgPath)
			if err != nil {
				log.Fatal().Str("path", cfgPath).Err(err).Msg("failed to create config file")
			}
			if _, err := io.Copy(cfgFile, bytes.NewReader(cfgBytes)); err != nil {
				log.Fatal().Err(err).Msg("failed to save default config")
			}
		} else {
			log.Fatal().Err(err).Str("path", cfgPath).Msg("failed to read config file")
		}
	}
	if err := toml.Unmarshal(contents, cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to unmarshal config")
	}
}
