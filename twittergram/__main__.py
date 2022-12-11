# SPDX-License-Identifier: GPL-3.0
# Copyright (c) 2021-2022 Luiz Renato (ruizlenato@proton.me)
import uvloop
import logging
import asyncio

from pyrogram import idle

from .bot import Client

# Custom logging format
log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format="\u001B[33m%(levelname)s | \u001B[35m%(name)s \u001B[31m| %(asctime)s | \u001B[37m%(message)s",
    datefmt="%m/%d %H:%M:%S",
)
logging.getLogger("pyrogram").setLevel(logging.WARNING)

uvloop.install()


async def main():
    client = Client()

    try:
        # start the bot
        await client.start()
        log.info("\033[92m[🚀] - Bot started.\033[0m")
        await idle()
    except KeyboardInterrupt:
        log.warning("Forced stop... Bye!")
    finally:
        await client.stop()
        log.warning("\033[93mBye!\033[0m")


if __name__ == "__main__":
    # open new asyncio event loop
    event_policy = asyncio.get_event_loop_policy()
    event_loop = event_policy.new_event_loop()
    asyncio.set_event_loop(event_loop)

    # start the bot
    event_loop.run_until_complete(main())

    # close asyncio event loop
    event_loop.close()
