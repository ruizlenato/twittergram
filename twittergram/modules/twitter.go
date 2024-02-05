package modules

import (
	"fmt"
	"strings"
	"twittergram/twittergram/localization"
	"twittergram/twittergram/twitter"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
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
		if media.Video {
			mediaItems = append(mediaItems, telegoutil.MediaVideo(telegoutil.FileFromURL(media.Source)).WithWidth(media.Width).WithHeight(media.Height))
		} else {
			mediaItems = append(mediaItems, telegoutil.MediaPhoto(telegoutil.FileFromURL(media.Source)))
		}
	}

	if len(mediaItems) < 2 && mediaItems[0].MediaType() == "photo" && !message.LinkPreviewOptions.IsDisabled {
		return
	}

	if len(mediaItems) > 0 {
		for _, media := range tweetMedias.Medias[:1] {
			if mediaItems[0].MediaType() == "photo" {
				mediaItems[0] = telegoutil.MediaPhoto(telegoutil.FileFromURL(media.Source)).WithCaption(tweetMedias.Caption)
			} else {
				mediaItems[0] = telegoutil.MediaVideo(telegoutil.FileFromURL(media.Source)).WithWidth(media.Width).WithHeight(media.Height).WithCaption(tweetMedias.Caption)
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
