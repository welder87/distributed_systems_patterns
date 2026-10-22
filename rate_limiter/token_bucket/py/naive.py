import random
import time


class NaiveTokenBucket:
    def __init__(self, capacity: float, refill_rate: float):
        """
        :param capacity: максимальное число токенов (размер "ведра").
        :param refill_rate: сколько токенов добавляется в секунду.
        """
        self._capacity: float = capacity
        self._refill_rate: float = refill_rate
        self._tokens: float = capacity  # старт с полным ведром
        # monotonic() гарантированно идёт только вперёд.
        # Нельзя использовать time.time(), та как время может прыгнуть назад (например,
        # при синхронизации NTP), что приведёт к отрицательному elapsed и некорректному
        # пересчёту.
        self._last_refill: float = time.monotonic()

    def allow(self, cost: int = 1) -> bool:
        # ленивый пересчет токенов - стандартная и рекомендуемая практика
        self._refill()
        if self._tokens < cost:
            return False
        self._tokens -= cost
        return True

    def _refill(self):
        now = time.monotonic()
        # сколько секунд прошло с момента последнего пополнения ведра
        elapsed = now - self._last_refill
        # пересчет токенов
        self._tokens = min(self._capacity, self._refill_calc(elapsed))
        # сохранение текущего времени
        self._last_refill = now

    def _refill_calc(self, elapsed: float) -> float:
        return self._tokens + elapsed * self._refill_rate


if __name__ == "__main__":
    bucket = NaiveTokenBucket(capacity=5, refill_rate=1)  # 5 токенов, +1 в секунду

    for i in range(20):
        if bucket.allow():
            print(f"{i}: разрешено (осталось ~{bucket._tokens:.2f})")
        else:
            print(f"{i}: отклонено (осталось ~{bucket._tokens:.2f})")
        time.sleep(0.5 + random.uniform(-0.5, 0.5))
