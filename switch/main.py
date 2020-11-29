#!/usr/bin/env python3

import asyncio
import os
from nats.aio.client import Client as NATS
import pigpio
import json

nc = NATS()
pi = pigpio.pi()

pi.set_mode(5, pigpio.INPUT)
pi.set_pull_up_down(5, pigpio.PUD_UP)
pi.set_mode(6, pigpio.INPUT)
pi.set_pull_up_down(6, pigpio.PUD_UP)


async def run(loop):
    await nc.connect(os.getenv('NATS_SERVER'), loop=loop)

    def cb_interrupt(gpio, level, tick):
        n = gpio - 4
        nc.publish("work.wtks.home.switch." + str(n), json.dumps({
            "t": tick,
            "l": level
        }).encode())

    pi.callback(5, pigpio.EITHER_EDGE, cb_interrupt)
    pi.callback(6, pigpio.EITHER_EDGE, cb_interrupt)


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(loop))
    try:
        loop.run_forever()
    finally:
        loop.close()
