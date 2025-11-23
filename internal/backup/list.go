package backup

import (
	"errors"
	"fmt"
	"prostic/internal/restic"
	"sort"
	"strings"
	"time"
)

type BackupEntry struct {
	BackupID string
	Time     time.Time
}

func ListBackups() error {
	snaps, err := restic.GetSnapshots()
	if err != nil {
		return err
	}

	entries := collectBackupIDs(snaps)
	if len(entries) == 0 {
		return errors.New("no backups found")
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})

	printBackupList(entries)
	return nil
}

func collectBackupIDs(snaps []restic.Snapshot) []BackupEntry {
	m := map[string]time.Time{}

	for _, s := range snaps {
		var id string
		for _, t := range s.Tags {
			if strings.HasPrefix(t, "id=") {
				id = strings.TrimPrefix(t, "id=")
				break
			}
		}
		if id == "" {
			continue
		}

		if t, ok := m[id]; !ok || s.Time.After(t) {
			m[id] = s.Time
		}
	}

	var out []BackupEntry
	for id, t := range m {
		out = append(out, BackupEntry{BackupID: id, Time: t})
	}
	return out
}

func printBackupList(list []BackupEntry) {
	fmt.Printf("%sAvailable Backups%s\n\n", cBold, cReset)

	for _, e := range list {
		fmt.Printf("%s- %s%s\n", cBlue, e.BackupID, cReset)
		fmt.Printf("    %s%s%s\n", cGreen, e.Time.Format(time.RFC3339), cReset)
	}

	fmt.Println()
}
