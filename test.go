//package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

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

func main() {
	jsonFile, _ := os.Open("dump2.json")

	defer jsonFile.Close()

	byteVal, _ := ioutil.ReadAll(jsonFile)

	jsonString := string(byteVal)

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
		} else {
			// fmt.Println("Not galery")
		}

		listing.PostList = append(listing.PostList, post)

		// fmt.Println("--------------------------------------")
		// fmt.Println(post)
		// fmt.Println("--------------------------------------\n")
	}

	// b, _ := json.Marshal(listing)

	r := gin.Default()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, listing)
	})

	r.Run()

}
