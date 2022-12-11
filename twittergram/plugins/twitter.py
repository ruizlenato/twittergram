# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import re
import shutil
import pyrogram

from ..bot import Client
from ..locales import strings
from ..utils import TwitterAPI

from pyrogram import filters
from pyrogram.types import Message
from pyrogram.enums import ChatAction

TWITTER_LINKS = r"(https?://(?:www.|mobile.)?twitter.com/.*?/.*?/([0-9]+))"


@Client.on_message(filters.command(["twitter", "uinfo"]))
async def uinfo(c: Client, m: Message):
    if len(m.command) > 1:
        username = m.text.split(None, 1)[1]
    else:
        return await m.reply_text(await strings(m, "Twitter.no_username"))

    user = await TwitterAPI.user(username)
    if user is None:
        return await m.reply_text(await strings(m, "Twitter.wrong_username"))

    try:
        rep = (await strings(m, "Twitter.acc_info")).format(username, username)
        rep += (await strings(m, "Twitter.name")).format(user["name"])
        rep += (await strings(m, "Twitter.verified")).format(user["verified"])
        rep += (await strings(m, "Twitter.bio")).format(user["description"])
        rep += (await strings(m, "Twitter.followers")).format(
            user["public_metrics"]["followers_count"]
        )
        rep += (await strings(m, "Twitter.following")).format(
            user["public_metrics"]["following_count"]
        )
        rep += (await strings(m, "Twitter.tweets")).format(
            user["public_metrics"]["tweet_count"]
        )
        await m.reply_text(rep, disable_web_page_preview=True)
    except AttributeError as error:
        print(error)
        return await m.reply_text(await strings(m, "Twitter.wrong_username"))


@Client.on_message(filters.regex(TWITTER_LINKS))
async def Twitter(c: Client, m: Message):
    rawM = await c.invoke(
        pyrogram.raw.functions.messages.GetMessages(
            id=[pyrogram.raw.types.InputMessageID(id=(m.id))]
        )
    )

    url = m.matches[0].group(0)
    files = await TwitterAPI.download(url, f"{m.id}{m.chat.id}")

    if files:
        if (
            rawM.messages[0].media
            and len(files) == 1
            and re.search(r"InputMediaPhoto", str(files[0]), re.M)
        ):
            return

        await c.send_chat_action(m.chat.id, ChatAction.UPLOAD_DOCUMENT)
        await m.reply_media_group(media=files)
    return shutil.rmtree(f"./downloads/{id}/", ignore_errors=True)
