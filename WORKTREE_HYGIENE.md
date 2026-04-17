# SNID Worktree Hygiene

This repo should stay small, portable, and deterministic.

## Do Not Commit

- `.DS_Store`
- `.venv/`
- `target/`
- `.gocache/`, `.gomodcache/`, `.gotmp/`
- native build outputs unless intentionally published
- editor-local settings such as `.claude/`

## Rules

1. Keep conformance artifacts intentional.
2. Prefer relative repo paths in scripts and tests.
3. Treat the spec and reference implementations as the canonical surfaces.
