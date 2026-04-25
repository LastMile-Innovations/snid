//! Conformance vector loading and validation tests.

#[allow(unused_imports)]
use crate::core::Snid;
#[allow(unused_imports)]
use crate::error::Error;
#[allow(unused_imports)]
#[cfg(feature = "data")]
use crate::types::{Bid, Eid, Kid, Lid, Nid, Wid, Xid};

#[cfg(feature = "data")]
use serde::Deserialize;

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct VectorFile {
    pub core: Vec<CoreVector>,
    pub spatial: SpatialVector,
    pub neural: NeuralVector,
    pub ledger: LedgerVector,
    pub world: WorldVector,
    pub edge: EdgeVector,
    pub capability: CapabilityVector,
    pub ephemeral: EIDVector,
    pub bid: BIDVector,
    pub compatibility: CompatibilityVector,
    pub uuidv7: UUIDv7Vector,
    pub negative: NegativeVector,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct CoreVector {
    pub atom: String,
    pub bytes_hex: String,
    pub wire: String,
    pub underscore_wire: String,
    pub timestamp_millis: i64,
    pub tensor_hi: i64,
    pub tensor_lo: i64,
    pub llm_format: LLMFormatVector,
    pub llm_format_v2: LLMFormatVectorV2,
    pub time_bin_hour: i64,
    pub ghosted_bytes_hex: String,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct LLMFormatVector {
    pub atom: String,
    pub timestamp_millis: i64,
    pub machine_or_shard: u32,
    pub sequence: u16,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct LLMFormatVectorV2 {
    pub kind: String,
    pub atom: String,
    pub timestamp_millis: Option<i64>,
    pub spatial_anchor: Option<u64>,
    pub machine_or_shard: Option<u32>,
    pub sequence: Option<u16>,
    pub ghosted: bool,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct SpatialVector {
    pub atom: String,
    pub bytes_hex: String,
    pub wire: String,
    pub h3_cell_hex: String,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct NeuralVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub semantic_hex: String,
    pub hamming_to_zero: i32,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct LedgerVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub prev_hex: String,
    pub payload_hex: String,
    pub key_hex: String,
    pub blake3_hex: String,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct WorldVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub scenario_hex: String,
    pub tensor_words: [i64; 4],
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct EdgeVector {
    pub head_hex: String,
    pub bytes_hex: String,
    pub edge_hex: String,
    pub tensor_words: [i64; 4],
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct CapabilityVector {
    pub head_hex: String,
    pub actor_hex: String,
    pub resource_hex: String,
    pub capability_hex: String,
    pub key_hex: String,
    pub bytes_hex: String,
    pub tensor_words: [i64; 4],
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct EIDVector {
    pub bytes_hex: String,
    pub timestamp_millis: u64,
    pub counter: u16,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct BIDVector {
    pub topology_hex: String,
    pub content_hex: String,
    pub wire: String,
    pub r2_key: String,
    pub neo4j_id: String,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct CompatibilityVector {
    pub bytes_hex: String,
    pub wire: String,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct UUIDv7Vector {
    pub bytes_hex: String,
    pub uuid_string: String,
    pub timestamp_millis: i64,
    pub version: i32,
    pub variant: i32,
}

#[cfg(feature = "data")]
#[derive(Deserialize)]
pub struct NegativeVector {
    pub invalid_atom_wire: String,
    pub invalid_binary_hex: String,
    pub invalid_wire_checksum: String,
    pub invalid_adapter_hex: String,
}

#[cfg(test)]
mod tests {
    #[allow(unused_imports)]
    use super::*;
    #[allow(unused_imports)]
    use crate::core::Snid;
    #[cfg(feature = "data")]
    #[allow(unused_imports)]
    use crate::helpers::{hex_decode_to, hex_decode_vec, hex_encode_fast};
    #[allow(unused_imports)]
    #[cfg(feature = "data")]
    use crate::types::{Bid, Eid, Kid, Lid, Nid, Wid, Xid};

    #[cfg(feature = "data")]
    #[test]
    fn validates_vectors() {
        let raw = std::fs::read_to_string("../conformance/vectors.json")
            .or_else(|_| std::fs::read_to_string("../SNID/conformance/vectors.json"))
            .expect("vectors.json");
        let vectors: VectorFile = serde_json::from_str(&raw).expect("parse vectors");

        for case in vectors.core {
            let expected = Snid::from_hex(&case.bytes_hex).unwrap();
            let (parsed, atom) = Snid::parse_wire(&case.wire).unwrap();
            assert_eq!(parsed, expected);
            assert_eq!(atom, case.atom);
            assert_eq!(expected.to_wire(&case.atom).unwrap(), case.wire);
            assert_eq!(expected.timestamp_millis(), case.timestamp_millis);
            assert_eq!(expected.to_tensor_words(), (case.tensor_hi, case.tensor_lo));
            let llm = expected.to_llm_format(&case.atom).unwrap();
            assert_eq!(llm.atom, case.llm_format.atom);
            assert_eq!(llm.timestamp_millis, case.llm_format.timestamp_millis);
            assert_eq!(llm.machine_or_shard, case.llm_format.machine_or_shard);
            assert_eq!(llm.sequence, case.llm_format.sequence);
            let llm_v2 = expected.to_llm_format_v2(&case.atom).unwrap();
            assert_eq!(llm_v2.kind, case.llm_format_v2.kind);
            assert_eq!(llm_v2.atom, case.llm_format_v2.atom);
            assert_eq!(llm_v2.timestamp_millis, case.llm_format_v2.timestamp_millis);
            assert_eq!(llm_v2.machine_or_shard, case.llm_format_v2.machine_or_shard);
            assert_eq!(llm_v2.sequence, case.llm_format_v2.sequence);
            assert_eq!(llm_v2.ghosted, case.llm_format_v2.ghosted);
            assert_eq!(expected.time_bin(3_600_000), case.time_bin_hour);
            assert_eq!(
                hex_encode_fast(&expected.with_ghost_bit(true).0),
                case.ghosted_bytes_hex
            );
            let (underscore, _) = Snid::parse_wire(&case.underscore_wire).unwrap();
            assert_eq!(underscore, expected);
        }

        let sgid = Snid::from_hex(&vectors.spatial.bytes_hex).unwrap();
        assert_eq!(
            sgid.to_wire(&vectors.spatial.atom).unwrap(),
            vectors.spatial.wire
        );
        assert_eq!(
            format!("{:x}", sgid.h3_cell().unwrap()),
            vectors.spatial.h3_cell_hex
        );

        let head = Snid::from_hex(&vectors.neural.head_hex).unwrap();
        let mut semantic = [0u8; 16];
        hex_decode_to(&vectors.neural.semantic_hex, &mut semantic).unwrap();
        let nid = Nid::from_parts(head, semantic);
        assert_eq!(hex_encode_fast(&nid.0), vectors.neural.bytes_hex);
        assert_eq!(
            nid.hamming_distance(&Nid([0u8; 32])) as i32,
            vectors.neural.hamming_to_zero
        );

        let head = Snid::from_hex(&vectors.ledger.head_hex).unwrap();
        let mut prev = [0u8; 32];
        hex_decode_to(&vectors.ledger.prev_hex, &mut prev).unwrap();
        let payload = hex_decode_vec(&vectors.ledger.payload_hex).unwrap();
        let key = hex_decode_vec(&vectors.ledger.key_hex).unwrap();
        let lid = Lid::from_parts(head, prev, &payload, &key).unwrap();
        assert_eq!(hex_encode_fast(&lid.0), vectors.ledger.bytes_hex);

        let world_head = Snid::from_hex(&vectors.world.head_hex).unwrap();
        let mut scenario = [0u8; 16];
        hex_decode_to(&vectors.world.scenario_hex, &mut scenario).unwrap();
        let wid = Wid::from_parts(world_head, scenario);
        assert_eq!(hex_encode_fast(&wid.0), vectors.world.bytes_hex);
        assert_eq!(
            wid.to_tensor256_words(),
            (
                vectors.world.tensor_words[0],
                vectors.world.tensor_words[1],
                vectors.world.tensor_words[2],
                vectors.world.tensor_words[3]
            )
        );

        let edge_head = Snid::from_hex(&vectors.edge.head_hex).unwrap();
        let mut edge_hash = [0u8; 16];
        hex_decode_to(&vectors.edge.edge_hex, &mut edge_hash).unwrap();
        let xid = Xid::from_parts(edge_head, edge_hash);
        assert_eq!(hex_encode_fast(&xid.0), vectors.edge.bytes_hex);
        assert_eq!(
            xid.to_tensor256_words(),
            (
                vectors.edge.tensor_words[0],
                vectors.edge.tensor_words[1],
                vectors.edge.tensor_words[2],
                vectors.edge.tensor_words[3]
            )
        );

        let kid_head = Snid::from_hex(&vectors.capability.head_hex).unwrap();
        let actor = Snid::from_hex(&vectors.capability.actor_hex).unwrap();
        let resource = hex_decode_vec(&vectors.capability.resource_hex).unwrap();
        let capability = hex_decode_vec(&vectors.capability.capability_hex).unwrap();
        let key = hex_decode_vec(&vectors.capability.key_hex).unwrap();
        let kid = Kid::from_parts(kid_head, actor, &resource, &capability, &key).unwrap();
        assert_eq!(hex_encode_fast(&kid.0), vectors.capability.bytes_hex);
        assert!(kid.verify(actor, &resource, &capability, &key));
        assert_eq!(
            kid.to_tensor256_words(),
            (
                vectors.capability.tensor_words[0],
                vectors.capability.tensor_words[1],
                vectors.capability.tensor_words[2],
                vectors.capability.tensor_words[3]
            )
        );

        let mut eid_bytes = [0u8; 8];
        hex_decode_to(&vectors.ephemeral.bytes_hex, &mut eid_bytes).unwrap();
        let eid = Eid(u64::from_be_bytes(eid_bytes));
        assert_eq!(eid.timestamp_millis(), vectors.ephemeral.timestamp_millis);
        assert_eq!(eid.counter(), vectors.ephemeral.counter);

        let topology = Snid::from_hex(&vectors.bid.topology_hex).unwrap();
        let mut content = [0u8; 32];
        hex_decode_to(&vectors.bid.content_hex, &mut content).unwrap();
        let bid = Bid::from_parts(topology, content);
        assert_eq!(bid.wire().unwrap(), vectors.bid.wire);
        assert_eq!(bid.r2_key(), vectors.bid.r2_key);
        assert_eq!(bid.topology.to_wire("MAT").unwrap(), vectors.bid.neo4j_id);
        assert_eq!(Bid::parse_wire(&vectors.bid.wire).unwrap(), bid);

        let compat = Snid::from_hex(&vectors.compatibility.bytes_hex).unwrap();
        assert_eq!(compat.to_wire("MAT").unwrap(), vectors.compatibility.wire);
        let (parsed, _) = Snid::parse_wire(&vectors.compatibility.wire).unwrap();
        assert_eq!(parsed, compat);

        let uuidv7 = Snid::from_hex(&vectors.uuidv7.bytes_hex).unwrap();
        let uuid_str = format!(
            "{:02x}{:02x}{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}-{:02x}{:02x}{:02x}{:02x}{:02x}{:02x}",
            uuidv7.0[0],
            uuidv7.0[1],
            uuidv7.0[2],
            uuidv7.0[3],
            uuidv7.0[4],
            uuidv7.0[5],
            uuidv7.0[6],
            uuidv7.0[7],
            uuidv7.0[8],
            uuidv7.0[9],
            uuidv7.0[10],
            uuidv7.0[11],
            uuidv7.0[12],
            uuidv7.0[13],
            uuidv7.0[14],
            uuidv7.0[15]
        );
        assert_eq!(uuid_str, vectors.uuidv7.uuid_string);
        assert_eq!(uuidv7.timestamp_millis(), vectors.uuidv7.timestamp_millis);
        let version = uuidv7.0[6] >> 4;
        assert_eq!(version as i32, vectors.uuidv7.version);
        let variant = (uuidv7.0[8] >> 6) & 0b11;
        assert_eq!(variant as i32, vectors.uuidv7.variant);

        assert!(Snid::parse_wire(&vectors.negative.invalid_atom_wire).is_err());
        assert!(Snid::from_hex(&vectors.negative.invalid_binary_hex).is_err());
        assert!(Snid::parse_wire(&vectors.negative.invalid_wire_checksum).is_err());
        assert!(Snid::from_hex(&vectors.negative.invalid_adapter_hex).is_err());
    }
}
