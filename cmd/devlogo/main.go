package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
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
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return build()
		},
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Build HTML files from markdown and launch local HTTP-server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve()
		},
	}

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
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

func serve() error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if err := builder.Build(cfg); err != nil {
		return fmt.Errorf("build site: %w", err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.ServerPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	fmt.Printf("Serving at http://%s\n", addr)

	server := &http.Server{
		Handler: http.FileServer(http.Dir(cfg.OutputDir)),
	}

	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
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
