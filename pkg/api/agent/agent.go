package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/auth"
	"github.com/jpconstantineau/dupectl/pkg/datastore"
	"github.com/jpconstantineau/dupectl/pkg/entities"
)

func HandleAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Fprintf(w, "super secret Create Agents")
	}
	if r.Method == http.MethodGet {
		data, err := datastore.GetAgent()
		if err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			response, _ := json.Marshal(data)
			w.Header().Set("content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}
	}
	if r.Method == http.MethodPut {
		fmt.Fprintf(w, "super secret Update Agents")
	}
	if r.Method == http.MethodDelete {
		fmt.Fprintf(w, "super secret Delete Agents")
	}
}

func RegisterAgent(w http.ResponseWriter, r *http.Request) {
	if r.Header["Clientid"] == nil {
		return
	}
	if r.Header["Uniqueid"] == nil {
		return
	}
	clientid, err := auth.DecodeMachineID(r.Header["Clientid"][0])
	if err != nil {
		fmt.Println("Invalid ClientID")
		return
	}
	uniqueid := string(r.Header["Uniqueid"][0])
	agent := entities.Agent{Name: clientid, Guid: uniqueid, Updated: time.Now(), Status: entities.StatusRegistered}
	data, err := datastore.PostAgent(agent)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	/*	response, _ := json.Marshal(data)
		w.Header().Set("content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	*/
	fmt.Println("registered: ", data.Id, data.Name, data.Guid, data.Status.String())
}
