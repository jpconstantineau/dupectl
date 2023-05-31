package api

import (
	"fmt"
	"net/http"
)

func HandlePurpose(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Fprintf(w, "super secret Create Purpose")
	}
	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "super secret Read Purpose")
	}
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "super secret Update Purpose")
	}
	if r.Method == http.MethodDelete {
		fmt.Fprintf(w, "super secret Delete Purpose")
	}

}
