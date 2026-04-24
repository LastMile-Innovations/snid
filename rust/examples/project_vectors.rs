use snid::{Bid, Eid, Kid, Lid, Nid, Snid, VectorFile, Wid, Xid};

fn hex_encode(bytes: &[u8]) -> String {
    const HEX: &[u8; 16] = b"0123456789abcdef";
    let mut out = Vec::with_capacity(bytes.len() * 2);
    for &byte in bytes {
        out.push(HEX[(byte >> 4) as usize]);
        out.push(HEX[(byte & 0x0f) as usize]);
    }
    String::from_utf8(out).expect("hex is utf-8")
}

fn hex_decode<const N: usize>(value: &str) -> [u8; N] {
    assert_eq!(value.len(), N * 2);
    let mut out = [0u8; N];
    let bytes = value.as_bytes();
    for idx in 0..N {
        out[idx] = (hex_value(bytes[idx * 2]) << 4) | hex_value(bytes[idx * 2 + 1]);
    }
    out
}

fn hex_decode_vec(value: &str) -> Vec<u8> {
    assert_eq!(value.len() % 2, 0);
    let mut out = vec![0u8; value.len() / 2];
    let bytes = value.as_bytes();
    for idx in 0..out.len() {
        out[idx] = (hex_value(bytes[idx * 2]) << 4) | hex_value(bytes[idx * 2 + 1]);
    }
    out
}

fn hex_value(byte: u8) -> u8 {
    match byte {
        b'0'..=b'9' => byte - b'0',
        b'a'..=b'f' => byte - b'a' + 10,
        b'A'..=b'F' => byte - b'A' + 10,
        _ => panic!("invalid hex"),
    }
}

fn main() {
    let raw = std::fs::read_to_string("../conformance/vectors.json")
        .or_else(|_| std::fs::read_to_string("conformance/vectors.json"))
        .expect("vectors.json");
    let vectors: VectorFile = serde_json::from_str(&raw).expect("parse vectors");

    let core = vectors
        .core
        .iter()
        .map(|case| {
            let snid = Snid::from_hex(&case.bytes_hex).unwrap();
            let llm = snid.to_llm_format(&case.atom).unwrap();
            serde_json::json!({
                "bytes_hex": hex_encode(&snid.0),
                "wire": snid.to_wire(&case.atom).unwrap(),
                "tensor_hi": snid.to_tensor_words().0,
                "tensor_lo": snid.to_tensor_words().1,
                "atom": llm.atom,
                "timestamp_millis": llm.timestamp_millis,
                "machine_or_shard": llm.machine_or_shard,
                "sequence": llm.sequence,
            })
        })
        .collect::<Vec<_>>();

    let sgid = Snid::from_hex(&vectors.spatial.bytes_hex).unwrap();
    let head = Snid::from_hex(&vectors.neural.head_hex).unwrap();
    let semantic = hex_decode::<16>(&vectors.neural.semantic_hex);
    let nid = Nid::from_parts(head, semantic);
    let ledger_head = Snid::from_hex(&vectors.ledger.head_hex).unwrap();
    let prev = hex_decode::<32>(&vectors.ledger.prev_hex);
    let payload = hex_decode_vec(&vectors.ledger.payload_hex);
    let key = hex_decode_vec(&vectors.ledger.key_hex);
    let lid = Lid::from_parts(ledger_head, prev, &payload, &key).unwrap();
    let world_head = Snid::from_hex(&vectors.world.head_hex).unwrap();
    let scenario = hex_decode::<16>(&vectors.world.scenario_hex);
    let wid = Wid::from_parts(world_head, scenario);
    let edge_head = Snid::from_hex(&vectors.edge.head_hex).unwrap();
    let edge_hash = hex_decode::<16>(&vectors.edge.edge_hex);
    let xid = Xid::from_parts(edge_head, edge_hash);
    let kid_head = Snid::from_hex(&vectors.capability.head_hex).unwrap();
    let actor = Snid::from_hex(&vectors.capability.actor_hex).unwrap();
    let resource = hex_decode_vec(&vectors.capability.resource_hex);
    let capability = hex_decode_vec(&vectors.capability.capability_hex);
    let kid_key = hex_decode_vec(&vectors.capability.key_hex);
    let kid = Kid::from_parts(kid_head, actor, &resource, &capability, &kid_key).unwrap();
    let eid_bytes = hex_decode::<8>(&vectors.ephemeral.bytes_hex);
    let eid = Eid(u64::from_be_bytes(eid_bytes));
    let topology = Snid::from_hex(&vectors.bid.topology_hex).unwrap();
    let content = hex_decode::<32>(&vectors.bid.content_hex);
    let bid = Bid::from_parts(topology, content);

    let out = serde_json::json!({
        "core": core,
        "spatial": {
            "bytes_hex": hex_encode(&sgid.0),
            "wire": sgid.to_wire(&vectors.spatial.atom).unwrap(),
            "h3_cell_hex": format!("{:x}", sgid.h3_cell().unwrap()),
        },
        "neural": {
            "bytes_hex": hex_encode(&nid.0),
            "hamming_to_zero": nid.hamming_distance(&Nid([0u8; 32])),
        },
        "ledger": {
            "bytes_hex": hex_encode(&lid.0),
        },
        "world": {
            "bytes_hex": hex_encode(&wid.0),
            "tensor_words": wid.to_tensor256_words(),
        },
        "edge": {
            "bytes_hex": hex_encode(&xid.0),
            "tensor_words": xid.to_tensor256_words(),
        },
        "capability": {
            "bytes_hex": hex_encode(&kid.0),
            "tensor_words": kid.to_tensor256_words(),
            "verified": kid.verify(actor, &resource, &capability, &kid_key),
        },
        "ephemeral": {
            "bytes_hex": hex_encode(&eid.to_bytes()),
            "timestamp_millis": eid.timestamp_millis(),
            "counter": eid.counter(),
        },
        "bid": {
            "wire": bid.wire().unwrap(),
            "r2_key": bid.r2_key(),
            "neo4j_id": bid.topology.to_wire("MAT").unwrap(),
        },
    });

    println!("{}", serde_json::to_string_pretty(&out).unwrap());
}
