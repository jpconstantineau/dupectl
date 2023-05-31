package entities

type StatusName int64

const (
	StatusUnknown StatusName = iota
	StatusRegistered
	StatusReady
	StatusNotReady
	StatusPending
	StatusScanning
	StatusProgressing
	StatusSynced
	StatusMissing
	StatusMarkForDeletion
	StatusReadyForDeletion
	StatusDeleting
	StatusDeleted
)

func (s StatusName) String() string {
	switch s {
	case StatusUnknown:
		return "Unknown"
	case StatusRegistered:
		return "Registered"
	case StatusReady:
		return "Ready"
	case StatusNotReady:
		return "Not Ready"
	case StatusPending:
		return "Pending"
	case StatusScanning:
		return "Scanning"
	case StatusProgressing:
		return "Progressing"
	case StatusSynced:
		return "Synced"
	case StatusMissing:
		return "Missing"
	case StatusMarkForDeletion:
		return "Marked for Deletion"
	case StatusReadyForDeletion:
		return "Ready for Deletion"
	case StatusDeleting:
		return "Deleting"
	case StatusDeleted:
		return "Deleted"
	}
	return "unknown"
}
