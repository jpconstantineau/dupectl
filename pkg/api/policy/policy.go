package api

import (
	"fmt"
	"net/http"
)

func HandlePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Fprintf(w, "super secret Create Policy")
	}
	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "super secret Read Policy")
	}
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "super secret Update Policy")
	}
	if r.Method == http.MethodDelete {
		fmt.Fprintf(w, "super secret Delete Policy")
	}

}
