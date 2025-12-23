package api

import (
	"fmt"
	"net/http"

	agent "github.com/jpconstantineau/dupectl/pkg/api/agent"
	host "github.com/jpconstantineau/dupectl/pkg/api/host"
	owner "github.com/jpconstantineau/dupectl/pkg/api/owner"
	policy "github.com/jpconstantineau/dupectl/pkg/api/policy"
	purpose "github.com/jpconstantineau/dupectl/pkg/api/purpose"
	root "github.com/jpconstantineau/dupectl/pkg/api/rootfolder"
	"github.com/jpconstantineau/dupectl/pkg/auth"
	"github.com/jpconstantineau/dupectl/pkg/web"
	"github.com/spf13/viper"
)

func ApiHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "super secret area")
}

func RunApi() {

	port := viper.GetString("server.apiport")

	http.Handle("/api", auth.ValidateJWT(ApiHome))
	http.Handle("/api/agent", auth.ValidateJWT(agent.HandleAgent))
	http.Handle("/api/host", auth.ValidateJWT(host.HandleHost))
	http.Handle("/api/owner", auth.ValidateJWT(owner.HandleOwner))
	http.Handle("/api/policy", auth.ValidateJWT(policy.HandlePolicy))
	http.Handle("/api/purpose", auth.ValidateJWT(purpose.HandlePurpose))
	http.Handle("/api/root", auth.ValidateJWT(root.HandleRoot))

	http.Handle("/api/agent/register", auth.RegisterJWT(agent.RegisterAgent))
	//http.HandleFunc("/register", auth.GetJWT)
	web.SetupStaticWeb()
	fmt.Println("serving on port " + port)
	http.ListenAndServe(":"+port, nil)
}

// agent
//   C: add (register)
//   R: get
//   U: ??
//   D: delete

// Host
//   C: add
//   R: get
//   U: apply???
//   D: delete

// Owner
//   C: add
//   R: get
//   U: apply???
//   D: delete

// Policy
//   C: add
//   R: get
//   U: apply???
//   D: delete

// Purpose
//   C: add
//   R: get
//   U: apply???
//   D: delete

// Root
//   C: add
//   R: get
//   U: apply???
//   D: delete

// Duplicates

// Marked for Deletion
