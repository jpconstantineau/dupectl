package apiclient

import (
	"io/ioutil"
	"net/http"

	"github.com/spf13/viper"
)

func saveToken(token string) {
	viper.Set("client.token", token)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".dupectl.yaml")
	viper.WriteConfigAs(".dupectl.yaml")
}

func RegisterClient() error {
	// get keys
	host := viper.GetString("client.apihost")
	port := viper.GetString("client.apiport")
	key := viper.GetString("client.apikey")
	clientid := viper.GetString("client.clientid")
	uniqueid := viper.GetString("client.uniqueid")
	protocol := "http://"
	url := protocol + host + ":" + port + "/api/agent/register"

	// form request
	client := &http.Client{
		Transport: &http.Transport{},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Key", key)
	req.Header.Add("Uniqueid", uniqueid)
	req.Header.Add("Clientid", clientid)

	// make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Process Response
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}
	// save token
	bodyString := string(bodyBytes)
	saveToken(bodyString)
	return nil
}
