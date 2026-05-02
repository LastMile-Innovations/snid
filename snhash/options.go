package snhash

import "time"

const (
	defaultChunkSize        = 4 * 1024 * 1024
	defaultCanonicalization = "raw-bytes"
	defaultMediaType        = "application/octet-stream"
)

// Options configures hashing and manifest construction.
type Options struct {
	Profile          Profile
	Algorithms       []Algorithm
	MediaType        string
	Canonicalization string
	SourceName       string
	Custody          []CustodyEvent
	ChunkSize        int64
	CreatedAt        time.Time
}

// Option mutates Options.
type Option func(*Options)

// WithProfile selects a profile. Explicit algorithms still win.
func WithProfile(profile Profile) Option {
	return func(o *Options) { o.Profile = profile }
}

// WithAlgorithms selects exact digest algorithms.
func WithAlgorithms(algorithms ...Algorithm) Option {
	return func(o *Options) {
		o.Algorithms = append(o.Algorithms[:0], algorithms...)
	}
}

// WithMediaType sets the manifest media type.
func WithMediaType(mediaType string) Option {
	return func(o *Options) { o.MediaType = mediaType }
}

// WithCanonicalization records the byte canonicalization contract.
func WithCanonicalization(canonicalization string) Option {
	return func(o *Options) { o.Canonicalization = canonicalization }
}

// WithSourceName records a display/source name without affecting content identity.
func WithSourceName(name string) Option {
	return func(o *Options) { o.SourceName = name }
}

// WithCustody records chain-of-custody events in the proof envelope.
func WithCustody(events ...CustodyEvent) Option {
	return func(o *Options) {
		o.Custody = append(o.Custody[:0], events...)
	}
}

// WithChunkSize enables chunk manifests when size is greater than zero.
func WithChunkSize(size int64) Option {
	return func(o *Options) { o.ChunkSize = size }
}

// WithCreatedAt fixes manifest creation time for deterministic outputs.
func WithCreatedAt(t time.Time) Option {
	return func(o *Options) { o.CreatedAt = t }
}

func applyOptions(options []Option) Options {
	out := Options{
		Profile:          ProfileFastCAS,
		MediaType:        defaultMediaType,
		Canonicalization: defaultCanonicalization,
		ChunkSize:        0,
		CreatedAt:        time.Now().UTC(),
	}
	for _, option := range options {
		if option != nil {
			option(&out)
		}
	}
	out.Profile = out.Profile.normalized()
	if out.MediaType == "" {
		out.MediaType = defaultMediaType
	}
	if out.Canonicalization == "" {
		out.Canonicalization = defaultCanonicalization
	}
	if out.ChunkSize < 0 {
		out.ChunkSize = 0
	}
	if out.CreatedAt.IsZero() {
		out.CreatedAt = time.Now().UTC()
	}
	if len(out.Algorithms) == 0 {
		out.Algorithms = out.Profile.Algorithms()
	}
	out.Algorithms = normalizeAlgorithms(out.Algorithms)
	return out
}
