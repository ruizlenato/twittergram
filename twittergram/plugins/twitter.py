# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)

import filetype
from pyrogram import filters
from pyrogram.enums import ChatAction, ChatType
from pyrogram.errors.exceptions import ChannelInvalid
from pyrogram.raw.functions import channels, messages
from pyrogram.raw.types import InputMessageID
from pyrogram.types import InputMediaPhoto, InputMediaVideo, Message

from ..bot import Client
from ..locales import locale
from ..utils import TwitterAPI

TWITTER_LINKS = r"(https?://(?:www.|mobile.)?(twitter|x).com/.*?/.*?/([0-9]+))"


@Client.on_message(filters.command(["twitter", "uinfo"]))
@locale("twitter")
async def uinfo(client: Client, message: Message, strings):
    if len(message.command) > 1:
        username = message.text.split(None, 1)[1]
    else:
        return await message.reply_text(strings["no_username"])

    try:
        user = (await TwitterAPI().user(username))["data"]["user"]["result"]["legacy"]
    except KeyError:
        return await message.reply_text(strings["wrong_username"])

    try:
        rep = strings["account_info"].format(username, username)
        rep += strings["account_name"].format(user["name"])
        rep += strings["account_verified"].format(user["verified"])
        rep += strings["account_bio"].format(user["description"])
        rep += strings["account_followers"].format(user["followers_count"])
        rep += strings["account_following"].format(user["friends_count"])
        rep += strings["account_tweets"].format(user["statuses_count"])
        await message.reply_text(rep, disable_web_page_preview=True)
    except AttributeError as error:
        print(error)
        return await message.reply_text(strings["wrong_username"])


@Client.on_message(filters.regex(TWITTER_LINKS))
async def Twitter(client: Client, message: Message):
    url = message.matches[0].group(0)
    path = f"{message.id}{message.chat.id}"
    files, caption = await TwitterAPI().download(url, path)

    if message.chat.type == ChatType.PRIVATE:
        method = messages.GetMessages(id=[InputMessageID(id=(message.id))])
    else:
        method = channels.GetMessages(
            channel=await client.resolve_peer(message.chat.id), id=[InputMessageID(id=(message.id))]
        )
    try:
        rawM = (await client.invoke(method)).messages[0].media
    except ChannelInvalid:
        return None

    medias = []
    for media in files:
        if filetype.is_video(media["media"]) and len(files) == 1:
            await client.send_chat_action(message.chat.id, ChatAction.UPLOAD_VIDEO)
            return await message.reply_video(
                video=media["media"],
                width=media["width"],
                height=media["height"],
                caption=caption,
            )

        if filetype.is_video(media["media"]):
            if medias:
                medias.append(
                    InputMediaVideo(media["media"], width=media["width"], height=media["height"])
                )
            else:
                medias.append(
                    InputMediaVideo(
                        media["media"],
                        width=media["width"],
                        height=media["height"],
                        caption=caption,
                    )
                )
        elif not medias:
            medias.append(InputMediaPhoto(media["media"], caption=caption))
        else:
            medias.append(InputMediaPhoto(media["media"]))

    if medias:
        if rawM and len(medias) == 1 and "InputMediaPhoto" in str(medias[0]):
            return None

        await client.send_chat_action(message.chat.id, ChatAction.UPLOAD_DOCUMENT)
        await message.reply_media_group(media=medias)
        return None
    return None
