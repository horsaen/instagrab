package insta

import (
	"encoding/json"
	"fmt"
	"horsaen/instagrab/util"
	"io"
	"log"
	"net/http"
)

type User struct {
	Data struct {
		User struct {
			Username   string `json:"username"`
			FullName   string `json:"full_name"`
			ProfilePic string `json:"profile_pic_url"`
			Id         string `json:"id"`
		} `json:"user"`
	} `json:"data"`
}

func GetUserId(username string) string {
	endpoint := fmt.Sprintf("https://i.instagram.com/api/v1/users/web_profile_info/?username=%s", username)

	req, _ := http.NewRequest("GET", endpoint, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 9; GM1903 Build/PKQ1.190110.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/75.0.3770.143 Mobile Safari/537.36 Instagram 103.1.0.15.119 Android (28/9; 420dpi; 1080x2260; OnePlus; GM1903; OnePlus7; qcom; sv_SE; 164094539")
	req.Header.Set("Cookie", util.LoadCookies()[1])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var user User
	json.Unmarshal(body, &user)

	return user.Data.User.Id
}
