# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import asyncio
import contextlib

from typing import Union

from pyrogram import filters
from pyrogram.helpers import ikb
from pyrogram.types import Message, CallbackQuery
from pyrogram.enums import ChatType, ChatMemberStatus
from pyrogram.errors import MessageNotModified, UserNotParticipant

from ..bot import Client
from ..database.locales import set_db_lang
from ..locales import strings, loaded_locales


@Client.on_message(filters.command("start"))
@Client.on_callback_query(filters.regex(r"start"))
async def start_command(c: Client, m: Union[Message, CallbackQuery]):
    if isinstance(m, CallbackQuery):
        chat_type = m.message.chat.type
        reply_text = m.edit_message_text
    else:
        chat_type = m.chat.type
        reply_text = m.reply_text

    if chat_type == ChatType.PRIVATE:
        keyboard = [
            [
                (await strings(m, "Main.btn_lang"), "setchatlang"),
                (await strings(m, "Main.btn_help"), "btn_help"),
            ]
        ]

        text = (await strings(m, "Main.private_start_msg")).format(
            m.from_user.first_name
        )
    else:
        keyboard = [[("Start", f"https://t.me/{c.me.username}?start=start", "url")]]
        text = await strings(m, "Main.start_message")

    await reply_text(text, reply_markup=ikb(keyboard), disable_web_page_preview=True)


@Client.on_callback_query(filters.regex("^set_lang (?P<code>.+)"))
async def set_lang(c: Client, m: Message):
    lang = m.matches[0]["code"]
    if m.message.chat.type is not ChatType.PRIVATE:
        try:
            member = await c.get_chat_member(
                chat_id=m.message.chat.id, user_id=m.from_user.id
            )
            if member.status not in (
                ChatMemberStatus.ADMINISTRATOR,
                ChatMemberStatus.OWNER,
            ):
                return
        except UserNotParticipant:
            return

    keyboard = [[(await strings(m, "Main.back_btn"), "setchatlang")]]
    if m.message.chat.type == ChatType.PRIVATE:
        await set_db_lang(m.from_user.id, lang, m.message.chat.type)
    elif m.message.chat.type in (ChatType.GROUP, ChatType.SUPERGROUP):
        await set_db_lang(m.message.chat.id, lang, m.message.chat.type)
    text = await strings(m, "Main.lang_saved")
    with contextlib.suppress(MessageNotModified):
        await m.edit_message_text(text, reply_markup=ikb(keyboard))


@Client.on_message(filters.command(["setlang", "language"]))
@Client.on_callback_query(filters.regex(r"setchatlang"))
async def setlang(c: Client, m: Union[Message, CallbackQuery]):
    if isinstance(m, CallbackQuery):
        chat_id = m.message.chat.id
        chat_type = m.message.chat.type
        reply_text = m.edit_message_text
    else:
        chat_id = m.chat.id
        chat_type = m.chat.type
        reply_text = m.reply_text
    langs = sorted(list(loaded_locales.keys()))
    keyboard = [
        [
            (
                f'{loaded_locales.get(lang)["core"]["flag"]} {loaded_locales.get(lang)["core"]["name"]} ({loaded_locales.get(lang)["core"]["code"]})',
                f"set_lang {lang}",
            )
            for lang in langs
        ]
    ]
    if chat_type == ChatType.PRIVATE:
        keyboard += [[(await strings(m, "Main.back_btn"), "start_command")]]
    else:
        try:
            member = await c.get_chat_member(chat_id=chat_id, user_id=m.from_user.id)
            if member.status not in (
                ChatMemberStatus.ADMINISTRATOR,
                ChatMemberStatus.OWNER,
            ):
                if isinstance(m, CallbackQuery):
                    await m.answer(
                        await strings(m, "Admin.not_admin"),
                        show_alert=True,
                        cache_time=60,
                    )
                else:
                    message = await reply_text(await strings(m, "Admin.not_admin"))
                    await asyncio.sleep(5.0)
                    await message.delete()
                return
        except AttributeError:
            message = await reply_text(await strings(m, "Main.change_lang_uchannel"))
            await asyncio.sleep(5.0)
            return await message.delete()
        except UserNotParticipant:
            return
    return await reply_text(
        await strings(m, "Main.select_lang"), reply_markup=ikb(keyboard)
    )


@Client.on_callback_query(filters.regex(r"btn_help"))
async def help(c: Client, cq: CallbackQuery):
    keyboard = [[(await strings(cq, "Main.back_btn"), "start_command")]]
    await cq.edit_message_text(
        await strings(cq, "Main.help_msg"), reply_markup=ikb(keyboard)
    )
