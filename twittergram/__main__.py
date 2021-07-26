import asyncio
import logging
import pyrogram

from rich.panel import Panel
from pyrogram import Client, idle
from rich.logging import RichHandler
from rich import box, print as rprint
from pyrogram.errors import BadRequest
from twittergram.config import *

# Enable logging
logging.basicConfig(format='%(asctime)s - %(message)s',
                    level='INFO')
logging.getLogger('pyrogram.client').setLevel(logging.WARNING)

log = logging.getLogger('rich')
logs = '[bold purple]TwitterGram Running[/bold purple]'
logs += f'\n[TwitterGram] Project maintained by: Renatoh'
rprint(Panel.fit(logs, border_style='turquoise2', box=box.ASCII))

# Pyrogram Client
plugins = dict(root='twittergram.plugins')

client = Client(
        "twittergram",
        api_id=API_ID,
        api_hash=API_HASH, 
        bot_token=BOT_TOKEN,
        parse_mode='html',
        plugins=plugins
)

async def main():
    await client.start()
    print("[TwitterGram] Starting...")
    await idle()

if __name__ == "__main__":
    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())
