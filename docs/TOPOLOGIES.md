# SNID Topologies

## Core SNID

Use for canonical entity, event, session, tenant, and matter identifiers. The atom is a presentation tag and is not stored in the 16-byte payload.

## SGID

Use for spatial anchoring and lexicographic range scans by H3 cell. The stored bytes preserve H3 locality while remaining SNID-compatible at the wire layer.

## NID

Use for vector-search adjacency. A NID keeps a causal SNID head and a 128-bit semantic tail that supports Hamming distance and similarity calculations.

## LID

Use for causal ledger chains. The head remains join-friendly as a SNID while the tail binds the current record to the previous record and payload with a keyed HMAC.

## WID

Use for world, scenario, and simulation partitioning. A WID keeps a causal SNID head and adds a scenario hash tail for world-level isolation.

## XID

Use for first-class relationship identity. An XID keeps a causal SNID head and adds an edge hash tail for relationship-level joins and auditing.

## KID

Use for self-verifying capability checks. A KID keeps a causal SNID head and adds a MAC tail bound to actor, resource, and capability bytes.

## EID

Use for very short-lived internal identifiers such as ticks or session-scoped telemetry where 64 bits are sufficient.

## BID

Use when topology and content identity both matter:
- topology for graph, joins, or object locality
- content hash for deduplication and integrity
