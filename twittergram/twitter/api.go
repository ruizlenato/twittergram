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

var headers = map[string]string{
		"Authorization":             "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA",
		"x-twitter-client-language": "en",
		"x-twitter-active-user":     "yes",
		"Accept-language":           "en",
		"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
	"content-type":              "application/json",
}

const (
	tweetDetail      = "https://twitter.com/i/api/graphql/5GOHgZe-8U2j5sVHQzEm9A/TweetResultByRestId"
	userByScreenName = "https://api.twitter.com/graphql/qW5u-DAuXpMEG0zA1F7UGQ/UserByScreenName"
	tweetIDRegex     = `(?:http(?:s)?://)?(?:www.|mobile.)?(?:twitter|x).com/.*?/([0-9]+)`
)

type requestParams struct {
	Method     string            // "GET", "OPTIONS" or "POST"
	Headers    map[string]string // Common headers for both GET and POST
	Query      map[string]string // Query parameters for GET
	BodyString []string          // Body of the request for POST
}

// Request sends a GET, OPTIONS or POST request to the specified link with the provided parameters and returns the response.
// The Link specifies the URL to send the request to.
// The params contain additional parameters for the request, such as headers, query parameters, and body.
// The Method field in params should be "GET" or "POST" to indicate the type of request.
//
// Example usage:
//
//	response := Request("https://api.example.com/users", RequestParams{
//		Method: "GET",
//		Headers: map[string]string{
//			"Authorization": "Bearer your-token",
//		},
//		Query: map[string]string{
//			"page":  "1",
//			"limit": "10",
//		},
//	})
//
//	response := Request("https://example.com/api", RequestParams{
//		Method: "POST",
//		Headers: map[string]string{
//			"Content-Type": "application/json",
//		},
//		BodyString: []string{
//			"param1=value1",
//			"param2=value2",
//		},
//	})
func request(Link string, params requestParams) *fasthttp.Response {
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()

	client := &fasthttp.Client{
		ReadBufferSize:  16 * 1024,
		MaxConnsPerHost: 1024,
	}

	request.Header.SetMethod(params.Method)
	for key, value := range params.Headers {
		request.Header.Set(key, value)
	}

	if params.Method == fasthttp.MethodGet {
		request.SetRequestURI(Link)
		for key, value := range params.Query {
			request.URI().QueryArgs().Add(key, value)
		}
	} else if params.Method == fasthttp.MethodOptions {
		request.SetRequestURI(Link)
		for key, value := range params.Query {
			request.URI().QueryArgs().Add(key, value)
		}
	} else if params.Method == fasthttp.MethodPost {
		request.SetBodyString(strings.Join(params.BodyString, "&"))
	request.SetRequestURI(Link)
	} else {
		log.Print("[request/Request] Error: Unsupported method ", params.Method)
		return response
	}

	err := client.Do(request, response)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			return response
		}
		log.Print("[request/Request] Error: ", err)
	}

	fmt.Println(string(request.RequestURI()))

	if params.Method == fasthttp.MethodGet {
		fmt.Println(response)
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
		"hidden_profile_likes_enabled":                                      true,
		"hidden_profile_subscriptions_enabled":                              true,
		"rweb_tipjar_consumption_enabled":                                   true,
		"responsive_web_graphql_exclude_directive_enabled":                  true,
		"verified_phone_label_enabled":                                      false,
		"subscriptions_verification_info_is_identity_verified_enabled":      true,
		"subscriptions_verification_info_verified_since_enabled":            true,
		"highlights_tweets_tab_ui_enabled":                                  true,
		"responsive_web_twitter_article_notes_tab_enabled":                  true,
		"creator_subscriptions_tweet_preview_api_enabled":                   true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled": false,
		"responsive_web_graphql_timeline_navigation_enabled":                true,
	}

	jsonMarshal := func(data interface{}) []byte {
		result, _ := json.Marshal(data)
		return result
	}

	body := request(userByScreenName, requestParams{
		Method: "GET",
		Query: map[string]string{
			"variables": string(jsonMarshal(variables)),
			"features":  string(jsonMarshal(features)),
		},
		Headers: headers,
	}).Body()

	var twitterAPIData *TwitterAPIData
	err := json.Unmarshal(body, &twitterAPIData)
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
	headers["x-guest-token"] = getGuestToken()
	headers["cookie"] = fmt.Sprintf("guest_id=v1:%v;", getGuestToken())

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

	}

	body := request("https://twitter.com/i/api/graphql/5GOHgZe-8U2j5sVHQzEm9A/TweetResultByRestId", requestParams{
		Method: "GET",
		Query: map[string]string{
			"variables": string(jsonMarshal(variables)),
			"features":  string(jsonMarshal(features)),
		},
		Headers: headers,
	}).Body()

	if body == nil {
		return TweetContent{}
	}

	var twitterAPIData *TwitterAPIData
	err := json.Unmarshal(body, &twitterAPIData)
	if err != nil {
		log.Println(err)
	}

	if twitterAPIData == nil || twitterAPIData.Data.TweetResults.Legacy == nil {
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
