from __future__ import annotations

import time

from snid import SNID


def bench(label: str, fn) -> None:
    start = time.perf_counter()
    result = fn()
    elapsed = time.perf_counter() - start
    size = len(result) if hasattr(result, "__len__") else 0
    print(f"{label}: {elapsed:.6f}s size={size}")


def main() -> None:
    n = 100_000
    bench("bytes", lambda: SNID.generate_batch(n, backend="bytes"))
    bench("tensor", lambda: SNID.generate_batch(n, backend="tensor"))
    try:
        import numpy  # noqa: F401

        bench("numpy", lambda: SNID.generate_batch(n, backend="numpy"))
    except Exception:
        print("numpy: skipped")


if __name__ == "__main__":
    main()
