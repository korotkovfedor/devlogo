package main

import (
	"fmt"
	"os"

	"github.com/korotkovfedor/devlogo/internal/builder"
	"github.com/korotkovfedor/devlogo/internal/config"
	"github.com/spf13/cobra"
)

const configPath = "config.yaml"

func main() {
	rootCmd := &cobra.Command{
		Use:   "devlogo",
		Short: "A static site generator for engineering logs, changelogs, and technical notes written in Markdown",
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build HTML files from markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			return build()
		},
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Build HTML files from markdown and launch local HTTP-server",
		Run: func(cmd *cobra.Command, args []string) {
			panic("unimplemented")
		},
	}

	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new markdown content file",
		Run: func(cmd *cobra.Command, args []string) {
			panic("unimplemented")
		},
	}

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(newCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func build() error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	return builder.Build(cfg)
}

func loadConfig(path string) (config.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	cfg, err := config.NewFromYAML(file)
	if err != nil {
		return config.Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}
