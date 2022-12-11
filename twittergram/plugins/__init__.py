# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
from pyrogram.types import Message
from pyrogram.enums import ChatType

from ..bot import Client
from ..locales import locales_name
from ..database.chats import add_chat, get_chat

# This is the first plugin run to guarantee
# that the actual chat is initialized in the DB.
@Client.on_message(group=-1)
async def check_chat(c: Client, m: Message):
    chat = m.chat
    user = m.from_user

    try:
        language_code = user.language_code
    except AttributeError:
        language_code: str = "en-us"

    if language_code not in locales_name:
        language_code: str = "en-us"

    if user and await get_chat(user.id, ChatType.PRIVATE) is None:
        await add_chat(user.id, language_code, ChatType.PRIVATE)

    if await get_chat(chat.id, chat.type) is None:
        await add_chat(chat.id, language_code, chat.type)
