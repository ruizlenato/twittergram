import os
import tweepy
import yt_dlp
import tempfile

from twittergram import apitweepy
from pyrogram.types import Message
from pyrogram import Client, filters

TWITTER_LINKS = (
    r"(http(s)?:\/\/(?:www\.)?(?:mobile\.)?(?:v\.)?(?:twitter.com)\/(?:.*?))(?:\s|$)"
)


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
            user = apitweepy.get_user(screen_name=username)
            rep = f"<b>Informações da conta <a href='https://twitter.com/{username}'>@{username}</a> (Twitter):</b>\n"
            rep += f"\n<b>Nome:</b> <code>{user.name}</code>"
            rep += f"\n<b>Bio:</b> <code>{user.description}</code>"
            rep += f"\n<b>Seguidores:</b> <code>{user.followers_count}</code>"
            rep += f"\n<b>Seguindo:</b> <code>{user.friends_count}</code>"
            rep += f"\n<b>Número de tweets:</b> <code>{user.statuses_count}</code>"
            await m.reply_text(rep, disable_web_page_preview=True)
        except AttributeError as error:
            rep = "Username errado"
            await m.reply_text(rep)
            print(error)
    else:
        rep = "Você esquceu do username"
        await m.reply_text(rep)
    return


@Client.on_message(filters.regex(TWITTER_LINKS))
async def ytdl(c: Client, m: Message):
    url = m.matches[0].group(0)
    with tempfile.TemporaryDirectory() as tempdir:
        path = os.path.join(tempdir, "ytdl")
    filename = f"{path}/%s%s.mp4" % (m.chat.id, m.id)
    ydl_opts = {"outtmpl": filename, "logger": MyLogger()}
    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        ydl.download([url])

    with open(filename, "rb") as video:
        await m.reply_video(video=video)
    os.remove(filename)


class MyLogger:
    def debug(self, msg):
        # For compatibility with youtube-dl, both debug and info are passed into debug
        # You can distinguish them by the prefix '[debug] '
        if not msg.startswith("[debug] "):
            self.info(msg)

    def info(self, msg):
        pass

    def warning(self, msg):
        pass

    @staticmethod
    def error(msg):
        if "There's no video" not in msg:
            print(msg)
