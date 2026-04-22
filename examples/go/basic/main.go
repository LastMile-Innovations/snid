package main

import (
	"fmt"

	"github.com/LastMile-Innovations/snid"
)

func main() {
	// Generate a new SNID
	id := snid.NewFast()
	fmt.Printf("Generated ID: %s\n", id.UUID())

	// Format as wire string with atom
	wire := id.String(snid.Matter)
	fmt.Printf("Wire format: %s\n", wire)

	// Parse a wire string
	parsed, atom, err := snid.FromString(wire)
	if err != nil {
		fmt.Printf("Error parsing: %v\n", err)
		return
	}
	fmt.Printf("Parsed ID: %s (atom: %s)\n", parsed.UUID(), atom)

	// Generate batch of IDs
	batch := snid.NewBatch(snid.Matter, 5)
	fmt.Printf("Generated %d batch IDs\n", len(batch))
	for i, id := range batch {
		fmt.Printf("  [%d] %s\n", i, id.String(snid.Matter))
	}

	// Generate a spatial ID
	spatialID := snid.NewSpatial(37.7749, -122.4194) // San Francisco
	fmt.Printf("Spatial ID: %s\n", spatialID.String(snid.Location))
}
