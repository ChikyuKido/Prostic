package snapshots

import (
	"encoding/json"
	"strconv"
	"strings"

	"prostic/internal/db/models"
	"prostic/internal/db/repo"
	"prostic/internal/restic"
)

func RefreshSnapshotCache() (int, error) {
	snapshots, err := restic.GetSnapshots()
	if err != nil {
		return 0, err
	}

	rows := make([]models.Snapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		rows = append(rows, mapSnapshot(snapshot))
	}

	if err := repo.ReplaceSnapshots(rows); err != nil {
		return 0, err
	}

	return len(rows), nil
}

func mapSnapshot(snapshot restic.Snapshot) models.Snapshot {
	tagsJSON, _ := json.Marshal(snapshot.Tags)
	pathsJSON, _ := json.Marshal(snapshot.Paths)
	tagMap := parseTags(snapshot.Tags)

	var vmID *int
	if rawVMID := tagMap["vm"]; rawVMID != "" {
		if parsed, err := strconv.Atoi(rawVMID); err == nil {
			vmID = &parsed
		}
	}

	return models.Snapshot{
		SnapshotID:   snapshot.ID,
		Time:         snapshot.Time,
		Hostname:     snapshot.Hostname,
		Tree:         snapshot.Tree,
		Paths:        string(pathsJSON),
		Tags:         string(tagsJSON),
		BackupID:     tagMap["id"],
		VMID:         vmID,
		Name:         tagMap["name"],
		VMType:       tagMap["vmtype"],
		SnapshotType: tagMap["type"],
		BackupDate:   tagMap["date"],
		DestFile:     tagMap["destFile"],
		SrcFile:      tagMap["srcFile"],
	}
}

func parseTags(tags []string) map[string]string {
	out := make(map[string]string, len(tags))
	for _, tag := range tags {
		key, value, found := strings.Cut(tag, "=")
		if !found {
			continue
		}
		out[key] = value
	}

	return out
}
