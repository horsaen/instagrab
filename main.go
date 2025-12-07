package main

import (
	"flag"
	"fmt"
	"horsaen/instagrab/insta"
	"horsaen/instagrab/util"
	"os"
)

func main() {
	username := flag.String("username", "", "Instagram username")
	id := flag.String("id", "", "Instagram user id, resilient to username changes")
	// poll := flag.Bool("poll", false, "Poll an account for changes based on mode")
	mode := flag.String("mode", "", "Operation mode")

	flag.Parse()

	util.InitConfDir()

	if *username == "" && *id == "" {
		fmt.Println("No username or UID provided.")
		os.Exit(1)
	}

	if *id == "" {
		*id = insta.GetUserId(*username)
	}

	switch *mode {
	case "id":
		fmt.Println(*username + " user id: " + *id)
	case "story":
		insta.DownloadUserStory(insta.GetUserStories([]string{*id}), *id, *username)
	case "reels":
		reels, hasNext, cursor := insta.GetInitialUserReels(*id)

		insta.DownloadReels(reels, *id, *username)

		for hasNext {
			reels, hasNext, cursor = insta.GetNextUserReels(*id, cursor)

			insta.DownloadReels(reels, *id, *username)
		}

		fmt.Println("Finished downloading all reels.")
	case "posts":
		posts, hasNext, cursor := insta.GetInitialUserTimeline(*username)

		insta.DownloadUserPost(posts, *id, *username)

		for hasNext {
			posts, hasNext, cursor = insta.GetNextUserTimeline(*username, cursor)

			insta.DownloadUserPost(posts, *id, *username)
		}

		fmt.Println("Finished downloading all posts.")
	case "highlights":
		ids := insta.GetUserHighlights(*id)

		hightlights := insta.GetHighlightVideos(ids)

		insta.DownloadHighlights(hightlights, *id, *username)

		fmt.Println("Finished downloading all highlights.")
	default:
		fmt.Println("not supported")
		os.Exit(1)
	}
}
