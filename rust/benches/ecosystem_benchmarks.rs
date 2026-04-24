use criterion::{criterion_group, criterion_main, Criterion};
use ksuid::Ksuid;
use snid::Snid;
use sonyflake::Sonyflake;
use std::hint::black_box;
use ulid::Ulid;
use uuid::Uuid;

// Industry Standard Baseline: UUIDv7
fn bench_uuidv7_new(c: &mut Criterion) {
    c.bench_function("uuidv7_new", |b| {
        b.iter(|| black_box(uuid::Uuid::now_v7()));
    });
}

// SNID Baseline
fn bench_snid_new(c: &mut Criterion) {
    c.bench_function("snid_new", |b| {
        b.iter(|| black_box(Snid::new_fast()));
    });
}

fn bench_uuid_new(c: &mut Criterion) {
    c.bench_function("uuid_new", |b| {
        b.iter(|| black_box(Uuid::new_v4()));
    });
}

fn bench_ulid_new(c: &mut Criterion) {
    c.bench_function("ulid_new", |b| {
        b.iter(|| black_box(Ulid::new()));
    });
}

fn bench_ksuid_new(c: &mut Criterion) {
    c.bench_function("ksuid_new", |b| {
        b.iter(|| black_box(Ksuid::new(0, [0u8; 16])));
    });
}

fn bench_sonyflake_new(c: &mut Criterion) {
    let sf = Sonyflake::new().expect("Sonyflake init");
    c.bench_function("sonyflake_new", |b| {
        b.iter(|| black_box(sf.next_id().expect("Sonyflake next_id")));
    });
}

criterion_group!(
    ecosystem_benches,
    bench_uuidv7_new,
    bench_snid_new,
    bench_uuid_new,
    bench_ulid_new,
    bench_ksuid_new,
    bench_sonyflake_new
);
criterion_main!(ecosystem_benches);
