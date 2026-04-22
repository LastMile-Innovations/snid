use pyo3::exceptions::PyValueError;
use pyo3::prelude::*;
use pyo3::types::PyBytes;
use pyo3::wrap_pyfunction;
use snid::{Bid, Eid, Kid, Lid, Nid, Snid, Wid, Xid};

fn py_err(err: snid::Error) -> PyErr {
    PyValueError::new_err(format!("{err:?}"))
}

#[pyclass(name = "SNID")]
struct PySnid {
    inner: Snid,
}

#[pymethods]
impl PySnid {
    #[staticmethod]
    fn new_fast() -> Self {
        Self {
            inner: Snid::new_fast(),
        }
    }

    #[staticmethod]
    fn new_safe() -> Self {
        Self {
            inner: Snid::new_safe(),
        }
    }

    #[staticmethod]
    fn from_bytes(data: Vec<u8>) -> PyResult<Self> {
        if data.len() != 16 {
            return Err(PyValueError::new_err("expected 16 bytes"));
        }
        let mut out = [0u8; 16];
        out.copy_from_slice(&data);
        Ok(Self {
            inner: Snid::from_bytes(out),
        })
    }

    #[staticmethod]
    fn parse_wire(value: &str) -> PyResult<(Self, String)> {
        let (inner, atom) = Snid::parse_wire(value).map_err(py_err)?;
        Ok((Self { inner }, atom))
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.to_bytes().to_vec()
    }

    #[staticmethod]
    fn new_uuidv7() -> Self {
        Self {
            inner: Snid::uuidv7(),
        }
    }

    fn to_uuid_string(&self) -> String {
        self.inner.to_uuid_string()
    }

    #[staticmethod]
    fn parse_uuid_string(value: &str) -> PyResult<Self> {
        Ok(Self {
            inner: Snid::from_uuid_string(value).map_err(py_err)?,
        })
    }

    fn to_wire(&self, atom: &str) -> PyResult<String> {
        self.inner.to_wire(atom).map_err(py_err)
    }

    fn to_base32(&self) -> String {
        self.inner.to_base32()
    }

    fn to_tensor(&self) -> (i64, i64) {
        self.inner.to_tensor_words()
    }

    fn timestamp_millis(&self) -> i64 {
        self.inner.timestamp_millis()
    }

    fn to_llm_format(&self, atom: &str) -> PyResult<(String, i64, u32, u16)> {
        let llm = self.inner.to_llm_format(atom).map_err(py_err)?;
        Ok((
            llm.atom,
            llm.timestamp_millis,
            llm.machine_or_shard,
            llm.sequence,
        ))
    }

    fn to_llm_format_v2(
        &self,
        atom: &str,
    ) -> PyResult<(
        String,
        String,
        Option<i64>,
        Option<u64>,
        Option<u32>,
        Option<u16>,
        bool,
    )> {
        let llm = self.inner.to_llm_format_v2(atom).map_err(py_err)?;
        Ok((
            llm.kind,
            llm.atom,
            llm.timestamp_millis,
            llm.spatial_anchor,
            llm.machine_or_shard,
            llm.sequence,
            llm.ghosted,
        ))
    }

    fn time_bin(&self, resolution_millis: i64) -> i64 {
        self.inner.time_bin(resolution_millis)
    }

    fn with_ghost_bit(&self, enabled: bool) -> Self {
        Self {
            inner: self.inner.with_ghost_bit(enabled),
        }
    }

    fn is_ghosted(&self) -> bool {
        self.inner.is_ghosted()
    }

    fn h3_feature_vector(&self) -> Vec<u64> {
        self.inner.h3_feature_vector()
    }

    #[staticmethod]
    fn from_hash_with_timestamp(unix_millis: u64, hash: Vec<u8>) -> Self {
        Self {
            inner: Snid::from_hash_with_timestamp(unix_millis, &hash),
        }
    }
}

