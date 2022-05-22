import asyncio
import logging
import pyrogram

from rich.panel import Panel
from pyrogram import Client, enums
from rich import box, print
from twittergram.config import API_HASH, API_ID, BOT_TOKEN

# Enable logging
logging.basicConfig(format="%(asctime)s - %(message)s", level="WARNING")
logging.getLogger("pyrogram.client").setLevel(logging.WARNING)
log = logging.getLogger("rich")
logs = "[bold purple]TwitterGram Running[/bold purple]"
logs += f"\n[TwitterGram] Project maintained by: Renatoh"
print(Panel.fit(logs, border_style="turquoise2", box=box.ASCII))

# Pyrogram Client
class Twittegram(Client):
    def __init__(self):
        name = self.__class__.__name__.lower()

        super().__init__(
            name=name,
            api_hash=API_HASH,
            api_id=API_ID,
            bot_token=BOT_TOKEN,
            parse_mode=enums.ParseMode.HTML,
            workers=24,
            workdir="twittergram",
            plugins={"root": "twittergram.plugins"},
        )

    async def start(self):
        await super().start()  # Connect to telegram's servers
        print("[green] TwitterGram Started...")

    async def stop(self, *args):
        await super().stop()  # Disconnect from telegram's servers
        print("[red] TwitterGram Stopped, Bye.")


if __name__ == "__main__":
    try:
        Twittegram().run()
    except KeyboardInterrupt:
        log.warning("Forced stop.")
