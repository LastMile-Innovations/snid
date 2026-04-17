use snid_core::{Bid, Eid, Kid, Lid, Nid, Snid, VectorFile, Wid, Xid};

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
                "bytes_hex": hex::encode(snid.0),
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
    let semantic: [u8; 16] = hex::decode(&vectors.neural.semantic_hex).unwrap().try_into().unwrap();
    let nid = Nid::from_parts(head, semantic);
    let ledger_head = Snid::from_hex(&vectors.ledger.head_hex).unwrap();
    let prev: [u8; 32] = hex::decode(&vectors.ledger.prev_hex).unwrap().try_into().unwrap();
    let payload = hex::decode(&vectors.ledger.payload_hex).unwrap();
    let key = hex::decode(&vectors.ledger.key_hex).unwrap();
    let lid = Lid::from_parts(ledger_head, prev, &payload, &key).unwrap();
    let world_head = Snid::from_hex(&vectors.world.head_hex).unwrap();
    let scenario: [u8; 16] = hex::decode(&vectors.world.scenario_hex).unwrap().try_into().unwrap();
    let wid = Wid::from_parts(world_head, scenario);
    let edge_head = Snid::from_hex(&vectors.edge.head_hex).unwrap();
    let edge_hash: [u8; 16] = hex::decode(&vectors.edge.edge_hex).unwrap().try_into().unwrap();
    let xid = Xid::from_parts(edge_head, edge_hash);
    let kid_head = Snid::from_hex(&vectors.capability.head_hex).unwrap();
    let actor = Snid::from_hex(&vectors.capability.actor_hex).unwrap();
    let resource = hex::decode(&vectors.capability.resource_hex).unwrap();
    let capability = hex::decode(&vectors.capability.capability_hex).unwrap();
    let kid_key = hex::decode(&vectors.capability.key_hex).unwrap();
    let kid = Kid::from_parts(kid_head, actor, &resource, &capability, &kid_key).unwrap();
    let eid_bytes: [u8; 8] = hex::decode(&vectors.ephemeral.bytes_hex).unwrap().try_into().unwrap();
    let eid = Eid(u64::from_be_bytes(eid_bytes));
    let topology = Snid::from_hex(&vectors.bid.topology_hex).unwrap();
    let content: [u8; 32] = hex::decode(&vectors.bid.content_hex).unwrap().try_into().unwrap();
    let bid = Bid::from_parts(topology, content);

    let out = serde_json::json!({
        "core": core,
        "spatial": {
            "bytes_hex": hex::encode(sgid.0),
            "wire": sgid.to_wire(&vectors.spatial.atom).unwrap(),
            "h3_cell_hex": format!("{:x}", sgid.h3_cell().unwrap()),
        },
        "neural": {
            "bytes_hex": hex::encode(nid.0),
            "hamming_to_zero": nid.hamming_distance(&Nid([0u8; 32])),
        },
        "ledger": {
            "bytes_hex": hex::encode(lid.0),
        },
        "world": {
            "bytes_hex": hex::encode(wid.0),
            "tensor_words": wid.to_tensor256_words(),
        },
        "edge": {
            "bytes_hex": hex::encode(xid.0),
            "tensor_words": xid.to_tensor256_words(),
        },
        "capability": {
            "bytes_hex": hex::encode(kid.0),
            "tensor_words": kid.to_tensor256_words(),
            "verified": kid.verify(actor, &resource, &capability, &kid_key),
        },
        "ephemeral": {
            "bytes_hex": hex::encode(eid.to_bytes()),
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
