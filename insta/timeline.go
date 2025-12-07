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

// media type 1 photo, type 2 video
type Timeline struct {
	Data struct {
		Timeline struct {
			Edges []struct {
				Node struct {
					Code  string `json:"code"`
					Image struct {
						Candidates []struct {
							URL string `json:"url"`
						} `json:"candidates"`
					} `json:"image_versions2"`
					ProductType   string `json:"product_type"`
					CarouselMedia []struct {
						MediaType int `json:"media_type"`
						Video     []struct {
							URL string `json:"url"`
						} `json:"video_versions"`
						Image struct {
							Candidates []struct {
								URL string `json:"url"`
							} `json:"candidates"`
						} `json:"image_versions2"`
					} `json:"carousel_media"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				EndCursor   string `json:"end_cursor"`
				HasNextPage bool   `json:"has_next_page"`
			} `json:"page_info"`
		} `json:"xdt_api__v1__feed__user_timeline_graphql_connection"`
	} `json:"data"`
}

type Post struct {
	Code  string
	Files []string
}

func GetInitialUserTimeline(username string) ([]Post, bool, string) {
	endpoint := "https://www.instagram.com/graphql/query"

	variables := fmt.Sprintf(`{"data":{"count":12,"include_reel_media_seen_timestamp":true,"include_relationship_info":true,"latest_besties_reel_media":true,"latest_reel_media":true},"username":"%s","__relay_internal__pv__PolarisIsLoggedInrelayprovider":true}`, username)

	form := url.Values{}
	form.Set("variables", variables)
	form.Set("doc_id", "24937007899300943")

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

	var timeline Timeline
	json.Unmarshal(body, &timeline)

	posts := make([]Post, 0, len(timeline.Data.Timeline.Edges))

	for _, edge := range timeline.Data.Timeline.Edges {
		p := Post{
			Code:  edge.Node.Code,
			Files: []string{},
		}

		switch edge.Node.ProductType {
		case "feed":
			p.Files = append(p.Files, edge.Node.Image.Candidates[0].URL)
		case "carousel_container":
			for _, media := range edge.Node.CarouselMedia {
				switch media.MediaType {
				case 1:
					p.Files = append(p.Files, media.Image.Candidates[0].URL)
				case 2:
					p.Files = append(p.Files, media.Video[0].URL)
				}
			}
		}

		posts = append(posts, p)
	}

	return posts, timeline.Data.Timeline.PageInfo.HasNextPage, timeline.Data.Timeline.PageInfo.EndCursor
}

func GetNextUserTimeline(username string, cursor string) ([]Post, bool, string) {
	endpoint := "https://www.instagram.com/graphql/query"

	variables := fmt.Sprintf(`{"after":"%s","before":null,"data":{"count":12,"include_reel_media_seen_timestamp":true,"include_relationship_info":true,"latest_besties_reel_media":true,"latest_reel_media":true},"first":12,"last":null,"username":"%s","__relay_internal__pv__PolarisIsLoggedInrelayprovider":true}`, cursor, username)

	form := url.Values{}
	form.Set("variables", variables)
	form.Set("doc_id", "25389305420706138")

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

	var timeline Timeline
	json.Unmarshal(body, &timeline)

	posts := make([]Post, 0, len(timeline.Data.Timeline.Edges))

	for _, edge := range timeline.Data.Timeline.Edges {
		p := Post{
			Code:  edge.Node.Code,
			Files: []string{},
		}

		switch edge.Node.ProductType {
		case "feed":
			p.Files = append(p.Files, edge.Node.Image.Candidates[0].URL)
		case "carousel_container":
			for _, media := range edge.Node.CarouselMedia {
				if media.MediaType == 1 {
					p.Files = append(p.Files, media.Image.Candidates[0].URL)
				} else if media.MediaType == 2 {
					p.Files = append(p.Files, media.Video[0].URL)
				}
			}
		}

		posts = append(posts, p)
	}

	return posts, timeline.Data.Timeline.PageInfo.HasNextPage, timeline.Data.Timeline.PageInfo.EndCursor
}

func DownloadUserPost(posts []Post, id string, username string) {
	os.MkdirAll(fmt.Sprintf("downloads/%s-%s/timeline", id, username), os.ModePerm)

	dupeCtr := 0

	for i, post := range posts {
		if dupeCtr < 3 {
			for idx, item := range post.Files {
				u, _ := url.Parse(item)

				ext := filepath.Ext(u.Path)

				if _, err := os.Stat(fmt.Sprintf("downloads/%s-%s/timeline/%s-%d%s", id, username, post.Code, idx, ext)); errors.Is(err, os.ErrNotExist) {

					file, _ := os.Create(fmt.Sprintf("downloads/%s-%s/timeline/%s-%d%s", id, username, post.Code, idx, ext))

					res, err := http.Get(item)
					if err != nil {
						log.Fatal(err)
					}

					_, err = io.Copy(file, res.Body)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Printf("\rDownloading %d/%d files | %d/%d post segment      \x1b[?25l", idx+1, len(post.Files), i+1, len(posts))
				} else {
					dupeCtr++
				}
			}
		} else {
			fmt.Println("Most recent posts downloaded, exiting")
			os.Exit(0)
		}
	}
}
