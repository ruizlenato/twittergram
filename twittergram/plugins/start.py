# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import asyncio
import contextlib

from babel import Locale
from pyrogram import filters
from pyrogram.enums import ChatMemberStatus, ChatType
from pyrogram.errors import MessageNotModified, UserNotParticipant
from pyrogram.helpers import array_chunk, ikb
from pyrogram.types import CallbackQuery, Message

from ..bot import Client
from ..database.locales import set_db_lang
from ..locales import locale
from ..plugins import Languages


@Client.on_message(filters.command("start"))
@Client.on_callback_query(filters.regex(r"start"))
@locale("start")
async def start_command(client: Client, message: Message | CallbackQuery, strings):
    if isinstance(message, CallbackQuery):
        chat_type = message.message.chat.type
        reply_text = message.edit_message_text
    else:
        chat_type = message.chat.type
        reply_text = message.reply_text

    if chat_type == ChatType.PRIVATE:
        keyboard = [
            [
                (strings["language_button"], "setchatlang"),
                (strings["help_button"], "help_button"),
            ]
        ]

        text = strings["private_start_message"].format(message.from_user.first_name)
    else:
        keyboard = [[("Start", f"https://t.me/{client.me.username}?start=start", "url")]]
        text = strings["start_message"]

    await reply_text(text, reply_markup=ikb(keyboard), disable_web_page_preview=True)


@Client.on_callback_query(filters.regex("^lang_set (?P<code>.+)"))
@locale("start")
async def set_lang(client: Client, message: Message, strings):
    lang = message.matches[0]["code"]
    if message.message.chat.type is not ChatType.PRIVATE:
        try:
            member = await client.get_chat_member(
                chat_id=message.message.chat.id, user_id=message.from_user.id
            )
            if member.status not in (
                ChatMemberStatus.ADMINISTRATOR,
                ChatMemberStatus.OWNER,
            ):
                return
        except UserNotParticipant:
            return

    keyboard = [[(strings["back_button"], "start")]]
    if message.message.chat.type == ChatType.PRIVATE:
        await set_db_lang(message.from_user.id, lang, message.message.chat.type)
    elif message.message.chat.type in (ChatType.GROUP, ChatType.SUPERGROUP):
        await set_db_lang(message.message.chat.id, lang, message.message.chat.type)
    text = strings["language_saved"]
    with contextlib.suppress(MessageNotModified):
        await message.edit_message_text(text, reply_markup=ikb(keyboard))


@Client.on_callback_query(filters.regex(r"setchatlang"))
@locale("start")
async def setlang(client: Client, message: CallbackQuery, strings):
    buttons: list = []
    for lang in list(Languages):
        text, data = (Locale.parse(lang).display_name.title(), f"lang_set {lang}")
        buttons.append((text, data))
    keyboard = array_chunk(buttons, 2)

    if message.message.chat.type == ChatType.PRIVATE:
        keyboard += [[(strings["back_button"], "start")]]
    else:
        try:
            member = await client.get_chat_member(
                chat_id=message.message.chat.id, user_id=message.from_user.id
            )
            if member.status not in (
                ChatMemberStatus.ADMINISTRATOR,
                ChatMemberStatus.OWNER,
            ):
                if isinstance(message, CallbackQuery):
                    await message.answer(
                        strings["not_admin"],
                        show_alert=True,
                        cache_time=60,
                    )
                else:
                    message = await message.edit_message_text(strings["not_admin"])
                    await asyncio.sleep(5.0)
                    await message.delete()
                return None
        except AttributeError:
            return None
        except UserNotParticipant:
            return None
    return await message.edit_message_text(strings["select_lang"], reply_markup=ikb(keyboard))


@Client.on_callback_query(filters.regex(r"help_button"))
@locale("start")
async def help(client: Client, cq: CallbackQuery, strings):
    keyboard = [[(strings["back_button"], "start_command")]]
    await cq.edit_message_text(strings["help_msg"], reply_markup=ikb(keyboard))
