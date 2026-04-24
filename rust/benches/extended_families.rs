// Extended Family Benchmarks for SNID
// Benchmarks for NID, LID, BID, EID operations

use criterion::{criterion_group, criterion_main, Criterion, Throughput};
use snid::{Bid, Eid, Nid, Snid, TraceId};
#[cfg(feature = "crypto")]
use snid::{GrantId, Kid, Lid};
use std::hint::black_box;
#[cfg(feature = "crypto")]
use std::time::Duration;

fn bench_nid_new(c: &mut Criterion) {
    let base = Snid::new_fast();
    let vec: Vec<f32> = (0..128).map(|i| i as f32 / 128.0).collect();
    let semantic_hash = Nid::quantize(&vec);

    c.benchmark_group("nid_new")
        .throughput(Throughput::Elements(1))
        .bench_function("nid_new", |b| {
            b.iter(|| black_box(Nid::from_parts(base, semantic_hash)));
        });
}

fn bench_nid_hamming(c: &mut Criterion) {
    let base = Snid::new_fast();
    let vec: Vec<f32> = (0..128).map(|i| i as f32 / 128.0).collect();
    let semantic_hash = Nid::quantize(&vec);
    let nid1 = Nid::from_parts(base, semantic_hash);
    let nid2 = Nid::from_parts(base, semantic_hash);

    c.benchmark_group("nid_hamming_distance")
        .throughput(Throughput::Elements(1))
        .bench_function("nid_hamming_distance", |b| {
            b.iter(|| black_box(nid1.hamming_distance(black_box(&nid2))));
        });
}

fn bench_nid_batch_100(c: &mut Criterion) {
    let base = Snid::new_fast();
    let vec: Vec<f32> = (0..128).map(|i| i as f32 / 128.0).collect();
    let semantic_hash = Nid::quantize(&vec);

    c.benchmark_group("nid_batch_100")
        .throughput(Throughput::Elements(100))
        .bench_function("nid_batch_100", |b| {
            b.iter(|| {
                for _ in 0..100 {
                    black_box(Nid::from_parts(base, semantic_hash));
                }
            });
        });
}

fn bench_nid_batch_1000(c: &mut Criterion) {
    let base = Snid::new_fast();
    let vec: Vec<f32> = (0..128).map(|i| i as f32 / 128.0).collect();
    let semantic_hash = Nid::quantize(&vec);

    c.benchmark_group("nid_batch_1000")
        .throughput(Throughput::Elements(1000))
        .bench_function("nid_batch_1000", |b| {
            b.iter(|| {
                for _ in 0..1000 {
                    black_box(Nid::from_parts(base, semantic_hash));
                }
            });
        });
}

#[cfg(feature = "crypto")]
fn bench_lid_new(c: &mut Criterion) {
    let base = Snid::new_fast();
    let prev = [0u8; 32];
    let payload = b"test payload";
    let key = b"test-key-32-bytes-long-1234567890";

    c.benchmark_group("lid_new")
        .throughput(Throughput::Elements(1))
        .bench_function("lid_new", |b| {
            b.iter(|| black_box(Lid::from_parts(base, prev, payload, key).unwrap()));
        });
}

#[cfg(feature = "crypto")]
fn bench_lid_head(c: &mut Criterion) {
    let base = Snid::new_fast();
    let prev = [0u8; 32];
    let payload = b"test payload";
    let key = b"test-key-32-bytes-long-1234567890";
    let lid = Lid::from_parts(base, prev, payload, key).unwrap();

    c.benchmark_group("lid_head")
        .throughput(Throughput::Elements(1))
        .bench_function("lid_head", |b| {
            b.iter(|| black_box(lid.head()));
        });
}

