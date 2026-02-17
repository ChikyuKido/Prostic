package main

import (
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"prostic/internal/backup"
	"prostic/internal/config"
	"prostic/internal/restic"
	"prostic/internal/util"
)

var log = util.GroupLogger("MAIN")

var (
	configPath string
	verbose    bool
	logPath    string
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
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "Enable verbose logging",
				Destination: &verbose,
			},
			&cli.StringFlag{
				Name:        "logpath",
				Value:       "/tmp/vmrestic.log",
				Usage:       "Path to log file",
				Destination: &logPath,
			},
		},
		Before: func(c *cli.Context) error {
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				logrus.Fatalf("failed to open log file %s: %v", logPath, err)
			}
			defer logFile.Close()

			multiWriter := io.MultiWriter(os.Stdout, logFile)
			logrus.SetOutput(multiWriter)
			if verbose {
				logrus.SetLevel(logrus.DebugLevel)
			} else {
				logrus.SetLevel(logrus.InfoLevel)
			}
			err = config.Load(configPath)
			if err != nil {
				return cli.Exit("Failed to load config: "+err.Error(), 1)
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "backup",
				Usage: "Run the full backup process",
				Action: func(c *cli.Context) error {
					err := backup.RunBackup()
					if err != nil {
						return cli.Exit("Failed to run backup: "+err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Print the current backup statistics",
				Action: func(c *cli.Context) error {
					err := backup.PrintStats()
					if err != nil {
						return cli.Exit("Failed to print stats: "+err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name:  "list",
				Usage: "List all existing backups",
				Action: func(c *cli.Context) error {
					err := backup.ListBackups()
					if err != nil {
						return cli.Exit("Failed to list backups: "+err.Error(), 1)
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
		log.Fatalf("Application error: %v", err)
	}
}
