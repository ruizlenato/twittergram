from pyrogram import Client, filters
from pyrogram.types import Message, InlineKeyboardButton, InlineKeyboardMarkup

START_MESSAGE = "🇧🇷 Oi! Eu sou um bot para baixar vídeos do twitter<strong> penas envie o link do vídeo e a mágica irá acontecer</strong>\n\n\n🇺🇸 Hi! I'm a bot to download videos from twitter, <strong>just send the video link and the magic will happen</strong>"
START_MESSAGE_PRIVATE = "🇧🇷 Clique no botão abaixo e veja minhas funções\n\n🇺🇸 Click on the button below and see my functions"


@Client.on_message(filters.command("start", prefixes="/"))
async def start_command(c: Client, m: Message):
    me = await c.get_me()
    if m.chat.type == "private":
        await m.reply_text(START_MESSAGE)
    else:
        keyboard = InlineKeyboardMarkup(
            inline_keyboard=[
                [
                    InlineKeyboardButton(
                        text="Start", url=f"https://t.me/{me.username}?start=start"
                    )
                ]
            ]
        )
        await m.reply_text(START_MESSAGE_PRIVATE, reply_markup=keyboard)