#[pyclass(name = "SGID")]
struct PySgid {
    inner: Snid,
}

#[pymethods]
impl PySgid {
    #[staticmethod]
    fn from_parts(cell: u64, entropy: u64) -> Self {
        Self {
            inner: Snid::from_spatial_parts(cell, entropy),
        }
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.to_bytes().to_vec()
    }

    fn to_wire(&self, atom: &str) -> PyResult<String> {
        self.inner.to_wire(atom).map_err(py_err)
    }

    fn h3_cell(&self) -> Option<u64> {
        self.inner.h3_cell()
    }
}

#[pyclass(name = "NID")]
struct PyNid {
    inner: Nid,
}

#[pymethods]
impl PyNid {
    #[staticmethod]
    fn from_parts(head: &PySnid, semantic_hash: Vec<u8>) -> PyResult<Self> {
        if semantic_hash.len() != 16 {
            return Err(PyValueError::new_err("expected 16-byte semantic hash"));
        }
        let mut hash = [0u8; 16];
        hash.copy_from_slice(&semantic_hash);
        Ok(Self {
            inner: Nid::from_parts(head.inner, hash),
        })
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.0.to_vec()
    }

    fn to_tensor(&self) -> (i64, i64, i64, i64) {
        let b = self.inner.0;
        (
            i64::from_be_bytes(b[0..8].try_into().unwrap()),
            i64::from_be_bytes(b[8..16].try_into().unwrap()),
            i64::from_be_bytes(b[16..24].try_into().unwrap()),
            i64::from_be_bytes(b[24..32].try_into().unwrap()),
        )
    }
}

#[pyclass(name = "WID")]
struct PyWid {
    inner: Wid,
}

#[pymethods]
impl PyWid {
    #[staticmethod]
    fn from_parts(head: &PySnid, scenario_hash: Vec<u8>) -> PyResult<Self> {
        if scenario_hash.len() != 16 {
            return Err(PyValueError::new_err("expected 16-byte scenario hash"));
        }
        let mut hash = [0u8; 16];
        hash.copy_from_slice(&scenario_hash);
        Ok(Self {
            inner: Wid::from_parts(head.inner, hash),
        })
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.0.to_vec()
    }

    fn to_tensor(&self) -> (i64, i64, i64, i64) {
        self.inner.to_tensor256_words()
    }
}

#[pyclass(name = "XID")]
struct PyXid {
    inner: Xid,
}

#[pymethods]
impl PyXid {
    #[staticmethod]
    fn from_parts(head: &PySnid, edge_hash: Vec<u8>) -> PyResult<Self> {
        if edge_hash.len() != 16 {
            return Err(PyValueError::new_err("expected 16-byte edge hash"));
        }
        let mut hash = [0u8; 16];
        hash.copy_from_slice(&edge_hash);
        Ok(Self {
            inner: Xid::from_parts(head.inner, hash),
        })
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.0.to_vec()
    }

    fn to_tensor(&self) -> (i64, i64, i64, i64) {
        self.inner.to_tensor256_words()
    }
}

#[pyclass(name = "KID")]
struct PyKid {
    inner: Kid,
}

#[pymethods]
impl PyKid {
    #[staticmethod]
    fn from_parts(
        head: &PySnid,
        actor: &PySnid,
        resource: Vec<u8>,
        capability: Vec<u8>,
        key: Vec<u8>,
    ) -> PyResult<Self> {
        Ok(Self {
            inner: Kid::from_parts(head.inner, actor.inner, &resource, &capability, &key)
                .map_err(py_err)?,
        })
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.0.to_vec()
    }

    fn to_tensor(&self) -> (i64, i64, i64, i64) {
        self.inner.to_tensor256_words()
    }

    fn verify(&self, actor: &PySnid, resource: Vec<u8>, capability: Vec<u8>, key: Vec<u8>) -> bool {
        self.inner.verify(actor.inner, &resource, &capability, &key)
    }
}

