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
	"github.com/tidwall/gjson"
)

type Config struct {
	Id     string `json:"agent_id"`
	Secret string `json:"agent_secret"`
}

type RedditListing struct {
	After    string       `json:"after"`
	PostList []RedditPost `json:"postList"`
}

type RedditPost struct {
	Subreddit     string   `json:"subreddit"`
	Text          string   `json:"selftext"`
	Title         string   `json:"title"`
	Img           string   `json:"url"`
	FullName      string   `json:"name"`
	AuthorName    string   `json:"authorName"`
	GalleryImages []string `json:"gallery"`
	PermaLink     string   `json:"permaLink"`
}

func getConfig() Config {
	jsonFile, _ := os.Open("config.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var conf Config

	json.Unmarshal(byteValue, &conf)

	return conf
}

func getToken(code string) (string, string) {
	conf := getConfig()

	client := &http.Client{}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "https://github.com/Typelias/RedditAppFrontend")
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))

	req.SetBasicAuth(conf.Id, "")

	req.Header.Add("User-Agent", "Satan")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	respBody := string(body)

	fmt.Println("Body:")
	fmt.Println(respBody)

	var respData map[string]string

	err = json.Unmarshal([]byte(respBody), &respData)

	if err != nil {
		print(err.Error())
	}

	for key, val := range respData {
		fmt.Println(key + "\t" + val)
	}

	fmt.Println(respData["access_token"], respData["refresh_token"])

	return respData["access_token"], respData["refresh_token"]

}

func convertJsonToListing(jsonString string) RedditListing {
	after := gjson.Get(jsonString, "data.after")

	children := gjson.Get(jsonString, "data.children")

	listing := RedditListing{
		After: after.String(),
	}

	for _, child := range children.Array() {
		subreddit := child.Get("data.subreddit")
		text := child.Get("data.selftext")
		title := child.Get("data.title")
		url := child.Get("data.url")
		fullname := child.Get("data.name")
		authorName := child.Get("data.author")
		permaLink := "http://reddit.com" + child.Get("data.permalink").String()

		post := RedditPost{
			Subreddit:  subreddit.String(),
			Text:       text.String(),
			Title:      title.String(),
			Img:        url.String(),
			FullName:   fullname.String(),
			AuthorName: authorName.String(),
			PermaLink:  permaLink,
		}

		mediaMetadata := child.Get("data.media_metadata")

		if mediaMetadata.Exists() {
			var gallery []string
			mediaMetadata.ForEach(func(key, value gjson.Result) bool {

				imgType := strings.Split(value.Get("m").String(), "/")[1]

				url := "https://i.redd.it/" + key.String() + "." + imgType
				//fmt.Println(url)
				gallery = append(gallery, url)
				return true
			})
			post.GalleryImages = gallery
		}

		listing.PostList = append(listing.PostList, post)
	}

	return listing

}

func getRefreshToken(refreshToken string) string {
	conf := getConfig()

	client := &http.Client{}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))

	req.SetBasicAuth(conf.Id, "")

	req.Header.Add("User-Agent", "Satan")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
	}

	respBody := string(body)

	fmt.Println("Body:")
	fmt.Println(respBody)

	var respData map[string]string

	err = json.Unmarshal([]byte(respBody), &respData)

	if err != nil {
		fmt.Println(err.Error())
	}

	return respData["access_token"]

}

func getFrontpage(token string) RedditListing {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://oauth.reddit.com/.json?limit=100", nil)

	req.Header.Add("Authorization", "bearer "+token)
	req.Header.Add("User-Agent", "Satan")

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	bodyString := string(body)

	return convertJsonToListing(bodyString)

}

type LoginData struct {
	Code string `json:"code"`
}

type RefreshResponse struct {
	Token string `json:"token"`
}

func main() {

	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		var user LoginData

		err := c.BindJSON(&user)

		if err != nil {
			fmt.Println(err)
		}

		code := user.Code

		token, refresh := getToken(code)

		c.JSON(http.StatusOK, gin.H{
			"Token":   token,
			"Refresh": refresh,
		})

	})

	r.GET("/frontpage", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		list := getFrontpage(token)

		c.JSON(http.StatusOK, list)

	})

	r.POST("/refresh", func(c *gin.Context) {
		refresh := c.GetHeader("Authorization")
		token := getRefreshToken(refresh)

		ret := RefreshResponse{
			Token: token,
		}

		c.JSON(http.StatusOK, ret)
	})

	r.Run()

}
