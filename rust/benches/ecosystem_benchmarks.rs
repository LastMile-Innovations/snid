use criterion::{black_box, criterion_group, criterion_main, Criterion};
use snid::Snid;
use uuid::Uuid;
use ulid::Ulid;
use ksuid::Ksuid;
use xid::Xid;
use sonyflake::Sonyflake;
use cuid::Cuid;

// Industry Standard Baseline: UUIDv7
fn bench_uuidv7_new(c: &mut Criterion) {
    c.bench_function("uuidv7_new", |b| {
        b.iter(|| black_box(Uuid::new_v7()));
    });
}

// SNID Baseline
fn bench_snid_new(c: &mut Criterion) {
    c.bench_function("snid_new", |b| {
        b.iter(|| black_box(Snid::new()));
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
        b.iter(|| black_box(Ksuid::new()));
    });
}

fn bench_xid_new(c: &mut Criterion) {
    c.bench_function("xid_new", |b| {
        b.iter(|| black_box(Xid::new()));
    });
}

fn bench_sonyflake_new(c: &mut Criterion) {
    let sf = Sonyflake::new();
    c.bench_function("sonyflake_new", |b| {
        b.iter(|| black_box(sf.next_id()));
    });
}

fn bench_cuid_new(c: &mut Criterion) {
    c.bench_function("cuid_new", |b| {
        b.iter(|| black_box(Cuid::new()));
    });
}

criterion_group!(
    ecosystem_benches,
    bench_uuidv7_new,
    bench_snid_new,
    bench_uuid_new,
    bench_ulid_new,
    bench_ksuid_new,
    bench_xid_new,
    bench_sonyflake_new,
    bench_cuid_new
);
criterion_main!(ecosystem_benches);
