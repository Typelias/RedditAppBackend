//package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Id     string `json:"agent_id"`
	Secret string `json:"agent_secret"`
}

func getConfig() Config {
	jsonFile, _ := os.Open("config.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var conf Config

	json.Unmarshal(byteValue, &conf)

	return conf
}

func main() {
	// conf := getConfig()

	client := &http.Client{}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", "1WSDo4tTq_XOcrIvsIycKJT985O_AQ")
	data.Set("redirect_uri", "https://github.com/Typelias/RedditAppFrontend")

	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))

	req.SetBasicAuth("kEVB3v54wz-xD9C94bWGnA", "")

	req.Header.Add("User-Agent", "Satan")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	respBody := string(body)

	var respData map[string]string

	json.Unmarshal([]byte(respBody), &respData)

	fmt.Println(respBody)

}
