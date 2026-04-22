use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use snid::{Bid, Eid, Lid, Nid, Snid, TurboStreamer};

fn bench_snid_new_fast(c: &mut Criterion) {
    c.benchmark_group("snid_new_fast")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_new_fast", |b| {
            b.iter(|| black_box(Snid::new_fast()));
        });
}

fn bench_snid_new_safe(c: &mut Criterion) {
    c.benchmark_group("snid_new_safe")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_new_safe", |b| {
            b.iter(|| black_box(Snid::new_safe()));
        });
}

fn bench_snid_to_wire(c: &mut Criterion) {
    let id = Snid::new_fast();
    c.benchmark_group("snid_to_wire_mat")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_to_wire_mat", |b| {
            b.iter(|| black_box(id.to_wire(black_box("MAT")).unwrap()));
        });
}

fn bench_snid_write_wire(c: &mut Criterion) {
    let id = Snid::new_fast();
    c.benchmark_group("snid_write_wire_mat")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_write_wire_mat", |b| {
            b.iter(|| {
                let mut out = [0u8; 28];
                black_box(
                    id.write_wire(black_box("MAT"), black_box(&mut out))
                        .unwrap()
                        .len(),
                )
            });
        });
}

fn bench_snid_append_wire(c: &mut Criterion) {
    let id = Snid::new_fast();
    c.benchmark_group("snid_append_wire_mat")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_append_wire_mat", |b| {
            let mut out = Vec::with_capacity(28);
            b.iter(|| {
                out.clear();
                id.append_wire(black_box("MAT"), black_box(&mut out))
                    .unwrap();
                black_box(out.len())
            });
        });
}

fn bench_snid_parse_wire(c: &mut Criterion) {
    let id = Snid::new_fast();
    let wire = id.to_wire("MAT").unwrap();
    c.benchmark_group("snid_parse_wire")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_parse_wire", |b| {
            b.iter(|| black_box(Snid::parse_wire(black_box(&wire)).unwrap()));
        });
}

fn bench_snid_batch_1000(c: &mut Criterion) {
    c.benchmark_group("snid_batch_1000")
        .throughput(Throughput::Elements(1000))
        .bench_function("snid_batch_1000", |b| {
            b.iter(|| black_box(Snid::batch(black_box(1000))));
        });
}

fn bench_snid_fill_slice_1000(c: &mut Criterion) {
    let mut ids = vec![Snid::from_bytes([0u8; 16]); 1000];
    c.benchmark_group("snid_fill_slice_1000")
        .throughput(Throughput::Elements(1000))
        .bench_function("snid_fill_slice_1000", |b| {
            b.iter(|| {
                Snid::fill_slice(black_box(&mut ids));
                black_box(ids.as_ptr())
            });
        });
}

fn bench_snid_fill_bytes_1000(c: &mut Criterion) {
    let mut ids = vec![0u8; 16_000];
    c.benchmark_group("snid_fill_bytes_1000")
        .throughput(Throughput::Elements(1000))
        .bench_function("snid_fill_bytes_1000", |b| {
            b.iter(|| black_box(Snid::fill_bytes(black_box(&mut ids))));
        });
}

fn bench_snid_append_binary_batch_1000(c: &mut Criterion) {
    let mut out = Vec::with_capacity(16_000);
    c.benchmark_group("snid_append_binary_batch_1000")
        .throughput(Throughput::Elements(1000))
        .bench_function("snid_append_binary_batch_1000", |b| {
            b.iter(|| {
                out.clear();
                Snid::append_binary_batch(black_box(1000), black_box(&mut out));
                black_box(out.len())
            });
        });
}

fn bench_turbo_streamer_next(c: &mut Criterion) {
    let mut streamer = TurboStreamer::new(4096);
    c.benchmark_group("turbo_streamer_next")
        .throughput(Throughput::Elements(1))
        .bench_function("turbo_streamer_next", |b| {
            b.iter(|| black_box(streamer.next_id()));
        });
}

fn bench_turbo_streamer_refill_4096(c: &mut Criterion) {
    let mut streamer = TurboStreamer::new(4096);
    c.benchmark_group("turbo_streamer_refill_4096")
        .throughput(Throughput::Elements(4096))
        .bench_function("turbo_streamer_refill_4096", |b| {
            b.iter(|| {
                streamer.refill();
                black_box(streamer.buffer().as_ptr())
            });
        });
}

