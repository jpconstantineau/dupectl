package entities

import (
	"time"
)

type Status struct {
	id   int
	name string
}

type Host struct {
	id   int
	name string
}

type Owner struct {
	id   int
	name string
}

type Policy struct {
	id   int
	name string
}

type Purpose struct {
	id   int
	name string
}

type Agent struct {
	id      int
	name    string
	guid    string
	enabled bool
	status  Status
}

type RootFolder struct {
	id      int
	name    string
	host    Host
	owner   Owner
	agent   Agent
	purpose Purpose
	status  Status
}

type Folder struct {
	id      int
	name    string
	owner   Owner
	agent   Agent
	purpose Purpose
	status  Status
}

type filemsg struct {
	host     string
	fullname string
	path     string
	name     string
	ext      string
	size     int64
	atime    time.Time
	mtime    time.Time
	ctime    time.Time
	btime    time.Time
	hash     string
}

type foldermsg struct {
	fullname string
	path     string
	name     string
	atime    time.Time
	mtime    time.Time
	ctime    time.Time
	btime    time.Time
}
