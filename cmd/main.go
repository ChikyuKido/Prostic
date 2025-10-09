package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"prostic/internal/backup"
	"prostic/internal/config"
	"prostic/internal/util"
)

var log = util.GroupLogger("MAIN")

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	verbose := flag.Bool("v", false, "enable verbose logging")
	logPath := flag.String("logpath", "/tmp/vmrestic.log", "path to log file")
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
	err = backup.RunBackup()
	if err != nil {
		log.Fatalf("failed to run backup: %v", err)
	}
}
