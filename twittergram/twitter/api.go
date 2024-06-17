package twitter

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sort"
	"strings"
)

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

	query := map[string]string{
		"variables": marshalJSON(variables),
		"features":  marshalJSON(features),
	}

	body := request(userByScreenName, requestParams{
		Method:  "GET",
		Query:   query,
		Headers: headers,
	}).Body()

	var twitterAPIData TwitterAPIData
	if err := json.Unmarshal(body, &twitterAPIData); err != nil {
		log.Println(err)
		return nil
	}

	if twitterAPIData.Data.User == nil {
		return nil
	}

	return twitterAPIData.Data.User.Result.Legacy
}

func TweetMedias(url string) TweetContent {
	tweetID, err := extractTweetID(url)
	if err != nil {
		log.Println(err)
		return TweetContent{}
	}

	guestToken := getGuestToken()
	headers["x-guest-token"] = guestToken
	headers["cookie"] = fmt.Sprintf("guest_id=v1:%v;", guestToken)

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

	query := map[string]string{
		"variables": marshalJSON(variables),
		"features":  marshalJSON(features),
	}

	body := request(tweetDetail, requestParams{
		Method:  "GET",
		Query:   query,
		Headers: headers,
	}).Body()

	var twitterAPIData *TwitterAPIData
	if err := json.Unmarshal(body, &twitterAPIData); err != nil {
		log.Printf("Error unmarshalling Twitter data: %v", err)
		return TweetContent{}
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

	if tweet := (*twitterAPIData).Data.TweetResults.Result.Legacy; tweet != nil {
		caption = fmt.Sprintf("<b>%s (<code>%s</code>)</b>:\n",
			(*twitterAPIData).Data.TweetResults.Result.Core.UserResults.Result.Legacy.Name,
			(*twitterAPIData).Data.TweetResults.Core.UserResults.Result.Legacy.ScreenName)

		if idx := strings.LastIndex(tweet.FullText, " https://t.co/"); idx != -1 {
			caption += tweet.FullText[:idx]
		}
	}

	medias := TweetContent{
		Medias:  tweetContent.Medias,
		Caption: caption,
	}

	return medias
}
