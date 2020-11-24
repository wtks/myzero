#!/usr/bin/env python3

import asyncio
import json
import os
from nats.aio.client import Client as NATS
from infrared import Infrared

nc = NATS()
ir = Infrared()


async def run(loop):
    await nc.connect(os.getenv('NATS_SERVER'), loop=loop)

    async def subscribe_handler(msg):
        code = json.loads(msg.data.decode())
        if isinstance(code, list):
            print(code)
            ir.send(code)

    await nc.subscribe("work.wtks.home.ir.send_code", cb=subscribe_handler)


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(loop))
    try:
        loop.run_forever()
    finally:
        loop.close()
