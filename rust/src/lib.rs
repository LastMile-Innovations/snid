//! # SNID - Polyglot Sortable Identifier Protocol
//!
//! SNID provides 128-bit time-ordered identifiers compatible with UUID v7,
//! plus extended families for spatial, neural, ledger, and capability use cases.
//!
//! ## Basic Usage
//!
//! ```no_run
//! use snid::Snid;
//!
//! let id = Snid::new_fast();
//! let wire = id.to_wire("MAT").unwrap();
//! let (parsed, atom) = Snid::parse_wire(&wire).unwrap();
//! ```
//!
//! ## Extended Identifier Families
//!
//! - **Snid** - Core 128-bit time-ordered ID (UUID v7 compatible)
//! - **Sgid** - Spatial ID with H3 geospatial encoding
//! - **Nid** - Neural ID with semantic tail for vector search
//! - **Lid** - Ledger ID with HMAC verification tail
//! - **Wid** - World/scenario ID for simulation isolation
//! - **Xid** - Edge ID for relationship identity
//! - **Kid** - Capability ID with MAC-based verification
//! - **Eid** - Ephemeral 64-bit ID
//! - **Bid** - Content-addressable ID (topology + content hash)
//! - **Akid** - Dual-part public+secret credentials (access keys)
//! - **GrantId** - Tamper-proof capability grants with expiration
//! - **ScopeId** - Logical group routing (tenant/region)
//! - **ShardId** - Fixed shard key routing
//! - **AliasId** - Human-readable aliases
//! - **ShortId** - 64-bit compact identifiers
//! - **TraceId** - W3C/OpenTelemetry compatible trace IDs
//!
//! ## Features
//!
//! - `data` - Enable serde serialization and JSON support
//!
//! For protocol specification, see <https://github.com/LastMile-Innovations/snid/blob/main/docs/SPEC.md>

// Module declarations
mod akid;
mod compact;
mod conformance;
mod core;
mod encoding;
mod error;
mod generator;
mod helpers;
mod neural;
mod projections;
mod routing;
mod spatial;
mod types;

// Public re-exports
pub use akid::Akid;
pub use compact::{ShortId, TraceId};
#[cfg(feature = "data")]
pub use conformance::VectorFile;
pub use core::Snid;
pub use encoding::{decode_fixed64_pair, encode_fixed64_pair};
pub use error::Error;
pub use generator::TurboStreamer;
pub use routing::{AliasId, GrantId, ScopeId, ShardId};
pub use types::{Bid, Eid, Kid, Lid, Nid, Wid, Xid};
