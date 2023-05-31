package entities

import (
	"time"
)

type Host struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Owner struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Policy struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Purpose struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Agent struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Guid    string     `json:"uid"`
	Enabled bool       `json:"enabled"`
	Updated time.Time  `json:"updated"`
	Status  StatusName `json:"status"`
}

type RootFolder struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Host    Host       `json:"host"`
	Owner   Owner      `json:"owner"`
	Agent   Agent      `json:"agent"`
	Purpose Purpose    `json:"purpose"`
	Status  StatusName `json:"status"`
}

type Folder struct {
	Id      int        `json:"id"`
	Name    string     `json:"name"`
	Owner   Owner      `json:"owner"`
	Agent   Agent      `json:"agent"`
	Purpose Purpose    `json:"purpose"`
	Status  StatusName `json:"status"`
}

type foldermsg struct {
	Host     string    `json:"host"`
	Fullname string    `json:"fullname"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Atime    time.Time `json:"atime"`
	Mtime    time.Time `json:"mtime"`
	Ctime    time.Time `json:"ctime"`
	Btime    time.Time `json:"btime"`
}

type filemsg struct {
	Host     string    `json:"host"`
	Fullname string    `json:"fullname"`
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Ext      string    `json:"ext"`
	Size     int64     `json:"size"`
	Atime    time.Time `json:"atime"`
	Mtime    time.Time `json:"mtime"`
	Ctime    time.Time `json:"ctime"`
	Btime    time.Time `json:"btime"`
	Hash     string    `json:"hash"`
}
