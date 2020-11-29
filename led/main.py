#!/usr/bin/env python3

import asyncio
import os
from nats.aio.client import Client as NATS
import pigpio


nc = NATS()
pi = pigpio.pi()

pi.set_mode(17, pigpio.OUTPUT)
pi.set_mode(18, pigpio.OUTPUT)
pi.set_mode(22, pigpio.OUTPUT)
pi.set_mode(27, pigpio.OUTPUT)

async def run(loop):
    await nc.connect(os.getenv('NATS_SERVER'), loop=loop)

    async def subscribe_handler(msg):
        num = int(msg.subject.rsplit('.', 1)[-1])
        on = msg.data.decode() == 'on'

        if num == 1:
            pi.write(17, on)
        elif num == 2:
            pi.write(18, on)
        elif num == 3:
            pi.write(22, on)
        elif num == 4:
            pi.write(27, on)

    await nc.subscribe("work.wtks.home.led.*", cb=subscribe_handler)


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(loop))
    try:
        loop.run_forever()
    finally:
        loop.close()
