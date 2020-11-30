#!/usr/bin/env python3

import asyncio
import os
from nats.aio.client import Client as NATS
import apigpio
import json

nc = NATS()
loop = asyncio.get_event_loop()
pi = apigpio.Pi(loop)


async def run(loop):
    await nc.connect(os.getenv('NATS_SERVER'), loop=loop)

    await pi.connect(('127.0.0.1', 8888))
    await pi.set_mode(5, apigpio.INPUT)
    await pi.set_pull_up_down(5, apigpio.PUD_UP)
    await pi.set_mode(6, apigpio.INPUT)
    await pi.set_pull_up_down(6, apigpio.PUD_UP)

    def cb_interrupt(gpio, level, tick):
        n = gpio - 4
        print(n, level, tick)
        nc.publish("work.wtks.home.switch." + str(n), json.dumps({
            "t": tick,
            "l": level
        }).encode())

    await pi.add_callback(5, apigpio.EITHER_EDGE, cb_interrupt)
    await pi.add_callback(6, apigpio.EITHER_EDGE, cb_interrupt)


if __name__ == '__main__':
    loop.run_until_complete(run(loop))
    try:
        loop.run_forever()
    finally:
        loop.close()
