package watch

type FileChanges struct {
	Modified  chan bool // Channel to get notified of modifications
	Truncated chan bool // Channel to get notified of truncations
	Deleted   chan bool // Channel to get notified of deletions
	Renamed   chan bool // Channel to get notified of renames
}

func NewFileChanges() *FileChanges {
	return &FileChanges{
		make(chan bool), make(chan bool), make(chan bool), make(chan bool)}
}

func (fc *FileChanges) NotifyModified() {
	sendOnlyIfEmpty(fc.Modified)
}

func (fc *FileChanges) NotifyTruncated() {
	sendOnlyIfEmpty(fc.Truncated)
}

func (fc *FileChanges) NotifyDeleted() {
	send(fc.Deleted)
}

func (fc *FileChanges) NotifyRenamed() {
	send(fc.Renamed)
}

// sendOnlyIfEmpty sends on a bool channel only if the channel has no
// backlog to be read by other goroutines. This concurrency pattern
// can be used to notify other goroutines if and only if they are
// looking for it (i.e., subsequent notifications can be compressed
// into one).
func sendOnlyIfEmpty(ch chan bool) {
	select {
	case ch <- true:
	default:
	}
}

// An inotify event can be delivered by a non-blocking sender before
// the receiver is ready, causing important events to be missed.
// If the receiver do not receive a delete event, it will no longer
// receive any events and will be waiting.
// Some important events need to be guaranteed delivery.
func send(ch chan bool) {
	ch <- true
}
