"""Extended Family Benchmarks for SNID Python
Benchmarks for SGID, NID, LID/KID, BID, EID operations
"""

import pytest
import snid


# SGID Benchmarks
def benchmark_sgid_new_spatial_precise(benchmark):
    benchmark(snid.new_spatial_precise, 37.7749, -122.4194, 9)


def benchmark_sgid_h3_cell(benchmark):
    id = snid.new_spatial_precise(37.7749, -122.4194, 9)
    benchmark(id.h3_cell)


def benchmark_sgid_h3_string(benchmark):
    id = snid.new_spatial_precise(37.7749, -122.4194, 9)
    benchmark(id.h3_string)


def benchmark_sgid_lat_lng(benchmark):
    id = snid.new_spatial_precise(37.7749, -122.4194, 9)
    benchmark(id.lat_lng)


def benchmark_sgid_spatial_parent(benchmark):
    id = snid.new_spatial_precise(37.7749, -122.4194, 9)
    benchmark(id.spatial_parent, 7)


# NID Benchmarks
def benchmark_nid_new_neural(benchmark):
    base = snid.new()
    vec = [0.5] * 128
    benchmark(snid.new_neural, base, vec)


def benchmark_nid_distance(benchmark):
    base = snid.new()
    vec = [0.5] * 128
    nid1 = snid.new_neural(base, vec)
    nid2 = snid.new_neural(base, vec)
    benchmark(nid1.distance, nid2)


def benchmark_nid_similarity(benchmark):
    base = snid.new()
    vec = [0.5] * 128
    nid1 = snid.new_neural(base, vec)
    nid2 = snid.new_neural(base, vec)
    benchmark(nid1.similarity, nid2)


def benchmark_nid_is_similar(benchmark):
    base = snid.new()
    vec = [0.5] * 128
    nid1 = snid.new_neural(base, vec)
    nid2 = snid.new_neural(base, vec)
    benchmark(nid1.is_similar, nid2, 10)


# LID Benchmarks
def benchmark_lid_new_lid(benchmark):
    prev = snid.LID()
    payload = b"test payload"
    key = b"test-key-32-bytes-long-1234567890"
    benchmark(snid.new_lid, prev, payload, key)


def benchmark_lid_verify(benchmark):
    prev = snid.LID()
    payload = b"test payload"
    key = b"test-key-32-bytes-long-1234567890"
    lid = snid.new_lid(prev, payload, key)
    benchmark(lid.verify, prev, payload, key)


# BID Benchmarks
def benchmark_bid_new_bid_from_content(benchmark):
    content = b"test content for hashing"
    benchmark(snid.new_bid_from_content, content)


def benchmark_bid_wire_format(benchmark):
    content = b"test content for hashing"
    bid = snid.new_bid_from_content(content)
    benchmark(bid.wire_format)


def benchmark_bid_parse_wire(benchmark):
    content = b"test content for hashing"
    bid = snid.new_bid_from_content(content)
    wire = bid.wire_format()
    benchmark(snid.parse_bid_wire, wire)


# EID Benchmarks
def benchmark_eid_new_ephemeral(benchmark):
    session = 12345
    benchmark(snid.new_ephemeral, session)


def benchmark_eid_bytes(benchmark):
    eid = snid.new_ephemeral(12345)
    benchmark(eid.bytes)


def benchmark_eid_time(benchmark):
    eid = snid.new_ephemeral(12345)
    benchmark(eid.time)


def benchmark_eid_string(benchmark):
    eid = snid.new_ephemeral(12345)
    benchmark(str, eid)


# Batch operations
def benchmark_sgid_batch_100(benchmark):
    def batch():
        for _ in range(100):
            snid.new_spatial_precise(37.7749, -122.4194, 9)
    benchmark(batch)


def benchmark_sgid_batch_1000(benchmark):
    def batch():
        for _ in range(1000):
            snid.new_spatial_precise(37.7749, -122.4194, 9)
    benchmark(batch)


def benchmark_nid_batch_100(benchmark):
    base = snid.new()
    vec = [0.5] * 128
    def batch():
        for _ in range(100):
            snid.new_neural(base, vec)
    benchmark(batch)


def benchmark_nid_batch_1000(benchmark):
    base = snid.new()
    vec = [0.5] * 128
    def batch():
        for _ in range(1000):
            snid.new_neural(base, vec)
    benchmark(batch)


def benchmark_eid_batch_100(benchmark):
    session = 12345
    def batch():
        for _ in range(100):
            snid.new_ephemeral(session)
    benchmark(batch)


def benchmark_eid_batch_1000(benchmark):
    session = 12345
    def batch():
        for _ in range(1000):
            snid.new_ephemeral(session)
    benchmark(batch)
