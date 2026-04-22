from __future__ import annotations

import time

from snid import SNID


def bench(label: str, fn) -> None:
    # Warmup
    for _ in range(3):
        fn()
    
    # Actual benchmark with multiple iterations
    import timeit
    iterations = 5
    times = []
    for _ in range(iterations):
        start = time.perf_counter()
        result = fn()
        elapsed = time.perf_counter() - start
        times.append(elapsed)
    
    avg_time = sum(times) / len(times)
    stddev = (sum((t - avg_time) ** 2 for t in times) / len(times)) ** 0.5
    size = len(result) if hasattr(result, "__len__") else 0
    print(f"{label}: {avg_time:.6f}s (±{stddev:.6f}s) size={size}")


def main() -> None:
    n = 100_000
    bench("bytes", lambda: SNID.generate_batch(n, backend="bytes"))
    bench("tensor", lambda: SNID.generate_batch(n, backend="tensor"))
    try:
        import numpy  # noqa: F401

        bench("numpy", lambda: SNID.generate_batch(n, backend="numpy"))
    except Exception:
        print("numpy: skipped")

    # Benchmark NewSafe
    n_safe = 10_000
    bench("new_safe", lambda: [SNID.new_safe() for _ in range(n_safe)])


if __name__ == "__main__":
    main()
