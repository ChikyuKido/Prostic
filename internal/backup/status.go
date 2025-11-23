package backup

import (
	"errors"
	"fmt"
	"prostic/internal/restic"
	"prostic/internal/util"
	"sort"
	"strconv"
	"strings"
	"time"
)

var statsLog = util.GroupLogger("STATS")

var (
	cBlue  = "\033[94m"
	cGreen = "\033[92m"
	cCyan  = "\033[96m"
	cBold  = "\033[1m"
	cReset = "\033[0m"
)

type VMGroup struct {
	VMID  int
	Name  string
	Snaps []restic.Snapshot
}

func getNewestBackupID(snaps []restic.Snapshot) string {
	if len(snaps) == 0 {
		return ""
	}
	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].Time.After(snaps[j].Time)
	})

	newest := snaps[0]
	for _, t := range newest.Tags {
		if strings.HasPrefix(t, "id=") {
			return strings.TrimPrefix(t, "id=")
		}
	}
	return ""
}

func filterByBackupID(snaps []restic.Snapshot, id string) []restic.Snapshot {
	var out []restic.Snapshot
	match := "id=" + id

	for _, s := range snaps {
		for _, t := range s.Tags {
			if t == match {
				out = append(out, s)
				break
			}
		}
	}
	return out
}
func groupSnapshotsByVM(snaps []restic.Snapshot) map[int]*VMGroup {
	m := map[int]*VMGroup{}

	for _, s := range snaps {
		vmid := -1
		var name string

		for _, t := range s.Tags {
			if strings.HasPrefix(t, "vm=") {
				vmid, _ = strconv.Atoi(strings.TrimPrefix(t, "vm="))
			}
			if strings.HasPrefix(t, "name=") {
				name = strings.TrimPrefix(t, "name=")
			}
		}

		if vmid < 0 {
			continue
		}

		g, ok := m[vmid]
		if !ok {
			g = &VMGroup{VMID: vmid, Name: name}
			m[vmid] = g
		}
		g.Snaps = append(g.Snaps, s)
	}

	return m
}

func extractAllSources(snaps []restic.Snapshot) []string {
	seen := map[string]bool{}
	var out []string

	for _, s := range snaps {
		for _, t := range s.Tags {
			if strings.HasPrefix(t, "srcFile=") {
				src := strings.TrimPrefix(t, "srcFile=")
				if !seen[src] {
					seen[src] = true
					out = append(out, src)
				}
			}
		}
	}

	sort.Strings(out)
	return out
}

func printRepoStats(s *restic.Stats) {
	fmt.Printf("%sRepository Statistics%s\n", cBold, cReset)
	fmt.Printf("  Size: %s%.2f GB%s\n", cGreen, float64(s.TotalSize)/1e9, cReset)
	fmt.Printf("  Uncompressed: %.2f GB\n", float64(s.TotalUncompressedSize)/1e9)
	fmt.Printf("  Ratio: %.2fx\n", s.CompressionRatio)
	fmt.Printf("  Blobs: %d\n", s.TotalBlobCount)
	fmt.Printf("  Snapshots: %d\n\n", s.SnapshotsCount)
}

func printVMStats(groups map[int]*VMGroup, backupID string) {
	fmt.Printf("%sBackup ID: %s%s\n\n", cBold, backupID, cReset)

	for _, g := range groups {
		fmt.Printf("%s- %s (%d)%s\n", cBlue, g.Name, g.VMID, cReset)

		if len(g.Snaps) == 0 {
			fmt.Println("    No snapshots")
			continue
		}

		sort.Slice(g.Snaps, func(i, j int) bool {
			return g.Snaps[i].Time.After(g.Snaps[j].Time)
		})

		newest := g.Snaps[0]
		fmt.Printf("    Most recent: %s%s%s\n", cGreen, newest.Time.Format(time.RFC3339), cReset)

		files := extractAllSources(g.Snaps)
		fmt.Println("    Source files:")
		for _, f := range files {
			fmt.Printf("      - %s\n", f)
		}

		fmt.Println()
	}
}

func PrintStats() error {
	stats, err := restic.GetStats()
	if err != nil {
		return err
	}

	snaps, err := restic.GetSnapshots()
	if err != nil {
		return err
	}

	backupID := getNewestBackupID(snaps)
	if backupID == "" {
		return errors.New("no backup_id found")
	}

	filtered := filterByBackupID(snaps, backupID)
	groups := groupSnapshotsByVM(filtered)

	printRepoStats(stats)
	printVMStats(groups, backupID)

	return nil
}
