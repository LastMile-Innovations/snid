package snhash

// Profile selects a deterministic set of digest algorithms and manifest defaults.
type Profile string

const (
	ProfileFastCAS  Profile = "FastCAS"
	ProfileEvidence Profile = "Evidence"
	ProfileAPI      Profile = "API"
	ProfileFIPS     Profile = "FIPS"
	ProfileCache    Profile = "Cache"
)

// Algorithms returns the canonical digest set for a profile.
func (p Profile) Algorithms() []Algorithm {
	switch p {
	case ProfileEvidence:
		return []Algorithm{AlgorithmBLAKE3_256, AlgorithmSHA256}
	case ProfileAPI:
		return []Algorithm{AlgorithmSHA256, AlgorithmBLAKE3_256}
	case ProfileFIPS:
		return []Algorithm{AlgorithmSHA256, AlgorithmSHA3_256, AlgorithmSHAKE256_256}
	case ProfileCache:
		return []Algorithm{AlgorithmXXH3_64, AlgorithmBLAKE3_256}
	default:
		return []Algorithm{AlgorithmBLAKE3_256}
	}
}

func (p Profile) normalized() Profile {
	switch p {
	case ProfileFastCAS, ProfileEvidence, ProfileAPI, ProfileFIPS, ProfileCache:
		return p
	default:
		return ProfileFastCAS
	}
}
