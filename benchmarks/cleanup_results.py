#!/usr/bin/env python3
"""
SNID Benchmark Results Cleanup
Implements retention policy: 90-day rolling window + keep last 10 runs forever.
"""

import os
import sys
import argparse
from datetime import datetime, timedelta
from pathlib import Path

RESULTS_DIR = Path(os.getenv("RESULTS_DIR", str(Path(__file__).parent / "results")))


def cleanup_results(days: int = 90, keep_last: int = 10, dry_run: bool = False) -> None:
    """Clean up old benchmark results according to retention policy."""
    if not RESULTS_DIR.exists():
        print(f"Results directory does not exist: {RESULTS_DIR}")
        return
    
    # Get all result files sorted by modification time (newest first)
    files = sorted(
        RESULTS_DIR.glob("*.json"),
        key=lambda f: f.stat().st_mtime,
        reverse=True
    )
    
    if not files:
        print("No result files found")
        return
    
    cutoff_date = datetime.now() - timedelta(days=days)
    files_to_delete = []
    
    # Keep last N files regardless of age
    files_to_keep = set(files[:keep_last])
    
    for file in files:
        if file in files_to_keep:
            continue
        
        file_mtime = datetime.fromtimestamp(file.stat().st_mtime)
        
        if file_mtime < cutoff_date:
            files_to_delete.append(file)
    
    if dry_run:
        print(f"[DRY RUN] Would delete {len(files_to_delete)} files:")
        for f in files_to_delete:
            print(f"  - {f.name} (modified {datetime.fromtimestamp(f.stat().st_mtime)})")
    else:
        for f in files_to_delete:
            try:
                f.unlink()
                print(f"Deleted: {f.name}")
            except OSError as e:
                print(f"Failed to delete {f.name}: {e}")
        
        print(f"Cleanup complete. Deleted {len(files_to_delete)} files.")
        print(f"Kept {len(files) - len(files_to_delete)} files.")


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(description="Clean up old benchmark results")
    parser.add_argument(
        "--days",
        type=int,
        default=90,
        help="Retention period in days (default: 90)"
    )
    parser.add_argument(
        "--keep-last",
        type=int,
        default=10,
        help="Number of recent files to keep regardless of age (default: 10)"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be deleted without actually deleting"
    )
    
    args = parser.parse_args()
    
    print(f"Results directory: {RESULTS_DIR}")
    print(f"Retention policy: {args.days} days, keep last {args.keep_last} files")
    
    cleanup_results(days=args.days, keep_last=args.keep_last, dry_run=args.dry_run)


if __name__ == "__main__":
    main()
