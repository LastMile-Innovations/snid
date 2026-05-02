use criterion::{Criterion, criterion_group, criterion_main};
use snhash::{Options, Profile, hash_bytes, hash_bytes_profile, hash_bytes_with_options};
use std::hint::black_box;

fn benchmarks(c: &mut Criterion) {
    let fast = vec![b'a'; 4096];
    c.bench_function("hash_bytes_fastcas_4k", |b| {
        b.iter(|| hash_bytes(black_box(&fast)).unwrap())
    });

    let cache = vec![b'c'; 40 * 1024];
    c.bench_function("hash_bytes_cache_40k", |b| {
        b.iter(|| hash_bytes_profile(black_box(&cache), Profile::Cache).unwrap())
    });

    let evidence = vec![b'e'; 1024 * 1024];
    c.bench_function("hash_bytes_evidence_1m", |b| {
        b.iter(|| hash_bytes_profile(black_box(&evidence), Profile::Evidence).unwrap())
    });

    c.bench_function("chunked_hash_1m", |b| {
        b.iter(|| {
            hash_bytes_with_options(
                black_box(&evidence),
                Options::profile(Profile::Evidence).chunk_size(64 * 1024),
            )
            .unwrap()
        })
    });

    let manifest = hash_bytes_profile(b"manifest proof root", Profile::Evidence).unwrap();
    c.bench_function("proof_root", |b| b.iter(|| manifest.proof_root().unwrap()));
    c.bench_function("hash_id_from_manifest", |b| {
        b.iter(|| manifest.hash_id().unwrap())
    });
}

criterion_group!(benches, benchmarks);
criterion_main!(benches);
