package snid

import "time"

// =============================================================================
// UNIFIED DNA INTERFACE
// =============================================================================
// This provides a single, polymorphic entry point for all ID generation.
// It wraps the high-performance engines (Turbo, Spatial, Asset) in a
// user-friendly Options API.

type config struct {
	mode     IDType
	lat, lng float64
	cid      ID
	tenantID string
	serial   string
	category string
	brand    string
	specs    string
}

type Option func(*config)

// WithLocation requests a Spatial DNA anchored to coordinates.
func WithLocation(lat, lng float64) Option {
	return func(c *config) {
		c.mode = TypeSpatial
		c.lat = lat
		c.lng = lng
	}
}

// WithAsset requests an Asset DNA for a physical item.
func WithAsset(cid ID, tenantID, serial string) Option {
	return func(c *config) {
		c.mode = TypeAsset
		c.cid = cid
		c.tenantID = tenantID
		c.serial = serial
	}
}

// WithCatalog requests a Catalog DNA for a product definition.
func WithCatalog(category, brand, specs string) Option {
	return func(c *config) {
		c.mode = TypeCatalog
		c.category = category
		c.brand = brand
		c.specs = specs
	}
}

// WithTime explicitly requests a Time-Ordered ID (Default).
func WithTime() Option {
	return func(c *config) {
		c.mode = TypeTime
	}
}

// NewGenerator generates a polymorphic ID based on the provided options.
// Defaults to Time-Ordered (NewFast) if no options provided.
func NewGenerator(opts ...Option) ID {
	if len(opts) == 0 {
		return NewFast()
	}

	cfg := config{mode: TypeTime}
	for _, opt := range opts {
		opt(&cfg)
	}

	switch cfg.mode {
	case TypeSpatial:
		return NewSpatial(cfg.lat, cfg.lng)
	case TypeAsset:
		return NewAsset(cfg.cid, cfg.tenantID, cfg.serial)
	case TypeCatalog:
		return NewCatalog(cfg.category, cfg.brand, cfg.specs)
	case TypeTime:
		return NewFast()
	}

	return NewFast()
}

// CreatedAt extracts the creation time for Time-based IDs.
// Returns zero time for other types (Spatial, Asset, etc).
func (id ID) CreatedAt() time.Time {
	if id.Type() != TypeTime {
		return time.Time{}
	}
	return id.Time()
}
