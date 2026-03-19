package backups

import "prostic/internal/config"

type EventType string

const (
	EventRunStarted   EventType = "run_started"
	EventItemStarted  EventType = "item_started"
	EventItemProgress EventType = "item_progress"
	EventItemDone     EventType = "item_done"
	EventLog          EventType = "log"
	EventRunDone      EventType = "run_done"
	EventRunFailed    EventType = "run_failed"
)

type Item struct {
	VM       config.VM
	ItemType string
	SrcFile  string
	DestFile string
}

type Event struct {
	Type           EventType
	BackupID       string
	TotalItems     int
	CompletedItems int
	Item           *Item
	BytesDone      int64
	BytesTotal     int64
	Message        string
}

type Observer interface {
	OnEvent(Event)
}

type ObserverFunc func(Event)

func (fn ObserverFunc) OnEvent(event Event) {
	fn(event)
}

type noopObserver struct{}

func (noopObserver) OnEvent(Event) {}

func normalizeObserver(observer Observer) Observer {
	if observer == nil {
		return noopObserver{}
	}

	return observer
}

type consoleObserver struct{}

func (consoleObserver) OnEvent(event Event) {
	switch event.Type {
	case EventRunStarted:
		println("Starting backup:", event.BackupID)
	case EventItemStarted:
		if event.Item != nil {
			println("Backing up:", event.Item.SrcFile)
		}
	case EventItemDone:
		if event.Item != nil {
			println("Finished:", event.Item.SrcFile)
		}
	case EventLog:
		if event.Message != "" {
			println(event.Message)
		}
	case EventRunDone:
		println("Backup completed:", event.BackupID)
	case EventRunFailed:
		println("Backup failed:", event.Message)
	}
}
