#!/usr/bin/env python3
"""Send bounded concurrent traffic to the example HTTP server."""

import argparse
import concurrent.futures
import itertools
import time
import urllib.request


def send_requests(base_url: str, duration: float, rate: float, workers: int) -> tuple[int, int]:
    paths = itertools.cycle(("/work?name=sen", "/?name=sen"))
    deadline = time.monotonic() + duration
    interval = 1 / rate

    def worker(path: str) -> bool:
        try:
            with urllib.request.urlopen(base_url + path, timeout=1) as response:
                response.read()
            return True
        except OSError:
            return False

    sent = failed = 0
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as pool:
        pending: set[concurrent.futures.Future[bool]] = set()
        while time.monotonic() < deadline or pending:
            while time.monotonic() < deadline and len(pending) < workers:
                pending.add(pool.submit(worker, next(paths)))
                time.sleep(interval)
            done, pending = concurrent.futures.wait(
                pending,
                timeout=max(0, deadline - time.monotonic()),
                return_when=concurrent.futures.FIRST_COMPLETED,
            )
            for result in done:
                sent += 1
                failed += not result.result()
    return sent, failed


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default="http://127.0.0.1:8080")
    parser.add_argument("--duration", type=float, default=10)
    parser.add_argument("--rate", type=float, default=40, help="total requests per second")
    parser.add_argument("--workers", type=int, default=8)
    parser.add_argument("--delay", type=float, default=0)
    args = parser.parse_args()
    if args.duration <= 0 or args.rate <= 0 or args.workers <= 0 or args.delay < 0:
        parser.error("duration, rate, and workers must be positive; delay cannot be negative")

    time.sleep(args.delay)
    sent, failed = send_requests(args.url.rstrip("/"), args.duration, args.rate, args.workers)
    print(f"sent={sent} failed={failed}")


if __name__ == "__main__":
    main()
