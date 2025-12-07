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

type Reels struct {
	Data struct {
		Clips struct {
			Edges []struct {
				Node struct {
					Media struct {
						Code string `json:"code"`
					} `json:"media"`
				} `json:"node"`
			} `json:"edges"`
			PageInfo struct {
				EndCursor       string `json:"end_cursor"`
				HasNextPage     bool   `json:"has_next_page"`
				HasPreviousPage bool   `json:"has_previous_page"`
				StartCursor     any    `json:"start_cursor"`
			} `json:"page_info"`
		} `json:"xdt_api__v1__clips__user__connection_v2"`
	} `json:"data"`
}

type Reel struct {
	Data struct {
		Media struct {
			Items []struct {
				Video []struct {
					Url string `json:"url"`
				} `json:"video_versions"`
			} `json:"items"`
		} `json:"xdt_api__v1__media__shortcode__web_info"`
	} `json:"data"`
}

func GetInitialUserReels(id string) ([]string, bool, string) {
	endpoint := "https://www.instagram.com/graphql/query"

	variables := fmt.Sprintf(`{"data":{"include_feed_video":true,"page_size":12,"target_user_id":"%s"}}`, id)

	form := url.Values{}
	form.Set("variables", variables)
	form.Set("doc_id", "24127588873492897")

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

	var reels Reels
	json.Unmarshal(body, &reels)

	codes := make([]string, 0, len(reels.Data.Clips.Edges))
	for _, edge := range reels.Data.Clips.Edges {
		codes = append(codes, edge.Node.Media.Code)
	}

	return codes, reels.Data.Clips.PageInfo.HasNextPage, reels.Data.Clips.PageInfo.EndCursor
}

func GetNextUserReels(id string, cursor string) ([]string, bool, string) {
	endpoint := "https://www.instagram.com/graphql/query"

	variablesJSON := fmt.Sprintf(
		`{"after":"%s","before":null,"data":{"include_feed_video":true,"page_size":12,"target_user_id":"%s"},"first":3,"last":null}`,
		cursor,
		id,
	)

	form := url.Values{}
	form.Set("doc_id", "9905035666198614")
	form.Set("variables", variablesJSON)

	payload := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", endpoint, payload)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:144.0) Gecko/20100101 Firefox/144.0")
	req.Header.Set("X-CSRFToken", util.LoadCookies()[0])
	req.Header.Set("Cookie", util.LoadCookies()[1])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var reels Reels
	json.Unmarshal(body, &reels)

	codes := make([]string, 0, len(reels.Data.Clips.Edges))
	for _, edge := range reels.Data.Clips.Edges {
		codes = append(codes, edge.Node.Media.Code)
	}

	return codes, reels.Data.Clips.PageInfo.HasNextPage, reels.Data.Clips.PageInfo.EndCursor
}

func DownloadReels(codes []string, id string, username string) {
	err := os.MkdirAll(fmt.Sprintf("downloads/%s-%s/reels", id, username), os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	dupeCtr := 0

	for i, code := range codes {
		if dupeCtr < 3 {
			if _, err := os.Stat(fmt.Sprintf("downloads/%s-%s/reels/%s.mp4", id, username, code)); errors.Is(err, os.ErrNotExist) {
				endpoint := "https://www.instagram.com/graphql/query"

				variables := fmt.Sprintf(`{"shortcode": "%s"}`, code)

				form := url.Values{}
				form.Set("variables", variables)
				form.Set("doc_id", "25152081234423663")

				payload := strings.NewReader(form.Encode())

				req, _ := http.NewRequest("POST", endpoint, payload)

				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:143.0) Gecko/20100101 Firefox/143.0")
				req.Header.Set("X-CSRFToken", util.LoadCookies()[0])
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Cookie", util.LoadCookies()[1])

				file, _ := os.Create(fmt.Sprintf("downloads/%s-%s/reels/%s.mp4", id, username, code))

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				defer res.Body.Close()

				body, _ := io.ReadAll(res.Body)

				var reel Reel
				json.Unmarshal(body, &reel)

				res, err = http.Get(reel.Data.Media.Items[0].Video[0].Url)
				if err != nil {
					log.Fatal(err)
				}
				defer res.Body.Close()

				_, err = io.Copy(file, res.Body)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("\rDownloading reels | Segment %d/%d      \x1b[?25l", i+1, len(codes))
			} else {
				dupeCtr++
			}
		} else {
			fmt.Println("Most recent reels downloaded, exiting")
			os.Exit(0)
		}
	}
}
