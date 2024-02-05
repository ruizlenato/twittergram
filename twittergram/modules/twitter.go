package modules

import (
	"fmt"
	"log"
	"strings"
	"twittergram/twittergram/localization"

	"github.com/goccy/go-json"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
)

type tweet struct {
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

func MediaDownloader(bot *telego.Bot, message telego.Message) {
	url := extractTwitterURL(message.Text)
	if url == "" {
		bot.SendMessage(telegoutil.Message(
			telegoutil.ID(message.Chat.ID),
			"No twitter url",
		))
		return
	}

	var tweet tweet
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

func AccountInfo(bot *telego.Bot, message telego.Message) {
	i18n := localization.Get(message.Chat)
	if len(strings.Split(message.Text, " ")) < 2 {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(message.Chat.ID),
			Text:      i18n("twitter_no_username"),
			ParseMode: "HTML",
		})
		return
	}
	username := strings.Split(message.Text, " ")[1]

	variables := map[string]interface{}{
		"screen_name":                username,
		"withSafetyModeUserFields":   true,
		"withSuperFollowsUserFields": true,
	}
	variablesJson, err := json.Marshal(variables)
	if err != nil {
		log.Print(err)
	}
	query := map[string]string{
		"variables": string(variablesJson),
	}
	body := TwitterAPI("https://twitter.com/i/api/graphql/cYsDlVss-qimNYmNlb6inw/UserByScreenName", query).Body()

	var twitterAPIData *TwitterAPIData
	err = json.Unmarshal(body, &twitterAPIData)
	if err != nil {
		log.Println(err)
	}

	if twitterAPIData.Data.User == nil {
		bot.SendMessage(&telego.SendMessageParams{
			ChatID:    telegoutil.ID(message.Chat.ID),
			Text:      i18n("twitter_invalid_username"),
			ParseMode: "HTML",
		})
		return
	}

	accountInfo := twitterAPIData.Data.User.Result.Legacy

	text := fmt.Sprintf(i18n("twitter_profile_info"), username, username)
	text += fmt.Sprintf(i18n("twitter_profile_name"), accountInfo.Name)
	text += fmt.Sprintf(i18n("twitter_profile_verified"), accountInfo.Verified)
	text += fmt.Sprintf(i18n("twitter_profile_bio"), accountInfo.Description)
	text += fmt.Sprintf(i18n("twitter_profile_followers"), accountInfo.FollowersCount)
	text += fmt.Sprintf(i18n("twitter_profile_following"), accountInfo.FriendsCount)
	text += fmt.Sprintf(i18n("twitter_profile_tweets"), accountInfo.StatusesCount)

	bot.SendMessage(&telego.SendMessageParams{
		ChatID:    telegoutil.ID(message.Chat.ID),
		Text:      text,
		ParseMode: "HTML",
	})

}
