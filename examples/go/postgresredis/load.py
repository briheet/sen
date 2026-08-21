#!/usr/bin/env python3
"""Drive sustained API, PostgreSQL, and Redis load for five minutes."""

import argparse
import collections
import concurrent.futures
import random
import threading
import time
import urllib.error
import urllib.request


USERS = ("alice", "bob", "carol", "dora", "eve", "frank", "grace", "heidi")


class RateLimiter:
    def __init__(self, rate: float) -> None:
        self.interval = 1.0 / rate if rate > 0 else 0.0
        self.next_request = time.monotonic()
        self.lock = threading.Lock()

    def wait(self, deadline: float) -> bool:
        if self.interval == 0:
            return time.monotonic() < deadline
        with self.lock:
            now = time.monotonic()
            scheduled = max(now, self.next_request)
            self.next_request = scheduled + self.interval
        if scheduled >= deadline:
            return False
        time.sleep(max(0.0, scheduled - now))
        return True


def worker(
    base_url: str,
    lane: str,
    worker_id: int,
    deadline: float,
    timeout: float,
    limiter: RateLimiter,
) -> tuple[collections.Counter[str], collections.Counter[str], int]:
    rng = random.Random(worker_id)
    completed: collections.Counter[str] = collections.Counter()
    failed: collections.Counter[str] = collections.Counter()
    response_bytes = 0
    sequence = 0

    while limiter.wait(deadline):
        if lane == "api":
            path = "/health" if rng.random() < 0.8 else "/"
        elif lane == "visit":
            user = USERS[(worker_id + sequence) % len(USERS)]
            path = f"/visit?user={user}&seed={worker_id}-{sequence}"
        else:
            path = "/report"
        sequence += 1

        request = urllib.request.Request(
            base_url + path,
            headers={"Connection": "close", "User-Agent": "sen-load/1"},
        )
        try:
            with urllib.request.urlopen(request, timeout=timeout) as response:
                response_bytes += len(response.read())
                if 200 <= response.status < 300:
                    completed[lane] += 1
                else:
                    failed[lane] += 1
        except (OSError, TimeoutError, urllib.error.URLError):
            failed[lane] += 1

    return completed, failed, response_bytes


def send_requests(
    base_url: str,
    duration: float,
    workers: int,
    rate: float,
    timeout: float,
) -> None:
    api_workers = max(1, workers * 65 // 100)
    visit_workers = max(1, workers * 30 // 100)
    report_workers = max(1, workers - api_workers - visit_workers)
    lanes = (["api"] * api_workers) + (["visit"] * visit_workers) + (["report"] * report_workers)
    limiter = RateLimiter(rate)
    started = time.monotonic()
    deadline = started + duration

    completed: collections.Counter[str] = collections.Counter()
    failed: collections.Counter[str] = collections.Counter()
    response_bytes = 0
    with concurrent.futures.ThreadPoolExecutor(max_workers=len(lanes)) as pool:
        futures = [
            pool.submit(worker, base_url, lane, index, deadline, timeout, limiter)
            for index, lane in enumerate(lanes)
        ]
        for future in concurrent.futures.as_completed(futures):
            worker_completed, worker_failed, worker_bytes = future.result()
            completed.update(worker_completed)
            failed.update(worker_failed)
            response_bytes += worker_bytes

    elapsed = time.monotonic() - started
    successful = sum(completed.values())
    failures = sum(failed.values())
    total = successful + failures
    print(
        f"duration={elapsed:.1f}s requests={total} rate={total / elapsed:.1f}/s "
        f"ok={successful} failed={failures} response_bytes={response_bytes}"
    )
    for lane in ("api", "visit", "report"):
        print(f"{lane}: ok={completed[lane]} failed={failed[lane]}")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default="http://127.0.0.1:8080")
    parser.add_argument("--duration", type=float, default=300, help="seconds to run")
    parser.add_argument("--workers", type=int, default=128, help="concurrent requests")
    parser.add_argument("--rate", type=float, default=0, help="total requests/s; 0 is uncapped")
    parser.add_argument("--timeout", type=float, default=5, help="request timeout in seconds")
    parser.add_argument("--delay", type=float, default=3, help="seconds to wait before starting")
    args = parser.parse_args()
    if args.duration <= 0 or args.workers < 3 or args.rate < 0 or args.timeout <= 0 or args.delay < 0:
        parser.error("duration and timeout must be positive, workers >= 3, rate >= 0, and delay >= 0")

    time.sleep(args.delay)
    send_requests(args.url.rstrip("/"), args.duration, args.workers, args.rate, args.timeout)


if __name__ == "__main__":
    main()
