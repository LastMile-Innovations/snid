//! Error types for SNID operations.

use hex::FromHexError;

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
    Hex(FromHexError),
    #[cfg(feature = "data")]
    Json(serde_json::Error),
}

impl From<FromHexError> for Error {
    fn from(value: FromHexError) -> Self {
        Self::Hex(value)
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
    fn test_error_from_hex() {
        let hex_err = hex::FromHexError::InvalidStringLength;
        let error = Error::from(hex_err);
        assert!(matches!(error, Error::Hex(_)));
    }

    #[test]
    fn test_error_debug() {
        let error = Error::InvalidLength;
        let debug_str = format!("{:?}", error);
        assert!(debug_str.contains("InvalidLength"));
    }

    #[cfg(feature = "data")]
    #[test]
    fn test_error_from_json() {
        let json_err = serde_json::Error::eof();
        let error = Error::from(json_err);
        assert!(matches!(error, Error::Json(_)));
    }
}
