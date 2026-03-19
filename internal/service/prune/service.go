package prune

import (
	"fmt"
	"strings"
	"time"

	appconfig "prostic/internal/config"
	"prostic/internal/db/repo"
	"prostic/internal/restic"
	cacheservice "prostic/internal/service/cache"
	snapshotservice "prostic/internal/service/snapshots"
)

const TaskPurposePruneNotInConfig = "prune_not_in_config"
const TaskPurposeDeleteSnapshot = "delete_snapshot"
const TaskPurposeDeleteBackupID = "delete_backup_id"

type SnapshotCandidate struct {
	SnapshotID   string `json:"snapshotID"`
	Time         string `json:"time"`
	BackupID     string `json:"backupID"`
	VMID         *int   `json:"vmid"`
	Name         string `json:"name"`
	VMType       string `json:"vmType"`
	SnapshotType string `json:"snapshotType"`
	SrcFile      string `json:"srcFile"`
}

func PreviewNotInConfig() ([]SnapshotCandidate, error) {
	if _, err := snapshotservice.RefreshSnapshotCache(); err != nil {
		return nil, err
	}

	snapshots, err := repo.ListSnapshots()
	if err != nil {
		return nil, err
	}

	candidates := make([]SnapshotCandidate, 0)
	for _, snapshot := range snapshots {
		if appconfig.SnapshotExists(snapshot.SnapshotType, snapshot.VMID, snapshot.SrcFile) {
			continue
		}

		candidates = append(candidates, SnapshotCandidate{
			SnapshotID:   snapshot.SnapshotID,
			Time:         snapshot.Time.Format(time.RFC3339),
			BackupID:     snapshot.BackupID,
			VMID:         snapshot.VMID,
			Name:         snapshot.Name,
			VMType:       snapshot.VMType,
			SnapshotType: snapshot.SnapshotType,
			SrcFile:      snapshot.SrcFile,
		})
	}

	return candidates, nil
}

func PreviewBackupID(backupID string) ([]SnapshotCandidate, error) {
	if _, err := snapshotservice.RefreshSnapshotCache(); err != nil {
		return nil, err
	}

	snapshots, err := repo.ListSnapshots()
	if err != nil {
		return nil, err
	}

	candidates := make([]SnapshotCandidate, 0)
	for _, snapshot := range snapshots {
		if snapshot.BackupID != backupID {
			continue
		}

		candidates = append(candidates, SnapshotCandidate{
			SnapshotID:   snapshot.SnapshotID,
			Time:         snapshot.Time.Format(time.RFC3339),
			BackupID:     snapshot.BackupID,
			VMID:         snapshot.VMID,
			Name:         snapshot.Name,
			VMType:       snapshot.VMType,
			SnapshotType: snapshot.SnapshotType,
			SrcFile:      snapshot.SrcFile,
		})
	}

	return candidates, nil
}

func RunNotInConfig(candidates []SnapshotCandidate) (string, error) {
	return runDeleteSnapshots("Prune not in config", candidates)
}

func RunDeleteSnapshot(candidate SnapshotCandidate) (string, error) {
	return runDeleteSnapshots("Delete snapshot", []SnapshotCandidate{candidate})
}

func RunDeleteBackupID(candidates []SnapshotCandidate) (string, error) {
	return runDeleteSnapshots("Delete backup ID", candidates)
}

func runDeleteSnapshots(title string, candidates []SnapshotCandidate) (string, error) {
	var logs strings.Builder
	logs.WriteString(title)
	logs.WriteString("\n")
	logs.WriteString(fmt.Sprintf("Snapshots requested: %d\n\n", len(candidates)))

	snapshotIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		snapshotIDs = append(snapshotIDs, candidate.SnapshotID)
		logs.WriteString(fmt.Sprintf("- %s | %s %v | %s | %s\n",
			candidate.SnapshotID,
			candidate.VMType,
			candidate.VMID,
			candidate.Name,
			candidate.SrcFile,
		))
	}

	if len(snapshotIDs) == 0 {
		logs.WriteString("\nNo snapshots selected.\n")
		return logs.String(), nil
	}

	output, err := restic.DeleteSnapshots(snapshotIDs)
	if output != "" {
		logs.WriteString("\nrestic output:\n")
		logs.WriteString(output)
		if !strings.HasSuffix(output, "\n") {
			logs.WriteString("\n")
		}
	}
	if err != nil {
		logs.WriteString("\nDelete failed.\n")
		return logs.String(), err
	}

	refreshResult, err := cacheservice.RefreshAll()
	if err != nil {
		logs.WriteString("\nSnapshots deleted, but cache refresh failed.\n")
		return logs.String(), err
	}

	logs.WriteString(fmt.Sprintf("\nDeleted snapshots: %d\n", len(snapshotIDs)))
	logs.WriteString(fmt.Sprintf("Cache refresh snapshot count: %d\n", refreshResult.SnapshotCount))
	return logs.String(), nil
}
