import asyncio
from hypercorn.config import Config
from hypercorn.asyncio import serve
from fastapi import FastAPI

from ..shared import config
from .. import db

app = FastAPI()


async def runserver():
    db.create_tables()

    cfg = Config()
    cfg.bind = [f"localhost:{config['port']}"]

    return serve(app, cfg, shutdown_trigger=lambda: asyncio.Future())
