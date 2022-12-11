# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
from pyrogram import Client
from pyrogram.enums import ParseMode

from .utils.twitter import http
from .database import database
from .config import API_HASH, API_ID, BOT_TOKEN


class Client(Client):
    def __init__(self):
        name = self.__class__.__name__.lower()

        super().__init__(
            name,
            bot_token=BOT_TOKEN,
            api_hash=API_HASH,
            api_id=API_ID,
            workers=24,
            parse_mode=ParseMode.HTML,
            workdir="twittergram",
            sleep_threshold=180,
            plugins={"root": "twittergram.plugins"},
        )

    async def start(self):
        await database.connect()
        await super().start()  # Connect to telegram's servers

    async def stop(self) -> None:
        await http.aclose()
        if database.is_connected:
            await database.close()
        await super().stop()
