package modules

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
)

type Tweet struct {
	URL    string
	Medias []struct {
		Width   int
		Height  int
		Src     string
		IsVideo bool `json:"is_video"`
	}
	Caption string
}

func extractTwitterURL(s string) string {
	prefixes := []string{"x.com/", "twitter.com/"}
	for _, prefix := range prefixes {
		index := strings.Index(s, prefix)
		if index != -1 {
			// Directly return the trimmed URL
			return strings.TrimSpace(s[index:])
		}
	}
	// Return an empty string if no match is found
	return ""
}

	url := extractTwitterURL(message.Text)
	if url == "" {
		bot.SendMessage(telegoutil.Message(
			telegoutil.ID(message.Chat.ID),
			"No twitter url",
		))
		return
	}

	var tweet Tweet
	_, body, _ := fasthttp.Get(nil, fmt.Sprintf("https://smudgeapi.ruizlenato.duckdns.org/twitter?url=%s", url))
	json.Unmarshal(body, &tweet)

	// Create the slice with a length of 0 and a capacity of 10
	mediaItems := make([]telego.InputMedia, 0, 10)

	for _, media := range tweet.Medias {
		if media.IsVideo {
			mediaItems = append(mediaItems, telegoutil.MediaVideo(telegoutil.FileFromURL(media.Src)).WithWidth(media.Width).WithHeight(media.Height))
		} else {
			mediaItems = append(mediaItems, telegoutil.MediaPhoto(telegoutil.FileFromURL(media.Src)))
		}
	}

	if len(mediaItems) < 2 && mediaItems[0].MediaType() == "photo" && !message.LinkPreviewOptions.IsDisabled {
		return
	}

	if len(mediaItems) > 0 {
		for _, media := range tweet.Medias[:1] {
			if mediaItems[0].MediaType() == "photo" {
				mediaItems[0] = telegoutil.MediaPhoto(telegoutil.FileFromURL(media.Src)).WithCaption(tweet.Caption)
			} else {
				mediaItems[0] = telegoutil.MediaVideo(telegoutil.FileFromURL(media.Src)).WithWidth(media.Width).WithHeight(media.Height).WithCaption(tweet.Caption)
			}
		}
	}

	bot.SendMediaGroup(telegoutil.MediaGroup(
		telegoutil.ID(message.Chat.ID),
		mediaItems...,
	))
}
