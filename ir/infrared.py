import pigpio

# 定数
FORMAT_AEHA = "AEHA"  # AEHAフォーマット
FORMAT_NEC = "NEC"  # NECフォーマット
FORMAT_SONY = "SONY"  # SONYフォーマット

_T_AEHA = 425  # AEHAフォーマットの基準周期[us]
_T_NEC = 560  # NECフォーマットの基準周期[us]
_T_SONY = 600  # SONYフォーマットの基準周期[us]
_T_WAIT = 10000  # encode, decodeで使用するフレーム間の時間[us]


class Infrared:

    # 初期化
    #   Parameters
    #     gpio_send  : 赤外線LEDのGPIO番号
    def __init__(self, gpio_send=13):
        self.gpio_send = gpio_send

    # gpio_sendで指定したGPIOから赤外線データを送信
    #   Return
    #     True  : 成功
    #     False : pigpioに接続失敗
    def send(self, code):
        pi = pigpio.pi()
        if not pi.connected:
            return False

        pi.set_mode(self.gpio_send, pigpio.OUTPUT)

        # 生成できる波形の長さには制限があるので、種類とcodeの長さごとにまとめて節約する
        mark_wids = {}  # Mark(38kHzパルス)波形, key:長さ, value:ID
        space_wids = {}  # Speace(待機)波形, key:長さ, value:ID
        send_wids = [0] * len(code)  # 送信する波形IDのリスト

        pi.wave_clear()

        for i in range(len(code)):
            if i % 2 == 0:
                # 同じ長さのMark波形が無い場合は新しく生成
                if code[i] not in mark_wids:
                    pulses = []
                    n = code[i] // 26  # 38kHz = 26us周期の繰り返し回数
                    for j in range(n):
                        pulses.append(pigpio.pulse(1 << self.gpio_send, 0, 8))  # 8us highパルス
                        pulses.append(pigpio.pulse(0, 1 << self.gpio_send, 18))  # 18us lowパルス
                    pi.wave_add_generic(pulses)
                    mark_wids[code[i]] = pi.wave_create()
                send_wids[i] = mark_wids[code[i]]
            else:
                # 同じ長さのSpace波形が無い場合は新しく生成
                if code[i] not in space_wids:
                    pi.wave_add_generic([pigpio.pulse(0, 0, code[i])])
                    space_wids[code[i]] = pi.wave_create()
                send_wids[i] = space_wids[code[i]]

        pi.wave_chain(send_wids)
        pi.wave_clear()
        pi.stop()

        return True

    # nをm単位で丸め処理を行う
    def _round(self, n, m):
        return (n + m // 2) // m * m

    # frames(バイト列データ)とフォーマットからcodeを生成する
    #   Return
    #     code
    #   Parameters
    #     ir_format : フォーマット. FORMAT_で始まる定数
    #     frames    : バイト列データ
    def encode(self, ir_format, frames):
        code = []

        if ir_format == FORMAT_AEHA:
            t = _T_AEHA
        elif ir_format == FORMAT_NEC:
            t = _T_NEC
        elif ir_format == FORMAT_SONY:
            t = _T_SONY
        else:
            return []

        first_frame = True

        for frame in frames:
            try:
                if len(frame) == 0:
                    return []
            except:
                return []

            # Wait部
            if not first_frame:
                if ir_format == FORMAT_AEHA:
                    code.append(_T_WAIT)
                elif ir_format == FORMAT_NEC:
                    code.append(self._round(108000 - t * t_count, 100))
                elif ir_format == FORMAT_SONY:
                    code.append(self._round(45000 - t * t_count, 100))

            t_count = 0

            # Leader
            if ir_format == FORMAT_AEHA:
                code.append(t * 8)
                code.append(t * 4)
                t_count += 12
            elif ir_format == FORMAT_NEC:
                code.append(t * 16)
                code.append(t * 8)
                t_count += 24
            elif ir_format == FORMAT_SONY:
                code.append(t * 4)
                t_count += 4
            else:
                return code

            # Data
            if ir_format == FORMAT_AEHA or ir_format == FORMAT_NEC:
                for byte in frame:
                    d = byte
                    for i in range(8):
                        bit = d & 1
                        if bit == 0:
                            code.append(t)
                            code.append(t)
                            t_count += 2
                        else:
                            code.append(t)
                            code.append(t * 3)
                            t_count += 4
                        d = d >> 1
            elif ir_format == FORMAT_SONY:
                d = frame[0] + (frame[1] << 7)
                if frame[1] >= 0x100:
                    bits = 20
                elif frame[1] >= 0x20:
                    bits = 15
                else:
                    bits = 12

                for i in range(bits):
                    bit = d & 1
                    if bit == 0:
                        code.append(t)
                        code.append(t)
                        t_count += 2
                    else:
                        code.append(t)
                        code.append(t * 2)
                        t_count += 3
                    d = d >> 1

            # Stop bit
            if ir_format == FORMAT_AEHA or ir_format == FORMAT_NEC:
                code.append(t)

            first_frame = False

        return code
