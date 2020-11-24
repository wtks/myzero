#!/usr/bin/env python3

from bme280i2c import BME280I2C
from tsl2572 import TSL2572
import asyncio
import json
import os
from nats.aio.client import Client as NATS

nc = NATS()


async def run(loop):
    await nc.connect(os.getenv('NATS_SERVER'), loop=loop)

    bme280ch1 = BME280I2C(0x76)
    bme280ch2 = BME280I2C(0x77)
    tsl2572 = TSL2572(0x39)

    while True:
        bme280ch1.meas()
        bme280ch2.meas()
        tsl2572.meas_single()

        await nc.publish("work.wtks.home.envsensor", json.dumps({
            "t1": bme280ch1.T,
            "p1": bme280ch1.P,
            "h1": bme280ch1.H,
            "t2": bme280ch2.T,
            "p2": bme280ch2.P,
            "h2": bme280ch2.H,
            "l": tsl2572.lux
        }).encode())
        await asyncio.sleep(10)


if __name__ == '__main__':
    try:
        loop = asyncio.get_event_loop()
        loop.run_until_complete(run(loop))
        loop.close()
    except KeyboardInterrupt:
        pass
    finally:
        await nc.close()