#[cfg(feature = "crypto")]
fn bench_kid_verify(c: &mut Criterion) {
    let head = Snid::new_fast();
    let actor = Snid::new_fast();
    let resource = b"resource";
    let capability = b"read";
    let key = b"test-key-32-bytes-long-1234567890";
    let kid = Kid::from_parts(head, actor, resource, capability, key).unwrap();

    c.benchmark_group("kid_verify")
        .throughput(Throughput::Elements(1))
        .bench_function("kid_verify", |b| {
            b.iter(|| {
                black_box(kid.verify(
                    black_box(actor),
                    black_box(resource),
                    black_box(capability),
                    black_box(key),
                ))
            });
        });
}

#[cfg(feature = "crypto")]
fn bench_grant_new(c: &mut Criterion) {
    let key = b"grant-key-32-bytes-long-1234567890";

    c.benchmark_group("grant_new")
        .throughput(Throughput::Elements(1))
        .bench_function("grant_new", |b| {
            b.iter(|| {
                black_box(
                    GrantId::new(
                        black_box("MAT"),
                        black_box(Some(Duration::from_secs(60))),
                        black_box(key),
                    )
                    .unwrap(),
                )
            });
        });
}

#[cfg(feature = "crypto")]
fn bench_grant_verify(c: &mut Criterion) {
    let key = b"grant-key-32-bytes-long-1234567890";
    let grant = GrantId::new("MAT", Some(Duration::from_secs(60)), key).unwrap();

    c.benchmark_group("grant_verify")
        .throughput(Throughput::Elements(1))
        .bench_function("grant_verify", |b| {
            b.iter(|| black_box(grant.verify(black_box(key))));
        });
}

fn bench_bid_new(c: &mut Criterion) {
    let base = Snid::new_fast();
    let content = [0u8; 32];

    c.benchmark_group("bid_new")
        .throughput(Throughput::Elements(1))
        .bench_function("bid_new", |b| {
            b.iter(|| black_box(Bid::from_parts(base, content)));
        });
}

fn bench_bid_wire(c: &mut Criterion) {
    let base = Snid::new_fast();
    let content = [0u8; 32];
    let bid = Bid::from_parts(base, content);

    c.benchmark_group("bid_wire")
        .throughput(Throughput::Elements(1))
        .bench_function("bid_wire", |b| {
            b.iter(|| black_box(bid.wire().unwrap()));
        });
}

fn bench_bid_write_wire(c: &mut Criterion) {
    let base = Snid::new_fast();
    let content = [0u8; 32];
    let bid = Bid::from_parts(base, content);

    c.benchmark_group("bid_write_wire")
        .throughput(Throughput::Elements(1))
        .bench_function("bid_write_wire", |b| {
            b.iter(|| {
                let mut out = [0u8; 81];
                black_box(bid.write_wire(black_box(&mut out)).unwrap().len())
            });
        });
}

fn bench_bid_append_wire(c: &mut Criterion) {
    let base = Snid::new_fast();
    let content = [0u8; 32];
    let bid = Bid::from_parts(base, content);

    c.benchmark_group("bid_append_wire")
        .throughput(Throughput::Elements(1))
        .bench_function("bid_append_wire", |b| {
            let mut out = Vec::with_capacity(81);
            b.iter(|| {
                out.clear();
                bid.append_wire(black_box(&mut out)).unwrap();
                black_box(out.len())
            });
        });
}

fn bench_bid_parse_wire(c: &mut Criterion) {
    let base = Snid::new_fast();
    let content = [0u8; 32];
    let bid = Bid::from_parts(base, content);
    let wire = bid.wire().unwrap();

    c.benchmark_group("bid_parse_wire")
        .throughput(Throughput::Elements(1))
        .bench_function("bid_parse_wire", |b| {
            b.iter(|| black_box(Bid::parse_wire(black_box(&wire)).unwrap()));
        });
}

fn bench_eid_new(c: &mut Criterion) {
    let unix_millis: u64 = 1700000000123;
    let counter: u16 = 12345;

    c.benchmark_group("eid_new")
        .throughput(Throughput::Elements(1))
        .bench_function("eid_new", |b| {
            b.iter(|| black_box(Eid::from_parts(black_box(unix_millis), black_box(counter))));
        });
}

