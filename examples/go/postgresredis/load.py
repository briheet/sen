#!/usr/bin/env python3
"""Drive mixed API, PostgreSQL, and Redis load for the Go example."""

import argparse
import concurrent.futures
import itertools
import random
import time
import urllib.error
import urllib.request


def build_paths() -> tuple[str, ...]:
    users = ("alice", "bob", "carol", "dora", "eve", "frank")
    visits = tuple(f"/visit?user={user}" for user in users)
    reports = ("/report", "/report", "/health")
    return visits + reports


def send_requests(base_url: str, duration: float, rate: float, workers: int) -> tuple[int, int]:
    paths = itertools.cycle(build_paths())
    deadline = time.monotonic() + duration
    interval = 1.0 / rate

    def worker(path: str) -> bool:
        try:
            with urllib.request.urlopen(base_url + path, timeout=3) as response:
                response.read()
            return True
        except (OSError, urllib.error.URLError):
            return False

    sent = 0
    failed = 0
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        pending: set[concurrent.futures.Future[bool]] = set()
        next_send = time.monotonic()
        while time.monotonic() < deadline or pending:
            now = time.monotonic()
            while now < deadline and len(pending) < workers and now >= next_send:
                path = next(paths)
                if path.startswith("/visit"):
                    path += f"&seed={random.randint(0, 1_000_000)}"
                pending.add(pool.submit(worker, path))
                next_send += interval
                now = time.monotonic()
            if not pending:
                time.sleep(min(0.05, max(0.0, next_send - time.monotonic())))
                continue
            done, pending = concurrent.futures.wait(
                pending,
                timeout=min(0.25, max(0.0, deadline - time.monotonic())),
                return_when=concurrent.futures.FIRST_COMPLETED,
            )
            for result in done:
                sent += 1
                failed += not result.result()
    return sent, failed


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default="http://127.0.0.1:8080")
    parser.add_argument("--duration", type=float, default=300, help="seconds to run")
    parser.add_argument("--rate", type=float, default=25, help="total requests per second")
    parser.add_argument("--workers", type=int, default=12)
    parser.add_argument("--delay", type=float, default=3, help="seconds to wait before starting")
    args = parser.parse_args()
    if args.duration <= 0 or args.rate <= 0 or args.workers <= 0 or args.delay < 0:
        parser.error("duration, rate, and workers must be positive; delay cannot be negative")

    time.sleep(args.delay)
    sent, failed = send_requests(args.url.rstrip("/"), args.duration, args.rate, args.workers)
    print(f"duration={args.duration:.0f}s sent={sent} failed={failed}")


if __name__ == "__main__":
    main()
