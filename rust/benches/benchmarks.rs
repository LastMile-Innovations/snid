use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use snid::{Bid, Eid, Lid, Nid, Snid};

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

criterion_group!(
    benches,
    bench_snid_new_fast,
    bench_snid_new_safe,
    bench_snid_to_wire,
    bench_nid_hamming,
    bench_bid_wire,
    bench_eid_from_parts,
    bench_lid_from_parts,
    bench_snid_to_uuid_string
);
criterion_main!(benches);
