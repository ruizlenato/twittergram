# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import contextlib
import shutil
import pyrogram

from ..bot import Client
from ..locales import strings
from ..utils import TwitterAPI

from pyrogram import filters
from pyrogram.types import Message, InputMediaPhoto, InputMediaVideo
from pyrogram.raw.types import InputMessageID
from pyrogram.enums import ChatAction, ChatType
from pyrogram.raw.functions import channels, messages
from pyrogram.errors.exceptions import ChannelInvalid

TWITTER_LINKS = r"(https?://(?:www.|mobile.)?twitter.com/.*?/.*?/([0-9]+))"


@Client.on_message(filters.command(["twitter", "uinfo"]))
async def uinfo(c: Client, m: Message):
    if len(m.command) > 1:
        username = m.text.split(None, 1)[1]
    else:
        return await m.reply_text(await strings(m, "Twitter.no_username"))

    user = await TwitterAPI().user(username)
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
    url = m.matches[0].group(0)
    path = f"{m.id}{m.chat.id}"
    files, caption = await TwitterAPI().download(url, path)

    if m.chat.type == ChatType.PRIVATE:
        method = messages.GetMessages(id=[InputMessageID(id=(m.id))])
    else:
        method = channels.GetMessages(
            channel=await c.resolve_peer(m.chat.id), id=[InputMessageID(id=(m.id))]
        )
    try:
        rawM = (await c.invoke(method)).messages[0].media
    except ChannelInvalid:
        return

    medias = []

    for media in files:
        if media["path"][-3:] == "mp4" and len(files) == 1:
            await c.send_chat_action(m.chat.id, ChatAction.UPLOAD_VIDEO)
            await m.reply_video(
                video=media["path"],
                width=media["width"],
                height=media["height"],
                caption=caption,
            )
            return shutil.rmtree(f"./downloads/{path}/", ignore_errors=True)
        mType = InputMediaVideo if media["path"][-3:] == "mp4" else InputMediaPhoto
        if not medias:
            medias.append(
                mType(
                    media["path"],
                    width=media["width"],
                    height=media["height"],
                    caption=caption,
                )
            )
        else:
            medias.append(
                mType(media["path"], width=media["width"], height=media["height"])
            )

    if medias:
        if rawM and len(medias) == 1 and "InputMediaPhoto" in str(medias[0]):
            return

        await c.send_chat_action(m.chat.id, ChatAction.UPLOAD_DOCUMENT)
        await m.reply_media_group(media=medias)
    return shutil.rmtree(f"./downloads/{path}/", ignore_errors=True)
