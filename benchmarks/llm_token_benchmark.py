#!/usr/bin/env python3
"""
SNID LLM Token Efficiency Benchmark
Measures token cost of SNID representations in LLM contexts.
"""

import json
import sys
from pathlib import Path

# Add parent directory to path for snid import
sys.path.insert(0, str(Path(__file__).parent.parent / "python"))

try:
    from snid import SNID
except ImportError:
    print("⚠️  SNID module not available. Run: cd python && maturin develop")
    SNID = None


def count_tokens(text, model="gpt-4"):
    """
    Estimate token count using tiktoken.
    Falls back to character-based estimation if tiktoken not available.
    """
    try:
        import tiktoken

        encoding = tiktoken.encoding_for_model(model)
        return len(encoding.encode(text))
    except ImportError:
        # Fallback: rough estimate (1 token ≈ 4 characters for English)
        return len(text) // 4


def benchmark_llm_formats():
    """Benchmark various SNID representations for LLM token efficiency."""
    if SNID is None:
        return {"error": "SNID module not available"}

    # Generate test ID
    snid = SNID.new_fast()

    # Different representations
    representations = {
        "wire_standard": snid.to_wire("MAT"),
        "wire_underscore": snid.to_wire("MAT").replace(":", "_"),
        "hex_bytes": snid.to_bytes().hex(),
        "uuid_format": str(snid.to_bytes()).replace("-", ""),  # UUID-like
        "llm_format_v1": str(snid.to_llm_format("MAT")),
        "llm_format_v2": str(snid.to_llm_format_v2("MAT")),
    }

    results = {}
    for name, repr_text in representations.items():
        token_count = count_tokens(repr_text)
        results[name] = {
            "representation": repr_text,
            "character_count": len(repr_text),
            "token_count": token_count,
            "tokens_per_char": token_count / len(repr_text) if repr_text else 0,
        }

    return results


def main():
    """Run LLM token efficiency benchmark."""
    print("=" * 60)
    print("SNID LLM Token Efficiency Benchmark")
    print("=" * 60)

    results = benchmark_llm_formats()

    if "error" in results:
        print(f"❌ Error: {results['error']}")
        return 1

    # Print results
    print("\n📊 Token Efficiency Results:\n")
    for name, data in results.items():
        print(f"{name:20} {data['token_count']:4} tokens ({data['character_count']:3} chars)")

    # Find most efficient
    most_efficient = min(results.items(), key=lambda x: x[1]["token_count"])
    print(f"\n✅ Most efficient: {most_efficient[0]} ({most_efficient[1]['token_count']} tokens)")

    # Save to results
    results_dir = Path(__file__).parent / "results"
    results_dir.mkdir(parents=True, exist_ok=True)
    output_file = results_dir / "llm_tokens.json"
    with open(output_file, "w") as f:
        json.dump(results, f, indent=2)

    print(f"📁 Results saved to: {output_file}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
