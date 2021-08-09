import os
import tweepy
import youtube_dl

from twittergram import apitweepy
from pyrogram.types import Message
from pyrogram import Client, filters

TWITTER_LINKS = r"(http(s)?:\/\/(?:www\.)?(?:v\.)?(?:twitter.com)\/(?:.*?))(?:\s|$)"


@Client.on_message(filters.regex(TWITTER_LINKS))
async def ytdl(c: Client, m: Message):
    url = m.matches[0].group(0)
    filename = "/tmp/%s%s.mp4" % (m.chat.id, m.message_id)
    ydl_opts = {"outtmpl": filename}
    with youtube_dl.YoutubeDL(ydl_opts) as ydl:
        ydl.download([url])

    with open(filename, "rb") as video:
        await m.reply_video(video=video)
    os.remove(filename)


@Client.on_message(filters.command("userinfo"))
async def uinfo(c: Client, m: Message):
    try:
        if m.reply_to_message and m.reply_to_message.text:
            username = m.reply_to_message.text
        elif m.text and m.text.split(maxsplit=1)[1]:
            username = m.text.split(maxsplit=1)[1]
    except IndexError:
        await m.reply_text("Você esqueceu do Username")
        return

    if username:
        try:
            user = apitweepy.get_user(username)
            rep = f"<b>Informações da conta <a href='https://twitter.com/{username}'>@{username}</a> (Twitter):</b>\n"
            rep += f"\n<b>Nome:</b> <code>{user.name}</code>"
            rep += f"\n<b>Bio:</b> <code>{user.description}</code>"
            rep += f"\n<b>Seguidores:</b> <code>{user.followers_count}</code>"
            rep += f"\n<b>Seguindo:</b> <code>{user.friends_count}</code>"
            rep += f"\n<b>Número de tweets:</b> <code>{user.statuses_count}</code>"
            await m.reply_text(rep, disable_web_page_preview=True)
        except tweepy.TweepError as error:
            rep = "Username errado"
            await m.reply_text(rep)
    else:
        rep = "Você esquceu do username"
        await m.reply_text(rep)
    return
