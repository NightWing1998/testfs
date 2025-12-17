package testfs

import "github.com/NightWing1998/testfs/mockfs"

// Re-export core types so consumers can import github.com/NightWing1998/testfs directly.
type (
	MockFS    = mockfs.MockFS
	Event     = mockfs.Event
	EventType = mockfs.EventType
)

const (
	EventCreate   = mockfs.EventCreate
	EventWrite    = mockfs.EventWrite
	EventRemove   = mockfs.EventRemove
	EventRead     = mockfs.EventRead
	EventRename   = mockfs.EventRename
	EventMkdir    = mockfs.EventMkdir
	EventRmdir    = mockfs.EventRmdir
	EventOpen     = mockfs.EventOpen
	EventTruncate = mockfs.EventTruncate
	LevelTrace    = mockfs.LevelTrace
)

var (
	ErrNotDir   = mockfs.ErrNotDir
	ErrExists   = mockfs.ErrExists
	ErrIsDir    = mockfs.ErrIsDir
	ErrNotEmpty = mockfs.ErrNotEmpty
	ErrNotFile  = mockfs.ErrNotFile
	ErrInvalid  = mockfs.ErrInvalid
)

// New constructs a new in-memory mock filesystem.
func New() *MockFS {
	return mockfs.New()
}
