// Extended Family Benchmarks for SNID
// Benchmarks for NID, LID, BID, EID operations

use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use snid::{Snid, Nid, Lid, Bid, Eid};

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
        bench_bid_new,
        bench_bid_wire,
        bench_bid_parse_wire,
        bench_eid_new,
        bench_eid_to_bytes,
        bench_eid_counter,
        bench_eid_timestamp,
        bench_eid_batch_100,
        bench_eid_batch_1000
}

criterion_main!(extended_families);
