package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"prostic/internal/config"
	"prostic/internal/restic"
	"prostic/internal/server"
)

var (
	configPath string
)

func main() {
	app := &cli.App{
		Name:    "prostic",
		Usage:   "A Restic-based backup utility.",
		Version: "1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "config.yaml",
				Usage:       "Path to configuration file",
				Destination: &configPath,
			},
		},
		Before: func(c *cli.Context) error {
			if err := config.Load(configPath); err != nil {
				return cli.Exit("Failed to load config: "+err.Error(), 1)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
		Commands: []*cli.Command{
			{
				Name:  "server",
				Usage: "Start the web server",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "port",
						Value: 8080,
						Usage: "Port for the web server",
					},
				},
				Action: func(c *cli.Context) error {
					addr := fmt.Sprintf(":%d", c.Int("port"))
					if err := server.Start(addr); err != nil {
						return cli.Exit("Failed to start server: "+err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:      "restic",
				Usage:     "Run a raw restic command (e.g., restic snapshots)",
				ArgsUsage: "[restic args...]",
				Action: func(c *cli.Context) error {
					args := c.Args().Slice()
					if len(args) == 0 {
						return cli.Exit("No restic arguments provided", 1)
					}
					err := restic.RunResticCommand(true, args...)
					if err != nil {
						return cli.Exit("Restic command failed: "+err.Error(), 1)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
