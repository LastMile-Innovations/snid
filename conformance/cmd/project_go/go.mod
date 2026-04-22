module github.com/neighbor/snid/conformance/project-go

go 1.25.0

require github.com/neighbor/snid/go v0.0.0

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/mr-tron/base58 v1.3.0 // indirect
	github.com/uber/h3-go/v4 v4.4.1 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
)

replace github.com/neighbor/snid/go => ../../../go
