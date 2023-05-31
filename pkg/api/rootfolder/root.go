package api

import (
	"fmt"
	"net/http"
)

func HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Fprintf(w, "super secret Create Root")
	}
	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "super secret Read Root")
	}
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "super secret Update Root")
	}
	if r.Method == http.MethodDelete {
		fmt.Fprintf(w, "super secret Delete Root")
	}

}