fn bench_snid_concurrent_4x1000(c: &mut Criterion) {
    c.benchmark_group("snid_concurrent_4x1000")
        .throughput(Throughput::Elements(4000))
        .bench_function("snid_concurrent_4x1000", |b| {
            b.iter(|| {
                let handles: Vec<_> = (0..4)
                    .map(|_| {
                        std::thread::spawn(|| {
                            for _ in 0..1000 {
                                black_box(Snid::new_fast());
                            }
                        })
                    })
                    .collect();
                for handle in handles {
                    handle.join().unwrap();
                }
            });
        });
}

fn bench_nid_hamming(c: &mut Criterion) {
    let head = Snid::new_fast();
    let left = Nid::from_parts(head, [0xAA; 16]);
    let right = Nid::from_parts(head, [0x55; 16]);
    c.benchmark_group("nid_hamming_distance")
        .throughput(Throughput::Elements(1))
        .bench_function("nid_hamming_distance", |b| {
            b.iter(|| black_box(left.hamming_distance(black_box(&right))));
        });
}

fn bench_bid_wire(c: &mut Criterion) {
    let bid = Bid::from_parts(Snid::new_fast(), [0x42; 32]);
    c.benchmark_group("bid_wire")
        .throughput(Throughput::Elements(1))
        .bench_function("bid_wire", |b| {
            b.iter(|| black_box(bid.wire().unwrap()));
        });
}

fn bench_eid_from_parts(c: &mut Criterion) {
    c.benchmark_group("eid_from_parts")
        .throughput(Throughput::Elements(1))
        .bench_function("eid_from_parts", |b| {
            b.iter(|| black_box(Eid::from_parts(black_box(1_700_000_000_123), black_box(42))));
        });
}

fn bench_lid_from_parts(c: &mut Criterion) {
    let head = Snid::new_fast();
    let prev = [0x11; 32];
    let payload = b"transaction_data";
    let key = b"secret";
    c.benchmark_group("lid_from_parts")
        .throughput(Throughput::Elements(1))
        .bench_function("lid_from_parts", |b| {
            b.iter(|| black_box(Lid::from_parts(head, prev, payload, key).unwrap()));
        });
}

fn bench_snid_to_uuid_string(c: &mut Criterion) {
    let id = Snid::new_fast();
    c.benchmark_group("snid_to_uuid_string")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_to_uuid_string", |b| {
            b.iter(|| black_box(id.to_uuid_string()));
        });
}

fn bench_snid_write_uuid_string(c: &mut Criterion) {
    let id = Snid::new_fast();
    c.benchmark_group("snid_write_uuid_string")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_write_uuid_string", |b| {
            b.iter(|| {
                let mut out = [0u8; 36];
                id.write_uuid_string(black_box(&mut out));
                black_box(out)
            });
        });
}

fn bench_snid_write_base32(c: &mut Criterion) {
    let id = Snid::new_fast();
    c.benchmark_group("snid_write_base32")
        .throughput(Throughput::Elements(1))
        .bench_function("snid_write_base32", |b| {
            b.iter(|| {
                let mut out = [0u8; 27];
                black_box(id.write_base32(black_box(&mut out)).len())
            });
        });
}

criterion_group!(
    benches,
    bench_snid_new_fast,
    bench_snid_new_safe,
    bench_snid_to_wire,
    bench_snid_write_wire,
    bench_snid_append_wire,
    bench_snid_parse_wire,
    bench_snid_batch_1000,
    bench_snid_fill_slice_1000,
    bench_snid_fill_bytes_1000,
    bench_snid_append_binary_batch_1000,
    bench_turbo_streamer_next,
    bench_turbo_streamer_refill_4096,
    bench_snid_concurrent_4x1000,
    bench_nid_hamming,
    bench_bid_wire,
    bench_eid_from_parts,
    bench_lid_from_parts,
    bench_snid_to_uuid_string,
    bench_snid_write_uuid_string,
    bench_snid_write_base32
);
criterion_main!(benches);
