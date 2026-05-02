package snhash

import (
	"encoding/binary"
	"io"
	"os"
	"sort"

	"github.com/zeebo/blake3"
	"github.com/zeebo/xxh3"
)

// HashBytes builds a manifest for an in-memory byte slice.
func HashBytes(data []byte, options ...Option) (Manifest, error) {
	opts := applyOptions(options)
	if len(opts.Algorithms) == 1 && opts.Algorithms[0] == AlgorithmBLAKE3_256 && opts.ChunkSize == 0 {
		sum := blake3.Sum256(data)
		digest, err := NewDigest(AlgorithmBLAKE3_256, sum[:])
		if err != nil {
			return Manifest{}, err
		}
		return Manifest{
			Version:          ManifestVersion,
			Profile:          opts.Profile,
			Size:             int64(len(data)),
			MediaType:        opts.MediaType,
			Canonicalization: opts.Canonicalization,
			SourceName:       opts.SourceName,
			Digests:          []Digest{digest},
			Custody:          opts.Custody,
			CreatedAt:        opts.CreatedAt.UTC(),
		}, nil
	}
	if len(opts.Algorithms) == 1 && opts.Algorithms[0] == AlgorithmXXH3_64 && opts.ChunkSize == 0 {
		digest, err := NewDigest(AlgorithmXXH3_64, xxh3Bytes(data))
		if err != nil {
			return Manifest{}, err
		}
		return Manifest{
			Version:          ManifestVersion,
			Profile:          opts.Profile,
			Size:             int64(len(data)),
			MediaType:        opts.MediaType,
			Canonicalization: opts.Canonicalization,
			SourceName:       opts.SourceName,
			Digests:          []Digest{digest},
			Custody:          opts.Custody,
			CreatedAt:        opts.CreatedAt.UTC(),
		}, nil
	}
	if isAlgorithmSet(opts.Algorithms, AlgorithmBLAKE3_256, AlgorithmXXH3_64) && opts.ChunkSize == 0 {
		blakeSum := blake3.Sum256(data)
		blakeDigest, err := NewDigest(AlgorithmBLAKE3_256, blakeSum[:])
		if err != nil {
			return Manifest{}, err
		}
		xxhDigest, err := NewDigest(AlgorithmXXH3_64, xxh3Bytes(data))
		if err != nil {
			return Manifest{}, err
		}
		return Manifest{
			Version:          ManifestVersion,
			Profile:          opts.Profile,
			Size:             int64(len(data)),
			MediaType:        opts.MediaType,
			Canonicalization: opts.Canonicalization,
			SourceName:       opts.SourceName,
			Digests:          []Digest{blakeDigest, xxhDigest},
			Custody:          opts.Custody,
			CreatedAt:        opts.CreatedAt.UTC(),
		}, nil
	}

	global, err := newStates(opts.Algorithms)
	if err != nil {
		return Manifest{}, err
	}
	if err := writeAll(global, data); err != nil {
		return Manifest{}, err
	}
	digests, err := digestsFromStates(global)
	if err != nil {
		return Manifest{}, err
	}

	var chunks []Chunk
	if opts.ChunkSize > 0 {
		for offset := int64(0); offset < int64(len(data)); offset += opts.ChunkSize {
			end := offset + opts.ChunkSize
			if end > int64(len(data)) {
				end = int64(len(data))
			}
			chunk, err := hashChunk(len(chunks), offset, data[offset:end], opts.Algorithms)
			if err != nil {
				return Manifest{}, err
			}
			chunks = append(chunks, chunk)
		}
	}
	treeRoot, err := treeRootFromChunks(chunks)
	if err != nil {
		return Manifest{}, err
	}

	return Manifest{
		Version:          ManifestVersion,
		Profile:          opts.Profile,
		Size:             int64(len(data)),
		MediaType:        opts.MediaType,
		Canonicalization: opts.Canonicalization,
		SourceName:       opts.SourceName,
		Digests:          digests,
		TreeRoot:         treeRoot,
		Chunks:           chunks,
		Custody:          opts.Custody,
		CreatedAt:        opts.CreatedAt.UTC(),
	}, nil
}

// HashReader builds a manifest from a stream. Size is recorded as bytes read.
func HashReader(r io.Reader, options ...Option) (Manifest, error) {
	opts := applyOptions(options)
	return hashReader(r, -1, opts)
}

// HashFile builds a manifest from a file path.
func HashFile(path string, options ...Option) (Manifest, error) {
	opts := applyOptions(options)
	if opts.SourceName == "" {
		opts.SourceName = path
	}
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Manifest{}, err
	}
	return hashReader(file, info.Size(), opts)
}