fn bench_eid_to_bytes(c: &mut Criterion) {
    let unix_millis: u64 = 1700000000123;
    let counter: u16 = 12345;
    let eid = Eid::from_parts(unix_millis, counter);

    c.benchmark_group("eid_to_bytes")
        .throughput(Throughput::Elements(1))
        .bench_function("eid_to_bytes", |b| {
            b.iter(|| black_box(eid.to_bytes()));
        });
}

fn bench_eid_counter(c: &mut Criterion) {
    let unix_millis: u64 = 1700000000123;
    let counter: u16 = 12345;
    let eid = Eid::from_parts(unix_millis, counter);

    c.benchmark_group("eid_counter")
        .throughput(Throughput::Elements(1))
        .bench_function("eid_counter", |b| {
            b.iter(|| black_box(eid.counter()));
        });
}

fn bench_eid_timestamp(c: &mut Criterion) {
    let unix_millis: u64 = 1700000000123;
    let counter: u16 = 12345;
    let eid = Eid::from_parts(unix_millis, counter);

    c.benchmark_group("eid_timestamp")
        .throughput(Throughput::Elements(1))
        .bench_function("eid_timestamp", |b| {
            b.iter(|| black_box(eid.timestamp_millis()));
        });
}

fn bench_eid_batch_100(c: &mut Criterion) {
    let unix_millis: u64 = 1700000000123;
    let counter: u16 = 12345;

    c.benchmark_group("eid_batch_100")
        .throughput(Throughput::Elements(100))
        .bench_function("eid_batch_100", |b| {
            b.iter(|| {
                for _ in 0..100 {
                    black_box(Eid::from_parts(unix_millis, counter));
                }
            });
        });
}

fn bench_eid_batch_1000(c: &mut Criterion) {
    let unix_millis: u64 = 1700000000123;
    let counter: u16 = 12345;

    c.benchmark_group("eid_batch_1000")
        .throughput(Throughput::Elements(1000))
        .bench_function("eid_batch_1000", |b| {
            b.iter(|| {
                for _ in 0..1000 {
                    black_box(Eid::from_parts(unix_millis, counter));
                }
            });
        });
}

fn bench_traceparent_write(c: &mut Criterion) {
    let trace = TraceId::new();
    let span = [0x42u8; 8];

    c.benchmark_group("traceparent_write")
        .throughput(Throughput::Elements(1))
        .bench_function("traceparent_write", |b| {
            b.iter(|| {
                let mut out = [0u8; 55];
                black_box(
                    trace
                        .write_traceparent(black_box(span), black_box(&mut out))
                        .len(),
                )
            });
        });
}

#[cfg(not(feature = "crypto"))]
criterion_group! {
    name = extended_families;
    config = Criterion::default();
    targets =
        bench_nid_new,
        bench_nid_hamming,
        bench_nid_batch_100,
        bench_nid_batch_1000,
        bench_bid_new,
        bench_bid_wire,
        bench_bid_write_wire,
        bench_bid_append_wire,
        bench_bid_parse_wire,
        bench_eid_new,
        bench_eid_to_bytes,
        bench_eid_counter,
        bench_eid_timestamp,
        bench_eid_batch_100,
        bench_eid_batch_1000,
        bench_traceparent_write
}

#[cfg(feature = "crypto")]
criterion_group! {
    name = extended_families;
    config = Criterion::default();
    targets =
        bench_nid_new,
        bench_nid_hamming,
        bench_nid_batch_100,
        bench_nid_batch_1000,
        bench_lid_new,
        bench_lid_head,
        bench_kid_verify,
        bench_grant_new,
        bench_grant_verify,
        bench_bid_new,
        bench_bid_wire,
        bench_bid_write_wire,
        bench_bid_append_wire,
        bench_bid_parse_wire,
        bench_eid_new,
        bench_eid_to_bytes,
        bench_eid_counter,
        bench_eid_timestamp,
        bench_eid_batch_100,
        bench_eid_batch_1000,
        bench_traceparent_write
}

criterion_main!(extended_families);