#[pyclass(name = "LID")]
struct PyLid {
    inner: Lid,
}

#[pymethods]
impl PyLid {
    #[staticmethod]
    fn from_parts(head: &PySnid, prev: Vec<u8>, payload: Vec<u8>, key: Vec<u8>) -> PyResult<Self> {
        if prev.len() != 32 {
            return Err(PyValueError::new_err("expected 32-byte previous lid"));
        }
        let mut prev_arr = [0u8; 32];
        prev_arr.copy_from_slice(&prev);
        Ok(Self {
            inner: Lid::from_parts(head.inner, prev_arr, &payload, &key).map_err(py_err)?,
        })
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.0.to_vec()
    }
}

#[pyclass(name = "EID")]
struct PyEid {
    inner: Eid,
}

#[pymethods]
impl PyEid {
    #[staticmethod]
    fn from_parts(unix_millis: u64, counter: u16) -> Self {
        Self {
            inner: Eid::from_parts(unix_millis, counter),
        }
    }

    fn to_bytes(&self) -> Vec<u8> {
        self.inner.to_bytes().to_vec()
    }

    fn counter(&self) -> u16 {
        self.inner.counter()
    }

    fn timestamp_millis(&self) -> u64 {
        self.inner.timestamp_millis()
    }
}

#[pyclass(name = "BID")]
struct PyBid {
    inner: Bid,
}

#[pymethods]
impl PyBid {
    #[staticmethod]
    fn from_parts(topology: &PySnid, content: Vec<u8>) -> PyResult<Self> {
        if content.len() != 32 {
            return Err(PyValueError::new_err("expected 32-byte content hash"));
        }
        let mut out = [0u8; 32];
        out.copy_from_slice(&content);
        Ok(Self {
            inner: Bid::from_parts(topology.inner, out),
        })
    }

    fn to_wire(&self) -> PyResult<String> {
        self.inner.wire().map_err(py_err)
    }

    fn r2_key(&self) -> String {
        self.inner.r2_key()
    }
}

#[pymodule]
fn snid_native(_py: Python<'_>, module: &Bound<'_, PyModule>) -> PyResult<()> {
    module.add_class::<PySnid>()?;
    module.add_class::<PySgid>()?;
    module.add_class::<PyNid>()?;
    module.add_class::<PyWid>()?;
    module.add_class::<PyXid>()?;
    module.add_class::<PyKid>()?;
    module.add_class::<PyLid>()?;
    module.add_class::<PyEid>()?;
    module.add_class::<PyBid>()?;
    module.add_function(wrap_pyfunction!(encode_fixed64_pair, module)?)?;
    module.add_function(wrap_pyfunction!(decode_fixed64_pair, module)?)?;
    module.add_function(wrap_pyfunction!(generate_batch_bytes, module)?)?;
    module.add_function(wrap_pyfunction!(generate_batch_tensor_bytes, module)?)?;
    Ok(())
}

#[pyfunction]
fn encode_fixed64_pair<'py>(py: Python<'py>, hi: i64, lo: i64) -> Bound<'py, PyBytes> {
    PyBytes::new(py, &snid::encode_fixed64_pair(hi, lo))
}

#[pyfunction]
fn decode_fixed64_pair(raw: Vec<u8>) -> PyResult<(i64, i64)> {
    snid::decode_fixed64_pair(&raw).map_err(py_err)
}

#[pyfunction]
fn generate_batch_bytes<'py>(py: Python<'py>, count: usize) -> Bound<'py, PyBytes> {
    PyBytes::new(py, &Snid::generate_binary_batch(count))
}

#[pyfunction]
fn generate_batch_tensor_bytes<'py>(py: Python<'py>, count: usize) -> Bound<'py, PyBytes> {
    PyBytes::new(py, &Snid::generate_tensor_batch_be_bytes(count))
}
