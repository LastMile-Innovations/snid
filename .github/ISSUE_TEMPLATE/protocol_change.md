---
name: Protocol Change
about: Propose a change to the SNID protocol specification
title: "[PROTOCOL] "
labels: protocol
assignees: ''
---

## Protocol Change Description

A clear and concise description of the proposed protocol change.

## Motivation

Why is this protocol change needed? What problem does it solve?

## Proposed Change

Describe the change in detail, including:

- Byte layout changes (if any)
- Wire format changes (if any)
- New identifier families (if any)
- New boundary projections (if any)
- Changes to verification contracts (if any)

## Impact Assessment

- [ ] This changes the byte layout
- [ ] This changes the wire format
- [ ] This adds a new identifier family
- [ ] This changes verification contracts
- [ ] This is additive only (backward compatible)
- [ ] This is a breaking change

## Implementation Plan

How would this be implemented across Go, Rust, and Python?

## Conformance Impact

- [ ] Requires new test vectors
- [ ] Requires changes to existing test vectors
- [ ] No conformance impact

## Migration Path

If this is a breaking change, describe the migration path for existing users.

## Additional Context

Add any other context, diagrams, or examples about the protocol change.

**Note:** Protocol changes require consensus and must be approved by maintainers. All three implementations must pass the updated conformance suite before the change can be merged.
