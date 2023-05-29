package api

import (
	"fmt"
	"net/http"

	"github.com/jpconstantineau/dupectl/pkg/auth"
	"github.com/spf13/viper"
)

func ApiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "super secret area")
}

func RunApi() {
	port := viper.GetString("server.port")

	http.Handle("/api", auth.ValidateJWT(ApiHome))
	http.HandleFunc("/register", auth.GetJWT)
	fmt.Println("serving on port " + port)
	http.ListenAndServe(":"+port, nil)
}
