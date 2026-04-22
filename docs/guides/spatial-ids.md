# Spatial IDs (SGID) Guide

SGID (Spatial ID) provides location-aware identifiers using H3 geospatial encoding.

## Overview

SGID is a 128-bit spatial identifier that preserves H3 locality for lexicographic scans. It's designed for:

- Building and location tracking
- Static sensor networks
- Geospatial indexing
- Location-based services

## H3 Integration

SGID uses Uber's H3 geospatial indexing system for spatial encoding.

**Key features:**
- Hierarchical hexagonal grid
- Multi-resolution indexing
- Spatial locality preservation
- Efficient range queries

## Generating SGIDs

### Go

```go
package main

import (
    "fmt"
    "github.com/neighbor/snid"
)

func main() {
    // Generate SGID for San Francisco (default resolution 12)
    sgid := snid.NewSpatial(37.7749, -122.4194)
    fmt.Printf("SGID: %s\n", sgid.String(snid.Location))

    // Generate with specific resolution
    sgid := snid.NewSpatialPrecise(37.7749, -122.4194, 10)
    fmt.Printf("SGID (res 10): %s\n", sgid.String(snid.Location))
}
```

### Rust

```rust
use snid::SGID;

fn main() {
    // Generate from H3 cell
    let h3_cell = 0x8a2a1072b59ffff; // San Francisco at resolution 12
    let entropy = 12345;
    let sgid = SGID::from_spatial_parts(h3_cell, entropy);
    println!("SGID: {}", sgid.to_wire("LOC"));
}
```

### Python

```python
import snid

# Generate from H3 cell
h3_cell = 0x8a2a1072b59ffff  # San Francisco at resolution 12
entropy = 12345
sgid = snid.SGID.from_spatial_parts(h3_cell, entropy)
print(f"SGID: {sgid.to_wire('LOC')}")
```

## H3 Resolution Guide

H3 supports resolutions from 0 (coarse) to 15 (fine):

| Resolution | Cell Area | Cell Edge | Use Case |
|------------|-----------|-----------|----------|
| 0 | 4,250,547 km² | 1,108 km | Countries |
| 1 | 607,221 km² | 418 km | Large regions |
| 2 | 86,745 km² | 158 km | States/provinces |
| 3 | 12,392 km² | 59 km | Metro areas |
| 4 | 1,770 km² | 22 km | Cities |
| 5 | 253 km² | 8 km | Districts |
| 6 | 36 km² | 3 km | Neighborhoods |
| 7 | 5.1 km² | 1 km | Blocks |
| 8 | 0.73 km² | 382 m | Streets |
| 9 | 0.10 km² | 136 m | Buildings |
| 10 | 0.015 km² | 48 m | Precise locations |
| 11 | 0.0021 km² | 17 m | Rooms |
| 12 | 0.00030 km² | 6 m | Furniture |
| 13 | 0.000043 km² | 2 m | Precise tracking |
| 14 | 0.0000061 km² | 0.7 m | High precision |
| 15 | 0.00000087 km² | 0.2 m | Ultra precision |

## Converting Lat/Lng to H3

### Go

```go
import "github.com/uber/h3-go/v4"

lat, lng := 37.7749, -122.4194
resolution := 12
cell := h3.FromGeo(h3.Geo{Latitude: lat, Longitude: lng}, resolution)
```

### Python

```python
import h3py

lat, lng = 37.7749, -122.4194
resolution = 12
cell = h3py.latlng_to_cell(lat, lng, resolution)
```

## Spatial Queries

### Nearby Cells

```python
import h3py
import snid

# Get nearby cells
cell = 0x8a2a1072b59ffff
k = 1  # Ring distance
nearby = h3py.grid_disk(cell, k)

# Generate SGIDs for nearby cells
for nearby_cell in nearby:
    sgid = snid.SGID.from_spatial_parts(nearby_cell, entropy)
    print(sgid)
```

### Cell to Boundary

```python
import h3py

cell = 0x8a2a1072b59ffff
boundary = h3py.cell_to_boundary(cell)
print(f"Boundary: {boundary}")
```

## Use Cases

### Building Tracking

```go
// Track buildings by location
buildingID := snid.NewSpatial(buildingLat, buildingLng)
```

### Sensor Networks

```python
# Static sensor locations
sensorID = snid.SGID.from_spatial_parts(sensorCell, sensorID)
```

### Geospatial Indexing

```sql
-- Store SGIDs in Postgres with spatial index
CREATE TABLE locations (
    id BYTEA PRIMARY KEY,
    lat FLOAT,
    lng FLOAT
);

CREATE INDEX idx_locations_spatial ON locations USING GIST (ST_Point(lng, lat));
```

### Location-Based Services

```python
# Find nearby locations
user_location = snid.SGID.from_spatial_parts(userCell, entropy)
nearby_cells = h3py.grid_disk(userCell, k=5)
```

## Performance Considerations

- SGID generation is slower than SNID (~20ns vs ~3.7ns in Go)
- H3 operations add overhead
- Use appropriate resolution for your use case
- Cache H3 cell calculations when possible

## Limitations

- SGID requires H3 library dependency
- Spatial queries require additional indexing
- Not suitable for high-frequency ID generation
- H3 resolution trade-offs between precision and performance

## Next Steps

- [Identifier Families](identifier-families.md) - All ID families
- [Neural IDs](neural-ids.md) - Semantic IDs for ML
- [Storage Contracts](storage-contracts.md) - Database integration
- [H3 Documentation](https://h3geo.org/docs/) - Full H3 reference
