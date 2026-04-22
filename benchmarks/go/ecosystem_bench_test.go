package main

import (
	"testing"

	"github.com/bwmarrin/snowflake"
	gofrsuuid "github.com/gofrs/uuid"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/rs/xid"
	"github.com/sony/sonyflake"
)

// Industry Standard Baseline: UUIDv7
func BenchmarkGoogleUUIDv7_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id, _ := uuid.NewV7()
		_ = id
	}
}

func BenchmarkGoogleUUIDv7_NewString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id, _ := uuid.NewV7()
		_ = id.String()
	}
}

// External ID packages (not SNID)
func BenchmarkGoogleUUID_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uuid.New()
	}
}

func BenchmarkGoogleUUID_NewString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = uuid.New().String()
	}
}

func BenchmarkOklogULID_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ulid.Make()
	}
}

func BenchmarkRsXID_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = xid.New()
	}
}

func BenchmarkSonySonyflake_New(b *testing.B) {
	sf := sonyflake.NewSonyflake(sonyflake.Settings{})
	for i := 0; i < b.N; i++ {
		_, _ = sf.NextID()
	}
}

func BenchmarkBwmarrinSnowflake_New(b *testing.B) {
	node, _ := snowflake.NewNode(1)
	for i := 0; i < b.N; i++ {
		_ = node.Generate()
	}
}

func BenchmarkGofrsUUID_New(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gofrsuuid.Must(gofrsuuid.NewV4())
	}
}

func BenchmarkGofrsUUID_NewString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gofrsuuid.Must(gofrsuuid.NewV4()).String()
	}
}
