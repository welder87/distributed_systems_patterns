import asyncio
import random

from rate_limiter.token_bucket.py.naive import NaiveTokenBucket


class AsyncNaiveTokenBucket(NaiveTokenBucket):
    ...
    # критическая секция синхронная, нет переключения корутин, мьютекс не нужен.


async def check(bucket: AsyncNaiveTokenBucket, i: int):
    await asyncio.sleep(0.5 + random.uniform(0.5, 6))
    if bucket.allow():
        print(f"{i}: разрешено (осталось ~{bucket._tokens:.2f})")
    else:
        print(f"{i}: отклонено (осталось ~{bucket._tokens:.2f})")


async def main():
    bucket = AsyncNaiveTokenBucket(capacity=5, refill_rate=1)

    async with asyncio.TaskGroup() as tg:
        for i in range(20):
            tg.create_task(check(bucket, i))


if __name__ == "__main__":
    asyncio.run(main())
