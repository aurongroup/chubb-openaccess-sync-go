package data

// SyncStatus represents the result of comparing a CSV record against the API.
type SyncStatus int

const (
	SyncNew SyncStatus = iota
	SyncExisting
	SyncUpdate
	SyncDelete
)

func (s SyncStatus) String() string {
	switch s {
	case SyncNew:
		return "new"
	case SyncExisting:
		return "existing"
	case SyncUpdate:
		return "update"
	case SyncDelete:
		return "delete"
	}
	return "unknown"
}
