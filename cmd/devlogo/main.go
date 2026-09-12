package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/korotkovfedor/devlogo/internal/config"
	"github.com/korotkovfedor/devlogo/internal/content"
	"github.com/spf13/cobra"
)

const configPath string = "config.yaml"

func main() {
	rootCmd := &cobra.Command{
		Use:   "devlogo",
		Short: "A static site generator for engineering logs, changelogs, and technical notes written in Markdown",
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build HTML files from markdown",
		Run: func(cmd *cobra.Command, args []string) {
			err := build()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
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
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	defer file.Close()

	cfg, err := config.NewFromYAML(file)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	entries, err := os.ReadDir(cfg.ContentDir)
	if err != nil {
		return fmt.Errorf("failed to read content dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		path := filepath.Join(cfg.ContentDir, entry.Name())
		page, err := parseEntry(path)
		if err != nil {
			return fmt.Errorf("process %q: %w", path, err)
		}

		fmt.Println(page)
	}

	return nil
}

func parseEntry(path string) (content.Content, error) {
	file, err := os.Open(path)
	if err != nil {
		return content.Content{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	result, err := content.ParseMarkdown(file)
	if err != nil {
		return content.Content{}, fmt.Errorf("parse markdown: %w", err)
	}

	return result, nil
}
