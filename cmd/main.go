package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"prostic/internal/backup"
	"prostic/internal/config"
	"prostic/internal/restic"
	"prostic/internal/util"
)

var log = util.GroupLogger("MAIN")

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	verbose := flag.Bool("v", false, "enable verbose logging")
	logPath := flag.String("logpath", "/tmp/vmrestic.log", "path to log file")
	runRestic := flag.Bool("restic", false, "run restic command instead of backup")
	runBackup := flag.Bool("backup", false, "run backup command instead of restore")
	runStatus := flag.Bool("status", false, "run backup status command instead of restore")
	runList := flag.Bool("list", false, "run backup list command instead of restore")
	flag.Parse()

	logFile, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed to open log file %s: %v", *logPath, err)
	}
	defer logFile.Close()
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logrus.SetOutput(multiWriter)
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	err = config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if *runRestic {
		log.Info("Running restic command mode")
		args := flag.Args()
		if len(args) == 0 {
			log.Fatal("no restic arguments provided")
		}
		err = restic.RunResticCommand(true, args...)
		if err != nil {
			log.Fatalf("restic command failed: %v", err)
		}
		return
	} else if *runBackup {
		err = backup.RunBackup()
		if err != nil {
			log.Fatalf("failed to run backup: %v", err)
		}
		return
	} else if *runStatus {
		err = backup.PrintStats()
		if err != nil {
			log.Fatalf("failed to print stats: %v", err)
		}
	} else if *runList {
		err = backup.ListBackups()
		if err != nil {
			log.Fatalf("failed to list backups: %v", err)
		}
	} else {
		log.Errorf("No execute flag set like: backup,restore,status,restic")
	}

}
