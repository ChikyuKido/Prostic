package restic

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"prostic/internal/config"
	"prostic/internal/util"
	"strings"
	"time"
)

type Snapshot struct {
	ID       string    `json:"id"`
	Time     time.Time `json:"time"`
	Tags     []string  `json:"tags"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Tree     string    `json:"tree"`
}
type Stats struct {
	TotalSize              int64   `json:"total_size"`
	TotalUncompressedSize  int64   `json:"total_uncompressed_size"`
	CompressionRatio       float64 `json:"compression_ratio"`
	CompressionProgress    int     `json:"compression_progress"`
	CompressionSpaceSaving float64 `json:"compression_space_saving"`
	TotalBlobCount         int64   `json:"total_blob_count"`
	SnapshotsCount         int64   `json:"snapshots_count"`
}

var resticLog = util.GroupLogger("RESTIC")

func RunResticCommand(showOutput bool, args ...string) error {
	env := os.Environ()
	for key, val := range config.Get().Restic.EnvVars {
		env = append(env, key+"="+val)
	}

	cmd := exec.Command("/usr/bin/restic", args...)
	resticLog.Debugf("Running RESTIC command: %s", strings.Join(cmd.Args, " "))
	cmd.Env = env

	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("restic command failed: %v", err)
		}
		return nil
	}

	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restic command failed: %v", err)
	}

	return nil
}
func RunResticOutput(args ...string) (string, error) {
	env := os.Environ()
	for key, val := range config.Get().Restic.EnvVars {
		env = append(env, key+"="+val)
	}
	cmd := exec.Command("/usr/bin/restic", args...)
	resticLog.Debugf("Running RESTIC command: %s", strings.Join(cmd.Args, " "))
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("restic command failed: %v", err)
	}
	return string(out), nil
}
func RunResticJSONStream(args []string, handler func(map[string]interface{})) error {
	hasJSON := false
	for _, a := range args {
		if a == "--json" {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		return fmt.Errorf("restic JSON stream requires an argument --json")
	}

	cmd := exec.Command("/usr/bin/restic", args...)
	e := os.Environ()
	for key, val := range config.Get().Restic.EnvVars {
		e = append(e, key+"="+val)
	}
	e = append(e, "RESTIC_PROGRESS_FPS=1")
	cmd.Env = e

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			line := s.Text()

			// skip unrelated messages for clean output
			if strings.HasPrefix(line, "subprocess /bin/dd:") {
				continue
			}
			if strings.Contains(line, "using parent snapshot") {
				continue
			}
			if strings.Contains(line, "old cache directories") {
				continue
			}
			fmt.Println(line)
		}
	}()

	dec := json.NewDecoder(stdout)

	for {
		var obj map[string]interface{}
		err := dec.Decode(&obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("json decode error: %w", err)
		}
		handler(obj)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("restic command failed: %w", err)
	}
	return nil
}

func GetSnapshots() ([]Snapshot, error) {
	out, err := RunResticOutput("snapshots", "--json")
	if err != nil {
		return nil, err
	}

	var snaps []Snapshot
	if err := json.Unmarshal([]byte(out), &snaps); err != nil {
		return nil, err
	}
	return snaps, nil
}

func GetStats() (*Stats, error) {
	out, err := RunResticOutput("stats", "--mode", "raw-data", "--json")
	if err != nil {
		return nil, err
	}

	var s Stats
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
