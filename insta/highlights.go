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
	"strings"
)

type UserHighlights struct {
	Data struct {
		Highlights struct {
			Edges []struct {
				Node struct {
					Id string `json:"id"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"highlights"`
	} `json:"data"`
}

type Highlights struct {
	Data struct {
		Highlights struct {
			Edges []struct {
				Node struct {
					Title           string `json:"title"`
					LatestReelMedia int    `json:"latest_reel_media"`
					Items           []struct {
						Id            string `json:"id"`
						Code          string `json:"code"`
						MediaType     int    `json:"media_type"`
						VideoVersions []struct {
							Url string `json:"url"`
						} `json:"video_versions"`
						ImageVersions struct {
							Candidates []struct {
								Url string `json:"url"`
							} `json:"candidates"`
						} `json:"image_versions2"`
					} `json:"items"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"xdt_api__v1__feed__reels_media__connection"`
	} `json:"data"`
}

func GetUserHighlights(id string) []string {
	endpoint := "https://www.instagram.com/graphql/query"

	variables := fmt.Sprintf(`{"user_id":"%s"}`, id)

	form := url.Values{}
	form.Set("variables", variables)
	form.Set("doc_id", "9814547265267853")

	payload := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", endpoint, payload)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRFToken", util.LoadCookies()[0])
	req.Header.Set("Cookie", util.LoadCookies()[1])

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var highlights UserHighlights
	json.Unmarshal(body, &highlights)

	var highlightsArr = make([]string, 0)
	for _, h := range highlights.Data.Highlights.Edges {
		highlightsArr = append(highlightsArr, h.Node.Id)
	}
	return highlightsArr
}

func GetHighlightVideos(highlightIds []string) Highlights {
	endpoint := "https://www.instagram.com/graphql/query"

	highlightIdsArr := "["
	for i, id := range highlightIds {
		highlightIdsArr += fmt.Sprintf(`"%s"`, id)
		if i < len(highlightIds)-1 {
			highlightIdsArr += ", "
		}
	}
	highlightIdsArr += "]"

	variables := fmt.Sprintf(`{"initial_reel_id":"%s","reel_ids":%s,"first":%d,"last":null}`, highlightIds[0], highlightIdsArr, len(highlightIds))

	form := url.Values{}
	form.Set("variables", variables)
	form.Set("doc_id", "25300536909575943")

	payload := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", endpoint, payload)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-CSRFToken", util.LoadCookies()[0])
	req.Header.Set("Cookie", util.LoadCookies()[1])

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var highlights Highlights
	json.Unmarshal(body, &highlights)

	return highlights
}

func DownloadHighlights(highlights Highlights, id, username string) {
	os.MkdirAll(fmt.Sprintf("downloads/%s-%s/highlights", id, username), os.ModePerm)

	for ie, edge := range highlights.Data.Highlights.Edges {
		path := fmt.Sprintf("downloads/%s-%s/highlights/%s", id, username, edge.Node.Title)
		os.MkdirAll(path, os.ModePerm)

		dupeCtr := 0

		for ii, item := range edge.Node.Items {
			if dupeCtr < 3 {
				switch item.MediaType {
				// image
				case 1:
					if _, err := os.Stat(path + "/" + item.Id + ".jpg"); errors.Is(err, os.ErrNotExist) {
						file, _ := os.Create(path + "/" + item.Id + ".jpg")
						res, _ := http.Get(item.ImageVersions.Candidates[0].Url)

						_, err := io.Copy(file, res.Body)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Printf("\rDownloading %d/%d items | Downloaded %d/%d highlights      \x1b[?25l", ii+1, len(edge.Node.Items), ie+1, len(highlights.Data.Highlights.Edges))
					} else {
						dupeCtr++
					}
					// video
				case 2:
					if _, err := os.Stat(path + "/" + item.Id + ".mp4"); errors.Is(err, os.ErrNotExist) {
						file, _ := os.Create(path + "/" + item.Id + ".mp4")
						res, _ := http.Get(item.VideoVersions[0].Url)

						_, err := io.Copy(file, res.Body)
						if err != nil {
							log.Fatal(err)
						}
						fmt.Printf("\rDownloading %d/%d items | Downloaded %d/%d highlights      \x1b[?25l", ii+1, len(edge.Node.Items), ie+1, len(highlights.Data.Highlights.Edges))
					} else {
						dupeCtr++
					}
				}
			} else {
				break
			}
		}
	}
}
