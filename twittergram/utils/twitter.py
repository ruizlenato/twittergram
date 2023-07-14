# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import io
import json
import re
import uuid

import filetype
import httpx

# httpx Things
http = httpx.AsyncClient(http2=True, timeout=30.0, follow_redirects=True)


class TwitterAPI:
    def __init__(self):
        self.TwitterBarrer: str = "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH\
5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA"
        self.files: list = []
        csrfToken = str(uuid.uuid4()).replace("-", "")
        self.headers = {
            "Authorization": self.TwitterBarrer,
            "Cookie": f"auth_token=ee4ebd1070835b90a9b8016d1e6c6130ccc89637; \
ct0={csrfToken}; ",
            "x-twitter-active-user": "yes",
            "x-twitter-auth-type": "OAuth2Session",
            "x-twitter-client-language": "en",
            "x-csrf-token": csrfToken,
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) \
Gecko/20100101 Firefox/116.0",
        }

    async def downloader(self, url: str, width: int, height: int):
        """
        Get the media from URL.

        Arguments:
            url (str): The URL of the post and res info.

        Returns:
            Dict: Media url and res info.
        """
        file = io.BytesIO((await http.get(url)).content)
        file.name = f"{url[60:80]}.{filetype.guess_extension(file)}"
        self.files.append({"media": file, "width": width, "height": height})

    async def user(self, username: str):
        params = {
            "variables": json.dumps(
                {
                    "screen_name": username,
                    "withSafetyModeUserFields": True,
                    "withSuperFollowsUserFields": True,
                }
            )
        }
        res = await http.get(
            "https://twitter.com/i/api/graphql/cYsDlVss-qimNYmNlb6inw/UserByScreenName",
            params=params,
            headers=self.headers,
        )
        return res.json()

    async def download(self, url: str, id: int):
        # Extract the tweet ID from the URL
        x = re.match(".*twitter.com/.+status/([A-Za-z0-9]+)", url)
        params = {
            "variables": json.dumps(
                {
                    "focalTweetId": x[1],
                    "referrer": "messages",
                    "includePromotedContent": True,
                    "withCommunity": True,
                    "withQuickPromoteEligibilityTweetFields": True,
                    "withBirdwatchNotes": True,
                    "withVoice": False,
                    "withV2Timeline": True,
                }
            ),
            "features": json.dumps(
                {
                    "rweb_lists_timeline_redesign_enabled": True,
                    "responsive_web_graphql_exclude_directive_enabled": True,
                    "verified_phone_label_enabled": False,
                    "creator_subscriptions_tweet_preview_api_enabled": True,
                    "responsive_web_graphql_timeline_navigation_enabled": True,
                    "responsive_web_graphql_skip_user_profile_image_\
extensions_enabled": False,
                    "tweetypie_unmention_optimization_enabled": True,
                    "responsive_web_edit_tweet_api_enabled": True,
                    "graphql_is_translatable_rweb_tweet_is_translatable_enabled": False,
                    "view_counts_everywhere_api_enabled": True,
                    "longform_notetweets_consumption_enabled": True,
                    "responsive_web_twitter_article_tweet_consumption_enabled": False,
                    "tweet_awards_web_tipping_enabled": False,
                    "freedom_of_speech_not_reach_fetch_enabled": True,
                    "standardized_nudges_misinfo": True,
                    "tweet_with_visibility_results_prefer_gql\
_limited_actions_policy_enabled": True,
                    "longform_notetweets_rich_text_read_enabled": True,
                    "longform_notetweets_inline_media_enabled": True,
                    "responsive_web_media_download_video_enabled": False,
                    "responsive_web_enhance_cards_enabled": False,
                }
            ),
            "fieldToggles": json.dumps(
                {"withAuxiliaryUserLabels": False, "withArticleRichContentState": False}
            ),
        }

        res = (
            await http.get(
                "https://twitter.com/i/api/graphql/NmCeCgkVlsRGS1cAwqtgmw/TweetDetail",
                params=params,
                headers=self.headers,
            )
        ).json()

        self.files: list = []
        try:
            tweet = res["data"]["threaded_conversation_with_injections_v2"]["instructions"][0][
                "entries"
            ][0]["content"]["itemContent"]["tweet_results"]["result"]
        except KeyError:
            return None

        user_name = tweet["core"]["user_results"]["result"]["legacy"]["name"]
        caption = f"<b>{user_name}</b>\n{tweet['legacy']['full_text']}"

        for media in tweet["legacy"]["extended_entities"]["media"]:
            if media["type"] in ("animated_gif", "video"):
                bitrate = [
                    a["bitrate"]
                    for a in media["video_info"]["variants"]
                    if a["content_type"] == "video/mp4"
                ]
                for a in media["video_info"]["variants"]:
                    if a["content_type"] == "video/mp4" and a["bitrate"] == max(bitrate):
                        url = a["url"]

                await self.downloader(
                    url,
                    media["original_info"]["width"],
                    media["original_info"]["height"],
                )
            else:
                await self.downloader(
                    media["media_url_https"],
                    media["original_info"]["width"],
                    media["original_info"]["height"],
                )
        return self.files, caption
