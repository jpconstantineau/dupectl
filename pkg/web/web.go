package web

import (
	"embed"
	_ "embed"
	"io/fs"
	"net/http"
)

//go:embed static
var embeddedFiles embed.FS

func SetupStaticWeb() {

	fstatic, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		panic(err)
	}
	http.Handle("/", http.FileServer(http.FS(fstatic)))

}
