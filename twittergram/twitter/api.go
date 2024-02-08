package twitter

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

const (
	tweetDetail      = "https://twitter.com/i/api/graphql/NmCeCgkVlsRGS1cAwqtgmw/TweetDetail"
	userByScreenName = "https://twitter.com/i/api/graphql/cYsDlVss-qimNYmNlb6inw/UserByScreenName"
	tweetIDRegex     = `.*(?:twitter|x).com/.+status/([A-Za-z0-9]+)`
)

func callTwitterAPI(link string, query map[string]string) *fasthttp.Response {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	client := &fasthttp.Client{ReadBufferSize: 8192}

	req.Header.SetMethod("GET")
	csrfToken := strings.ReplaceAll((uuid.New()).String(), "-", "")
	headers := map[string]string{
		"Authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"Cookie":                    fmt.Sprintf("auth_token=ee4ebd1070835b90a9b8016d1e6c6130ccc89637; ct0=%v; ", csrfToken),
		"x-twitter-active-user":     "yes",
		"x-twitter-auth-type":       "OAuth2Session",
		"x-twitter-client-language": "en",
		"x-csrf-token":              csrfToken,
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	req.SetRequestURI(link)
	for key, value := range query {
		req.URI().QueryArgs().Add(key, value)
	}

	err := client.Do(req, res)
	if err != nil {
		log.Fatal(err)
	}

	return res
}

func UserInfo(username string) Legacy {
	variables := map[string]interface{}{
		"screen_name":                username,
		"withSafetyModeUserFields":   true,
		"withSuperFollowsUserFields": true,
	}

	variablesJson, err := json.Marshal(variables)
	if err != nil {
		log.Print(err)
	}

	query := map[string]string{"variables": string(variablesJson)}
	body := callTwitterAPI(userByScreenName, query).Body()

	var twitterAPIData *TwitterAPIData
	err = json.Unmarshal(body, &twitterAPIData)
	if err != nil {
		log.Println(err)
	}

	if twitterAPIData.Data.User == nil {
		return nil
	}

	return twitterAPIData.Data.User.Result.Legacy
}

func TweetMedias(url string) TweetContent {
	tweetID := (regexp.MustCompile((tweetIDRegex))).FindStringSubmatch(url)[1]
	variables := map[string]interface{}{
		"focalTweetId":                           tweetID,
		"referrer":                               "messages",
		"includePromotedContent":                 true,
		"withCommunity":                          true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withBirdwatchNotes":                     true,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	}

	features := map[string]interface{}{
		"rweb_lists_timeline_redesign_enabled":                                    true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              false,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_enhance_cards_enabled":                                    false,
	}
	fieldtoggles := map[string]interface{}{
		"withAuxiliaryUserLabels":     false,
		"withArticleRichContentState": false,
	}

	featuresJson, _ := json.Marshal(features)
	fieldTogglesJson, _ := json.Marshal(fieldtoggles)
	variablesJson, err := json.Marshal(variables)

	if err != nil {
		log.Print(err)
	}

	query := map[string]string{
		"variables":    string(variablesJson),
		"features":     string(featuresJson),
		"fieldToggles": string(fieldTogglesJson),
	}

	body := callTwitterAPI(tweetDetail, query).Body()
	var twitterAPIData *TwitterAPIData
	err = json.Unmarshal(body, &twitterAPIData)
	if err != nil {
		log.Println(err)
	}

	var tweetResult interface{}
	for _, entry := range twitterAPIData.Data.ThreadedConversationWithInjectionsV2.Instructions[0].Entries {
		if entry.EntryID == fmt.Sprintf("tweet-%v", tweetID) {
			if entry.Content.ItemContent.TweetResults.Result.Typename == "TweetWithVisibilityResults" {
				tweetResult = entry.Content.ItemContent.TweetResults.Result.Tweet.Legacy
			} else {
				tweetResult = entry.Content.ItemContent.TweetResults.Result.Legacy
			}
			break
		}
	}
	tweetContent := &TweetContent{}

	if tweetResult.(Legacy) == nil {
		return TweetContent{}
	}

	for _, media := range tweetResult.(Legacy).ExtendedEntities.Media {
		var videoType string
		if slices.Contains([]string{"animated_gif", "video"}, media.Type) {
			videoType = "video"
		}
		if videoType != "video" {
			tweetContent.Medias = append(tweetContent.Medias, Medias{
				Height: media.OriginalInfo.Height,
				Width:  media.OriginalInfo.Width,
				Source: media.MediaURLHTTPS,
				Video:  false,
			})
		} else {
			tweetContent.Medias = append(tweetContent.Medias, Medias{
				Height: media.OriginalInfo.Height,
				Width:  media.OriginalInfo.Width,
				Source: media.VideoInfo.Variants[0].URL,
				Video:  true,
			})
		}
	}
	var caption string
	if tweet, ok := tweetResult.(Legacy); ok {
		caption = tweet.FullText
	}

	medias := TweetContent{
		Medias:  tweetContent.Medias,
		Caption: caption}

	return medias
}
