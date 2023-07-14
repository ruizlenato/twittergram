# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import os

from babel import Locale
from pyrogram.enums import ChatType
from pyrogram.types import Message

from ..bot import Client
from ..database.chats import add_chat, get_chat

Languages: list[str] = []  # Loaded Locales

for file in os.listdir("twittergram/locales"):
    if file not in ("__init__.py", "__pycache__"):
        Languages.append(file.replace(".yaml", ""))


# This is the first plugin run to guarantee
# that the actual chat is initialized in the DB.
@Client.on_message(group=-1)
async def check_chat(client: Client, message: Message):
    chat = message.chat
    user = message.from_user

    try:
        language_code = str(Locale.parse(user.language_code, sep="-"))
    except (AttributeError, TypeError):
        language_code: str = "en_US"

    if language_code not in Languages:
        language_code: str = "en-us"

    if user and await get_chat(user.id, ChatType.PRIVATE) is None:
        await add_chat(user.id, language_code, ChatType.PRIVATE)

    if await get_chat(chat.id, chat.type) is None:
        await add_chat(chat.id, language_code, chat.type)
