package insta

import (
	"encoding/json"
	"errors"
	"fmt"
	"horsaen/instagrab/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Stories struct {
	Data struct {
		Feed struct {
			Stories []struct {
				// user id
				Id string `json:"id"`
				// ReelType string `json:"reel_type"`
				Story []struct {
					// story id
					Id        string `json:"id"`
					EpochTime int    `json:"taken_at"`
					MediaType int    `json:"media_type"`
					Video     []struct {
						Url string `json:"url"`
					} `json:"video_versions"`
					Image struct {
						Candidates []struct {
							Url string `json:"url"`
						} `json:"candidates"`
					} `json:"image_versions2"`
				} `json:"items"`
			} `json:"reels_media"`
		} `json:"xdt_api__v1__feed__reels_media"`
	} `json:"data"`
}

type Story struct {
	Id         string
	UploadTime int
	Url        string
}

func GetUserStories(ids []string) []Story {
	// allows for multiple ids to be passed in
	// returns all stories for all ids
	endpoint := "https://www.instagram.com/graphql/query"

	reelIDs := "["
	for i, id := range ids {
		reelIDs += fmt.Sprintf(`"%s"`, id)
		if i < len(ids)-1 {
			reelIDs += ", "
		}
	}
	reelIDs += "]"

	variables := fmt.Sprintf(`{"reel_ids_arr":%s}`, reelIDs)

	form := url.Values{}
	form.Set("variables", variables)
	form.Set("doc_id", "25196122473307472")

	payload := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", endpoint, payload)

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0")
	req.Header.Set("X-CSRFToken", util.LoadCookies()[0])
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", util.LoadCookies()[1])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var stories Stories
	json.Unmarshal(body, &stories)

	storiesList := make([]Story, 0)

	for _, story := range stories.Data.Feed.Stories {
		for _, item := range story.Story {
			s := Story{
				Id:         story.Id,
				UploadTime: item.EpochTime,
				Url:        "",
			}
			if item.MediaType == 2 {
				s.Url = item.Video[0].Url
			} else {
				s.Url = item.Image.Candidates[0].Url
			}
			storiesList = append(storiesList, s)
		}
	}

	return storiesList
}

func DownloadUserStory(story []Story, id, username string) {
	os.MkdirAll(fmt.Sprintf("downloads/%s-%s/story", id, username), os.ModePerm)

	dupeCtr := 0

	for i, item := range story {
		if dupeCtr < 3 {
			u, _ := url.Parse(item.Url)
			ext := filepath.Ext(u.Path)
			if _, err := os.Stat(fmt.Sprintf("downloads/%s-%s/story/%d%s", id, username, item.UploadTime, ext)); errors.Is(err, os.ErrNotExist) {
				file, _ := os.Create(fmt.Sprintf("downloads/%s-%s/story/%d%s", id, username, item.UploadTime, ext))
				res, err := http.Get(item.Url)
				if err != nil {
					log.Fatal(err)
				}

				_, err = io.Copy(file, res.Body)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("\rDownloading %d/%d stories      \x1b[?25l", i+1, len(story))
			} else {
				dupeCtr++
			}
		} else {
			fmt.Println("Most recent stories downloaded, exiting")
			os.Exit(0)
		}
	}
}
