package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Id     string `json:"agent_id"`
	Secret string `json:"agent_secret"`
	// Temporary just for testing big XD
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedditPost struct {
	Subreddit string `json:"subreddit"`
	Text      string `json:"selftext"`
	Title     string `json:"title"`
	Img       string `json:"url"`
	FullName  string `json:"name"`
}

type RedditPostChild struct {
	Post RedditPost `json:"data"`
}

type FrontPageData struct {
	After    string            `json:"after"`
	PostList []RedditPostChild `json:"children"`
}

type FrontPageResponse struct {
	Data FrontPageData `json:"data"`
}

type RedditListingResponse struct {
	After    string       `json:"after"`
	PostList []RedditPost `json:"PostList"`
}

func getConfig() Config {
	jsonFile, _ := os.Open("config.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var conf Config

	json.Unmarshal(byteValue, &conf)

	return conf
}

func login(username string, password string) string {

	conf := getConfig()

	client := &http.Client{}

	// jsonStr := []byte(`{"grant_type": "password", "username": ` + username + `, "password":` + password + ` "lego123}`)

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)

	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))

	req.SetBasicAuth(conf.Id, conf.Secret)
	req.Header.Add("User-Agent", "Satan")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	respBody := string(body)

	var respData map[string]string

	json.Unmarshal([]byte(respBody), &respData)

	// fmt.Println(respData["access_token"])

	fmt.Println(respData["expires_in"])

	ret := string(respData["access_token"])

	return ret

}

func getFrontpage(token string) RedditListingResponse {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://oauth.reddit.com/.json?limit=5", nil)

	req.Header.Add("Authorization", "bearer "+token)
	req.Header.Add("User-Agent", "Satan")

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	var data FrontPageResponse

	body, _ := ioutil.ReadAll(resp.Body)

	bodyString := string(body)

	json.Unmarshal([]byte(bodyString), &data)

	var ret RedditListingResponse

	ret.After = data.Data.After

	for _, children := range data.Data.PostList {
		ret.PostList = append(ret.PostList, children.Post)
	}

	return ret

}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {

	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		var user LoginData

		err := c.BindJSON(&user)

		fmt.Println(user.Username)

		if err != nil {
			fmt.Println(err)
		}

		token := login(user.Username, user.Password)

		c.JSON(http.StatusOK, gin.H{
			"Token": token,
		})

	})

	r.GET("/frontpage", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		list := getFrontpage(token)

		c.JSON(http.StatusOK, list)

	})

	r.Run()

}
