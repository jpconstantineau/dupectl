package api

import (
	"fmt"
	"net/http"
)

func HandleOwner(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Fprintf(w, "super secret Create Owner")
	}
	if r.Method == http.MethodGet {
		fmt.Fprintf(w, "super secret Read Owner")
	}
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "super secret Update Owner")
	}
	if r.Method == http.MethodDelete {
		fmt.Fprintf(w, "super secret Delete Owner")
	}

}