func hashReader(r io.Reader, expectedSize int64, opts Options) (Manifest, error) {
	global, err := newStates(opts.Algorithms)
	if err != nil {
		return Manifest{}, err
	}
	var chunks []Chunk
	var total int64

	if opts.ChunkSize > 0 {
		buf := make([]byte, opts.ChunkSize)
		for {
			n, readErr := io.ReadFull(r, buf)
			if n > 0 {
				part := buf[:n]
				if err := writeAll(global, part); err != nil {
					return Manifest{}, err
				}
				chunk, err := hashChunk(len(chunks), total, part, opts.Algorithms)
				if err != nil {
					return Manifest{}, err
				}
				chunks = append(chunks, chunk)
				total += int64(n)
			}
			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				break
			}
			if readErr != nil {
				return Manifest{}, readErr
			}
		}
	} else {
		buf := make([]byte, 128*1024)
		for {
			n, readErr := r.Read(buf)
			if n > 0 {
				part := buf[:n]
				if err := writeAll(global, part); err != nil {
					return Manifest{}, err
				}
				total += int64(n)
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return Manifest{}, readErr
			}
		}
	}

	if expectedSize >= 0 && total != expectedSize {
		return Manifest{}, io.ErrUnexpectedEOF
	}
	digests, err := digestsFromStates(global)
	if err != nil {
		return Manifest{}, err
	}
	treeRoot, err := treeRootFromChunks(chunks)
	if err != nil {
		return Manifest{}, err
	}
	return Manifest{
		Version:          ManifestVersion,
		Profile:          opts.Profile,
		Size:             total,
		MediaType:        opts.MediaType,
		Canonicalization: opts.Canonicalization,
		SourceName:       opts.SourceName,
		Digests:          digests,
		TreeRoot:         treeRoot,
		Chunks:           chunks,
		Custody:          opts.Custody,
		CreatedAt:        opts.CreatedAt.UTC(),
	}, nil
}

func hashChunk(index int, offset int64, data []byte, algorithms []Algorithm) (Chunk, error) {
	states, err := newStates(algorithms)
	if err != nil {
		return Chunk{}, err
	}
	if err := writeAll(states, data); err != nil {
		return Chunk{}, err
	}
	digests, err := digestsFromStates(states)
	if err != nil {
		return Chunk{}, err
	}
	return Chunk{
		Index:   index,
		Offset:  offset,
		Size:    int64(len(data)),
		Digests: digests,
	}, nil
}

type namedState struct {
	algorithm Algorithm
	state     digestState
}

func newStates(algorithms []Algorithm) ([]namedState, error) {
	out := make([]namedState, 0, len(algorithms))
	for _, algorithm := range normalizeAlgorithms(algorithms) {
		state, err := newDigestState(algorithm)
		if err != nil {
			return nil, err
		}
		out = append(out, namedState{algorithm: algorithm, state: state})
	}
	return out, nil
}

func writeAll(states []namedState, data []byte) error {
	for i := range states {
		if _, err := states[i].state.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func digestsFromStates(states []namedState) ([]Digest, error) {
	out := make([]Digest, 0, len(states))
	for _, state := range states {
		digest, err := digestFromState(state.algorithm, state.state)
		if err != nil {
			return nil, err
		}
		out = append(out, digest)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Algorithm < out[j].Algorithm })
	return out, nil
}

func normalizeAlgorithms(algorithms []Algorithm) []Algorithm {
	seen := make(map[Algorithm]struct{}, len(algorithms))
	out := make([]Algorithm, 0, len(algorithms))
	for _, algorithm := range algorithms {
		if algorithm == "" {
			continue
		}
		if _, ok := seen[algorithm]; ok {
			continue
		}
		seen[algorithm] = struct{}{}
		out = append(out, algorithm)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func isAlgorithmSet(actual []Algorithm, expected ...Algorithm) bool {
	actual = normalizeAlgorithms(actual)
	expected = normalizeAlgorithms(expected)
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}

func xxh3Bytes(data []byte) []byte {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], xxh3.Hash(data))
	return out[:]
}

func treeRootFromChunks(chunks []Chunk) (*Digest, error) {
	if len(chunks) == 0 {
		return nil, nil
	}
	hasher := blake3.New()
	_, _ = hasher.WriteString("snhash-tree-v1")
	_, _ = hasher.Write([]byte{0})
	var buf [24]byte
	for _, chunk := range chunks {
		binary.BigEndian.PutUint64(buf[0:8], uint64(chunk.Index))
		binary.BigEndian.PutUint64(buf[8:16], uint64(chunk.Offset))
		binary.BigEndian.PutUint64(buf[16:24], uint64(chunk.Size))
		if _, err := hasher.Write(buf[:]); err != nil {
			return nil, err
		}
		for _, digest := range normalizeDigests(chunk.Digests) {
			if _, err := hasher.Write([]byte(digest.Wire())); err != nil {
				return nil, err
			}
			if _, err := hasher.Write([]byte{0}); err != nil {
				return nil, err
			}
		}
	}
	sum := hasher.Sum(nil)
	digest, err := NewDigest(AlgorithmBLAKE3_256, sum)
	if err != nil {
		return nil, err
	}
	return &digest, nil
}
