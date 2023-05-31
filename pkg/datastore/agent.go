package datastore

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jpconstantineau/dupectl/pkg/entities"
	"github.com/spf13/viper"

	_ "modernc.org/sqlite"
)

func initAgentTable() error {
	db, err := startDb()
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	dbtype := viper.GetString("server.database.type")
	if dbtype == "sqlite" {
		if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			guid TEXT NOT NULL,
			enabled INTEGER,
			updated INTEGER,
			status INTEGER
			);
	`); err != nil {
			return err
		}
		fmt.Println("Table: agents")
	}
	return nil
}

func PostAgent(data entities.Agent) (entities.Agent, error) {
	db, err := startDb()
	if err != nil {
		return entities.Agent{}, err
	}
	defer db.Close()

	name := data.Name
	guid := data.Guid
	enabled := data.Enabled
	updated := time.Now().Unix()
	status := data.Status
	found := false

	// check if Agent Already exists
	selDB, err := db.Query("SELECT id, name, guid FROM agents WHERE name=? AND guid =?", name, guid)
	if err != nil {
		return entities.Agent{}, err
	}

	for selDB.Next() {
		var qid int
		var qname, qguid string
		err = selDB.Scan(&qid, &qname, &qguid)
		if err != nil {
			return entities.Agent{}, err
		}
		//fmt.Println("FOUND: Agent: " + qname + " | guid: " + qguid)
		found = true
	}
	selDB.Close()
	if found { // update if it does exist
		upForm, err := db.Prepare("UPDATE agents SET enabled=?, updated=?, status=? WHERE name=? AND guid=?")
		if err != nil {
			return entities.Agent{}, err
		}
		upForm.Exec(enabled, updated, status, name, guid)
		//fmt.Println("UPDATE: Agent: " + name + " | guid: " + guid)
		upForm.Close()
	} else { // insert if it doesn't exist
		insForm, err := db.Prepare("INSERT INTO agents(name, guid, enabled, updated, status ) VALUES(?,?,?,?,?)")
		if err != nil {
			return entities.Agent{}, err
		}
		insForm.Exec(name, guid, enabled, updated, status)
		//fmt.Println("INSERT: Agent: " + name + " | guid: " + guid)
		insForm.Close()
	}

	resDB, err := db.Query("SELECT id, name, guid, enabled, updated, status FROM agents WHERE name=? AND guid =?", name, guid)
	if err != nil {
		return entities.Agent{}, err
	}

	var response entities.Agent
	for resDB.Next() {
		var qid int
		var qname, qguid string
		var qenabled bool
		var updatedstr string
		var qstatus entities.StatusName

		err = resDB.Scan(&qid, &qname, &qguid, &qenabled, &updatedstr, &qstatus)
		if err != nil {
			return entities.Agent{}, err
		}

		i, err := strconv.ParseInt(updatedstr, 10, 64)
		if err != nil {
			return entities.Agent{}, err
		}
		updatedT := time.Unix(i, 0)

		response = entities.Agent{Id: qid, Name: qname, Guid: qguid, Enabled: qenabled, Updated: updatedT, Status: qstatus}

	}
	resDB.Close()

	return response, nil
}

func GetAgent() ([]entities.Agent, error) {
	db, err := startDb()
	if err != nil {
		return []entities.Agent{}, err
	}
	defer db.Close()

	selDB, err := db.Query("SELECT id, name, guid, status, enabled, updated FROM agents ORDER BY id DESC")
	if err != nil {
		return []entities.Agent{}, err
	}

	items := []entities.Agent{}
	for selDB.Next() {
		var id int
		var name, guid, updatedstr string
		var enabled bool
		var status entities.StatusName

		var updated time.Time

		err = selDB.Scan(&id, &name, &guid, &status, &enabled, &updatedstr)
		if err != nil {
			return []entities.Agent{}, err
		}

		i, err := strconv.ParseInt(updatedstr, 10, 64)
		if err != nil {
			return []entities.Agent{}, err
		}
		updated = time.Unix(i, 0)
		if err != nil {
			return []entities.Agent{}, err
		}
		var item = entities.Agent{Id: id, Name: name, Guid: guid, Enabled: enabled, Updated: updated, Status: status}
		items = append(items, item)
		fmt.Println(item.Id, item.Name, item.Guid, item.Enabled, item.Updated)
	}
	selDB.Close()
	return items, nil
}
