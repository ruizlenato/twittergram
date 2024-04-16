package modules

import (
	"fmt"
	"log"
	"os"
	"strings"

	"twittergram/twittergram/localization"
	"twittergram/twittergram/twitter"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
)

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

var mimeExtensions = map[string]string{
	"image/jpeg":      "jpg",
	"image/png":       "png",
	"image/gif":       "gif",
	"image/webp":      "webp",
	"video/mp4":       "mp4",
	"video/webm":      "webm",
	"video/quicktime": "mov",
	"video/x-msvideo": "avi",
}

func downloader(url string) (*os.File, error) {
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	client := &fasthttp.Client{ReadBufferSize: 16 * 1024}
	request.SetRequestURI(url)
	err := client.Do(request, response)
	if err != nil {
		log.Println(err)
	}

	extension := func(contentType []byte) string {
		extension, ok := mimeExtensions[string(contentType)]
		if !ok {
			return ""
		}
		return extension
	}

	file, err := os.CreateTemp("", fmt.Sprintf("twittergram*.%s", extension(response.Header.ContentType())))
	if err != nil {
		return nil, err
	}

	_, err = file.Write(response.Body()) // Write the byte slice to the file
	if err != nil {
		file.Close()
		return nil, err
	}

	_, err = file.Seek(0, 0) // Seek back to the beginning of the file
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, err
}

func MediaDownloader(bot *telego.Bot, message telego.Message) {
	url := extractTwitterURL(message.Text)
	if url == "" {
		bot.SendMessage(telegoutil.Message(
			telegoutil.ID(message.Chat.ID),
			"No twitter url",
		))
		return
	}

	tweetMedias := twitter.TweetMedias(url)
	if len(tweetMedias.Medias) < 1 {
		return
	}

	// Create the slice with a length of 0 and a capacity of 10
	mediaItems := make([]telego.InputMedia, 0, 10)

	for _, media := range tweetMedias.Medias {
		if file, err := downloader(media.Source); err == nil {
			if media.Video {
				mediaItems = append(mediaItems, telegoutil.MediaVideo(telegoutil.File(file)).WithWidth(media.Width).WithHeight(media.Height))
			} else {
				mediaItems = append(mediaItems, telegoutil.MediaPhoto(telegoutil.File(file)))
			}
		}
	}

	if len(mediaItems) < 2 && mediaItems[0].MediaType() == "photo" && !message.LinkPreviewOptions.IsDisabled {
		return
	}

	if len(mediaItems) > 0 {
		for _, media := range mediaItems[:1] {
			switch media.MediaType() {
			case "photo":
				if photo, ok := media.(*telego.InputMediaPhoto); ok {
					photo.WithCaption(tweetMedias.Caption).WithParseMode("HTML")
				}
			case "video":
				if video, ok := media.(*telego.InputMediaVideo); ok {
					video.WithCaption(tweetMedias.Caption).WithParseMode("HTML")
				}
			}
		}
	}

	_, err := bot.SendMediaGroup(telegoutil.MediaGroup(
		telegoutil.ID(message.Chat.ID),
		mediaItems...,
	))
	if err != nil {
		fmt.Println(err)
	}
}

func AccountInfo(bot *telego.Bot, message telego.Message) {
	i18n := localization.Get(message.Chat)
	if len(strings.Fields(message.Text)) < 2 {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(message.Chat.ID),
			Text:      i18n("twitter_no_username"),
			ParseMode: "HTML",
		})
		return
	}

	username := strings.Fields(message.Text)[1]
	accountInfo := twitter.UserInfo(username)

	if accountInfo == nil {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(message.Chat.ID),
			Text:      i18n("twitter_invalid_username"),
			ParseMode: "HTML",
		})
		return
	}

	text := fmt.Sprintf(
		"%s%s%s%s%s%s%s",
		i18n("twitter_profile_info"),
		i18n("twitter_profile_name"),
		i18n("twitter_profile_verified"),
		i18n("twitter_profile_bio"),
		i18n("twitter_profile_followers"),
		i18n("twitter_profile_following"),
		i18n("twitter_profile_tweets"),
	)
	text = fmt.Sprintf(
		text,
		username,
		username,
		accountInfo.Name,
		accountInfo.Verified,
		accountInfo.Description,
		accountInfo.FollowersCount,
		accountInfo.FriendsCount,
		accountInfo.StatusesCount,
	)

	bot.SendMessage(&telego.SendMessageParams{
		ChatID:    telegoutil.ID(message.Chat.ID),
		Text:      text,
		ParseMode: "HTML",
	})
}
