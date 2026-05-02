module github.com/LastMile-Innovations/snhash/snidbridge

go 1.25.0

require (
	github.com/LastMile-Innovations/snhash v0.0.0
	github.com/LastMile-Innovations/snid v0.0.0
)

require (
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/uber/h3-go/v4 v4.4.1 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/LastMile-Innovations/snhash => ..

replace github.com/LastMile-Innovations/snid => ../../go
