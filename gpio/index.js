const pigpio = require('pigpio-client').pigpio({ host: 'zero.home.wtks.work' })

const ready = new Promise((resolve, reject) => {
  pigpio.once('connected', resolve)
  pigpio.once('error', reject)
})

function wait (ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

async function sendIR (code) {
  const led = pigpio.gpio(13)
  await led.modeSet('output')

  const mark_wids = {}
  const space_wids = {}
  const send_wids = []

  await led.waveClear()

  for (let i = 0; i < code.length; i++) {
    if (i % 2 === 0) {
      if (!(code[i] in mark_wids)) {
        const pulses = []
        const n = Math.floor(code[i] / 26)
        for (let j = 0; j < n; j++) {
          pulses.push([1, 0, 8])
          pulses.push([0, 1, 18])
        }
        await led.waveAddPulse(pulses)
        mark_wids[code[i]] = await led.waveCreate()
      }
      send_wids.push(mark_wids[code[i]])
    } else {
      if (!(code[i] in space_wids)) {
        await led.waveAddPulse([[0, 0, code[i]]])
        space_wids[code[i]] = await led.waveCreate()
      }
      send_wids.push(space_wids[code[i]])
    }
  }

  await led.waveChainTx([{ waves: send_wids }])
}

ready.then(async (info) => {
  // display information on pigpio and connection status
  console.log(JSON.stringify(info, null, 2))

  const btn1 = pigpio.gpio(5) // red button
  await btn1.modeSet('input')
  await btn1.pullUpDown(2)
  btn1.notify((level, tick) => {
    console.log(`btn1 ${level}`)
  })

  const btn2 = pigpio.gpio(6) // black button
  await btn2.modeSet('input')
  await btn2.pullUpDown(2)
  btn2.notify((level, tick) => {
    console.log(`btn2 ${level}`)
  })

}).catch(console.error)