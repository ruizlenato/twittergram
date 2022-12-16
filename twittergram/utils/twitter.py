# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import re
import os
import json
import httpx
import contextlib

from ..config import BARRER_TOKEN

# httpx Things
http = httpx.AsyncClient(http2=True, timeout=30.0, follow_redirects=True)


class TwitterAPI:
    def __init__(self):
        self.TwitterAPI: str = "https://api.twitter.com/2/"

    async def user(self, username: str):
        params: str = f"?usernames={username}&user.fields=name,protected,url,username,verified,public_metrics,description"
        res = await http.get(
            f"{self.TwitterAPI}users/by{params}",
            headers={"Authorization": f"Bearer {BARRER_TOKEN}"},
        )

        try:
            return json.loads(res.content)["data"][0]
        except KeyError:
            return None

    async def download(self, url: str, id: int):
        x = re.match(".*twitter.com/.+status/([A-Za-z0-9]+)", url)
        params: str = "?expansions=attachments.media_keys,author_id&media.fields=type,variants,url,height,width&tweet.fields=entities"
        res = await http.get(
            f"{self.TwitterAPI}tweets/{x[1]}{params}",
            headers={"Authorization": f"Bearer {BARRER_TOKEN}"},
        )
        tweet = json.loads(res.content)

        caption = (
            f"<b>{tweet['includes']['users'][0]['name']}</b>\n{tweet['data']['text']}"
        )
        self.files: list = []

        try:
            tweet_medias = tweet["includes"]["media"]
        except KeyError:
            return

        for media in tweet_medias:
            if media["type"] in ("animated_gif", "video"):
                bitrate = [
                    a["bit_rate"]
                    for a in media["variants"]
                    if a["content_type"] == "video/mp4"
                ]
                key = media["media_key"]
                for a in media["variants"]:
                    with contextlib.suppress(FileExistsError):
                        os.mkdir(f"./downloads/{id}/")
                    if a["content_type"] == "video/mp4" and a["bit_rate"] == max(
                        bitrate
                    ):
                        path = f"./downloads/{id}/{key}.mp4"
                        with open(path, "wb") as f:
                            f.write((await http.get(a["url"])).content)
                        self.files.append(
                            {
                                "path": path,
                                "width": media["width"],
                                "height": media["height"],
                            }
                        )
            else:
                self.files.append(
                    {
                        "path": media["url"],
                        "width": media["width"],
                        "height": media["height"],
                    }
                )
        return self.files, caption
