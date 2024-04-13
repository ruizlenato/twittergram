package twitter

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/valyala/fasthttp"
)

func getCommonHeaders() map[string]string {
	return map[string]string{
		"Authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"x-twitter-client-language": "en",
		"x-twitter-active-user":     "yes",
		"Accept-language":           "en",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	}
}

const (
	tweetDetail      = "https://twitter.com/i/api/graphql/5GOHgZe-8U2j5sVHQzEm9A/TweetResultByRestId"
	userByScreenName = "https://api.twitter.com/graphql/qW5u-DAuXpMEG0zA1F7UGQ/UserByScreenName"
	tweetIDRegex     = `(?:http(?:s)?://)?(?:www.|mobile.)?(?:twitter|x).com/.*?/([0-9]+)`
)

func requestGET(link string, query map[string]string) *fasthttp.Response {
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	client := &fasthttp.Client{ReadBufferSize: 16 * 1024}

	request.Header.SetMethod(fasthttp.MethodGet)
	headers := mergeMaps(getCommonHeaders(),
		map[string]string{
			"content-type":  "application/json",
			"x-guest-token": getGuestToken(),
			"cookie":        fmt.Sprintf("guest_id=v1:%v;", getGuestToken()),
		})
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	request.SetRequestURI(link)
	for key, value := range query {
		request.URI().QueryArgs().Add(key, value)
	}

	err := client.Do(request, response)
	if err != nil {
		log.Fatal(err)
	}

	return response
}

func requestPOST(Link string, bodyString []string) *fasthttp.Response {
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	client := &fasthttp.Client{
		ReadBufferSize:  16 * 1024,
		MaxConnsPerHost: 1024,
	}

	request.Header.SetMethod(fasthttp.MethodPost)
	for key, value := range getCommonHeaders() {
		request.Header.Set(key, value)
	}

	request.SetBodyString(strings.Join(bodyString, "&"))
	request.SetRequestURI(Link)

	err := client.Do(request, response)
	if err != nil {
		log.Print("[request/RequestPOST] Error:", err)
	}

	return response
}

func getGuestToken() string {
	type guestToken struct {
		GuestToken string `json:"guest_token"`
	}
	body := requestPOST("https://api.twitter.com/1.1/guest/activate.json", []string{}).Body()
	var res guestToken
	err := json.Unmarshal(body, &res)
	if err != nil {
		log.Print(err)
	}

	return res.GuestToken
}

func UserInfo(username string) Legacy {
	variables := map[string]interface{}{
		"screen_name":                username,
		"withSafetyModeUserFields":   true,
		"withSuperFollowsUserFields": true,
	}
	features := map[string]interface{}{
		"hidden_profile_likes_enabled":                                      "true",
		"hidden_profile_subscriptions_enabled":                              "true",
		"rweb_tipjar_consumption_enabled":                                   "true",
		"responsive_web_graphql_exclude_directive_enabled":                  "true",
		"verified_phone_label_enabled":                                      "false",
		"subscriptions_verification_info_is_identity_verified_enabled":      "true",
		"subscriptions_verification_info_verified_since_enabled":            "true",
		"highlights_tweets_tab_ui_enabled":                                  "true",
		"responsive_web_twitter_article_notes_tab_enabled":                  "true",
		"creator_subscriptions_tweet_preview_api_enabled":                   "true",
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": "false",
		"responsive_web_graphql_timeline_navigation_enabled":                "true",
	}
	variablesJson, err := json.Marshal(variables)
	featuresJson, _ := json.Marshal(features)
	if err != nil {
		log.Print(err)
	}

	body := requestGET(userByScreenName, map[string]string{
		"variables": string(variablesJson),
		"features":  string(featuresJson),
	})
	var twitterAPIData *TwitterAPIData
	err = json.Unmarshal(body.Body(), &twitterAPIData)
	if err != nil {
		log.Println(err)
	}

	if twitterAPIData.Data.User == nil {
		return nil
	}

	return twitterAPIData.Data.User.Result.Legacy
}

func mergeMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func TweetMedias(url string) TweetContent {
	tweetID := (regexp.MustCompile((tweetIDRegex))).FindStringSubmatch(url)[1]

	variables := map[string]interface{}{
		"tweetId":                                tweetID,
		"referrer":                               "messages",
		"includePromotedContent":                 true,
		"withCommunity":                          true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withBirdwatchNotes":                     true,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	}

	features := map[string]interface{}{
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"c9s_tweet_anatomy_moderator_badge_enabled":                               true,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"responsive_web_home_pinned_timelines_enabled":                            true,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_enhance_cards_enabled":                                    false,
	}

	featuresJson, _ := json.Marshal(features)
	variablesJson, err := json.Marshal(variables)
	if err != nil {
		log.Print(err)
	}

	body := requestGET(tweetDetail, map[string]string{
		"variables": string(variablesJson),
		"features":  string(featuresJson),
	}).Body()
	var twitterAPIData *TwitterAPIData
	err = json.Unmarshal(body, &twitterAPIData)
	if err != nil {
		log.Println(err)
	}

	if twitterAPIData.Data.TweetResults == nil || twitterAPIData.Data == nil {
		return TweetContent{}
	}
	tweetContent := &TweetContent{}

	for _, media := range twitterAPIData.Data.TweetResults.Result.Legacy.ExtendedEntities.Media {
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
			sort.Slice(media.VideoInfo.Variants, func(i, j int) bool {
				return media.VideoInfo.Variants[i].Bitrate < media.VideoInfo.Variants[j].Bitrate
			})
			tweetContent.Medias = append(tweetContent.Medias, Medias{
				Height: media.OriginalInfo.Height,
				Width:  media.OriginalInfo.Width,
				Source: media.VideoInfo.Variants[len(media.VideoInfo.Variants)-1].URL,
				Video:  true,
			})
		}
	}
	var caption string

	if tweet := twitterAPIData.Data.TweetResults.Result.Legacy; tweet != nil {
		caption = tweet.FullText
	}

	medias := TweetContent{
		Medias:  tweetContent.Medias,
		Caption: caption,
	}

	return medias
}
