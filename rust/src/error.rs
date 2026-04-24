//! Error types for SNID operations.

use std::fmt;

#[derive(Debug)]
pub enum Error {
    InvalidLength,
    InvalidFormat,
    InvalidAtom,
    InvalidPayload,
    ChecksumMismatch,
    InvalidContentHash,
    InvalidKey,
    InvalidSignature,
    Random(getrandom::Error),
    #[cfg(feature = "data")]
    Json(serde_json::Error),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::InvalidLength => write!(f, "invalid length"),
            Error::InvalidFormat => write!(f, "invalid format"),
            Error::InvalidAtom => write!(f, "invalid atom"),
            Error::InvalidPayload => write!(f, "invalid payload"),
            Error::ChecksumMismatch => write!(f, "checksum mismatch"),
            Error::InvalidContentHash => write!(f, "invalid content hash"),
            Error::InvalidKey => write!(f, "invalid key"),
            Error::InvalidSignature => write!(f, "invalid signature"),
            Error::Random(e) => write!(f, "random source error: {}", e),
            #[cfg(feature = "data")]
            Error::Json(e) => write!(f, "json error: {}", e),
        }
    }
}

impl std::error::Error for Error {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Error::Random(e) => Some(e),
            #[cfg(feature = "data")]
            Error::Json(e) => Some(e),
            _ => None,
        }
    }
}

impl From<getrandom::Error> for Error {
    fn from(value: getrandom::Error) -> Self {
        Self::Random(value)
    }
}

#[cfg(feature = "data")]
impl From<serde_json::Error> for Error {
    fn from(value: serde_json::Error) -> Self {
        Self::Json(value)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_error_debug() {
        let error = Error::InvalidLength;
        let debug_str = format!("{:?}", error);
        assert!(debug_str.contains("InvalidLength"));
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_error_from_json() {
        let json_err = serde_json::from_str::<serde_json::Value>("").unwrap_err();
        let error = Error::from(json_err);
        assert!(matches!(error, Error::Json(_)));
    }
}
