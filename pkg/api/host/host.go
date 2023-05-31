package api

import (
	"fmt"
	"net/http"
)

func HandleHost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Fprintf(w, "super secret Create Hosts")
	}
	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "super secret Read Hosts")
	}
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "super secret Update Hosts")
	}
	if r.Method == http.MethodDelete {
		fmt.Fprintf(w, "super secret Delete Hosts")
	}

}
