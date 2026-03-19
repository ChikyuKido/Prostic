package runner

import (
	"errors"
	"sync"
	"time"
)

var ErrBusy = errors.New("another job is already running")

type Status struct {
	Running   bool       `json:"running"`
	Kind      string     `json:"kind,omitempty"`
	Purpose   string     `json:"purpose,omitempty"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
}

type Handle struct {
	released bool
}

var (
	mu      sync.Mutex
	current *Status
)

func Start(kind string, purpose string) (*Handle, error) {
	mu.Lock()
	defer mu.Unlock()

	if current != nil {
		return nil, ErrBusy
	}

	startedAt := time.Now()
	current = &Status{
		Running:   true,
		Kind:      kind,
		Purpose:   purpose,
		StartedAt: &startedAt,
	}

	return &Handle{}, nil
}

func (h *Handle) Release() {
	mu.Lock()
	defer mu.Unlock()

	if h == nil || h.released {
		return
	}

	current = nil
	h.released = true
}

func GetStatus() Status {
	mu.Lock()
	defer mu.Unlock()

	if current == nil {
		return Status{Running: false}
	}

	copy := *current
	return copy
}
