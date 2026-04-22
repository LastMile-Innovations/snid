package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var (
	benchmark = flag.String("benchmark", "all", "Benchmark to run: insert-throughput, index-size, query-performance, network-overhead, all")
	rows      = flag.Int("rows", 10000000, "Number of rows to insert")
	workers   = flag.Int("workers", 10, "Number of concurrent workers")
	runs      = flag.Int("runs", 10, "Number of benchmark runs")
	dbURL     = flag.String("db", "postgres://benchmark:benchmark@localhost:5432/snid_bench?sslmode=disable", "Database connection string")
)

func main() {
	flag.Parse()

	db, err := sql.Open("postgres", *dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Wait for database to be ready
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	switch *benchmark {
	case "insert-throughput":
		runInsertThroughputBenchmark(db, *rows, *workers, *runs)
	case "index-size":
		runIndexSizeBenchmark(db)
	case "query-performance":
		runQueryPerformanceBenchmark(db)
	case "network-overhead":
		runNetworkOverheadBenchmark()
	case "all":
		runInsertThroughputBenchmark(db, *rows, *workers, *runs)
		runIndexSizeBenchmark(db)
		runQueryPerformanceBenchmark(db)
		runNetworkOverheadBenchmark()
	default:
		log.Fatalf("Unknown benchmark: %s", *benchmark)
	}
}

func runInsertThroughputBenchmark(db *sql.DB, rows, workers, runs int) {
	fmt.Printf("Running insert throughput benchmark (%d rows, %d workers, %d runs)\n", rows, workers, runs)

	idTypes := []string{"snid_binary", "snid_uuid", "uuidv7", "uuidv4", "ulid", "sequential"}

	for _, idType := range idTypes {
		fmt.Printf("\nBenchmarking %s...\n", idType)

		totalTime := time.Duration(0)
		for run := 0; run < runs; run++ {
			// Clear table
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", idType))

			// Warmup
			insertRows(db, idType, rows/10, workers)

			// Actual benchmark
			start := time.Now()
			insertRows(db, idType, rows, workers)
			elapsed := time.Since(start)
			totalTime += elapsed

			fmt.Printf("  Run %d: %v (%.2f rows/sec)\n", run+1, elapsed, float64(rows)/elapsed.Seconds())
		}

		avgTime := totalTime / time.Duration(runs)
		avgThroughput := float64(rows) / avgTime.Seconds()

		fmt.Printf("  Average: %v (%.2f rows/sec)\n", avgTime, avgThroughput)
	}
}

func insertRows(db *sql.DB, table string, rows, workers int) {
	// TODO: Implement actual ID generation and insertion
	// This is a skeleton - needs:
	// - SNID generation (native and UUIDv7 mode)
	// - UUIDv7 generation
	// - UUIDv4 generation
	// - ULID generation
	// - Sequential BIGINT (auto-increment)
	// - Batched inserts (1000-10000 rows per transaction)
	// - Concurrent worker coordination

	stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s (id, data) VALUES ($1, $2)", table))
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	for i := 0; i < rows; i++ {
		// Generate ID based on table type
		var id interface{}
		switch table {
		case "snid_binary":
			// id = snid.NewFast().ToBytes()
			id = []byte("placeholder")
		case "snid_uuid", "uuidv7":
			// id = uuid.NewV7()
			id = "00000000-0000-0000-0000-000000000000"
		case "uuidv4":
			// id = uuid.NewV4()
			id = "00000000-0000-0000-0000-000000000000"
		case "ulid":
			// id = ulid.New()
			id = "placeholder"
		case "sequential":
			// Auto-increment, no ID generation needed
			id = nil
		}

		if id != nil {
			_, err = tx.Exec(fmt.Sprintf("INSERT INTO %s (id, data) VALUES ($1, $2)", table), id, "test data")
		} else {
			_, err = tx.Exec(fmt.Sprintf("INSERT INTO %s (data) VALUES ($1)", table), "test data")
		}
		if err != nil {
			log.Fatalf("Failed to insert row: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}
}

func runIndexSizeBenchmark(db *sql.DB) {
	fmt.Println("Running index size benchmark")

	idTypes := []string{"snid_binary", "snid_uuid", "uuidv7", "uuidv4", "ulid", "sequential"}

	for _, idType := range idTypes {
		fmt.Printf("\nAnalyzing %s...\n", idType)

		var tableSize, indexSize, totalSize int64
		var leafPageDensity float64

		// Get table size
		db.QueryRow(fmt.Sprintf("SELECT pg_relation_size('%s')", idType)).Scan(&tableSize)

		// Get index size
		db.QueryRow(fmt.Sprintf("SELECT pg_indexes_size('%s')", idType)).Scan(&indexSize)

		// Get total size
		db.QueryRow(fmt.Sprintf("SELECT pg_total_relation_size('%s')", idType)).Scan(&totalSize)

		// Get leaf page density (requires pgstattuple)
		db.QueryRow(fmt.Sprintf("SELECT leaf_fragmentation FROM pgstattuple('%s')", idType)).Scan(&leafPageDensity)

		fmt.Printf("  Table size: %d bytes (%.2f MB)\n", tableSize, float64(tableSize)/(1024*1024))
		fmt.Printf("  Index size: %d bytes (%.2f MB)\n", indexSize, float64(indexSize)/(1024*1024))
		fmt.Printf("  Total size: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/(1024*1024))
		fmt.Printf("  Leaf page density: %.2f%%\n", leafPageDensity)
	}
}

func runQueryPerformanceBenchmark(db *sql.DB) {
	fmt.Println("Running query performance benchmark")

	idTypes := []string{"snid_binary", "snid_uuid", "uuidv7", "uuidv4", "ulid", "sequential"}

	for _, idType := range idTypes {
		fmt.Printf("\nBenchmarking %s...\n", idType)

		// Point lookup benchmark
		var id string
		db.QueryRow(fmt.Sprintf("SELECT id FROM %s LIMIT 1", idType)).Scan(&id)

		iterations := 10000
		start := time.Now()
		for i := 0; i < iterations; i++ {
			var data string
			db.QueryRow(fmt.Sprintf("SELECT data FROM %s WHERE id = $1", idType), id).Scan(&data)
		}
		elapsed := time.Since(start)

		avgLatency := elapsed / time.Duration(iterations)
		fmt.Printf("  Point lookup: %v avg latency (%.2f ops/sec)\n", avgLatency, float64(iterations)/elapsed.Seconds())

		// Range scan benchmark
		start = time.Now()
		for i := 0; i < iterations; i++ {
			rows, _ := db.Query(fmt.Sprintf("SELECT * FROM %s WHERE id > $1 ORDER BY id LIMIT 1000", idType), id)
			rows.Close()
		}
		elapsed = time.Since(start)

		avgLatency = elapsed / time.Duration(iterations)
		fmt.Printf("  Range scan: %v avg latency (%.2f ops/sec)\n", avgLatency, float64(iterations)/elapsed.Seconds())
	}
}

func runNetworkOverheadBenchmark() {
	fmt.Println("Running network overhead benchmark")

	idTypes := map[string]string{
		"snid":   "22",
		"uuidv7": "36",
		"uuidv4": "36",
		"ulid":   "26",
		"nanoid": "21",
	}

	for idType, length := range idTypes {
		fmt.Printf("\n%s: %s chars\n", idType, length)

		// Calculate JSON payload size
		jsonSize := len(fmt.Sprintf(`{"id":"%s","data":"test"}`, idType))
		fmt.Printf("  JSON payload size: %d bytes\n", jsonSize)

		// Calculate bandwidth at 10k req/sec
		bandwidthPerSec := jsonSize * 10000
		bandwidthPerDay := bandwidthPerSec * 86400
		fmt.Printf("  Bandwidth at 10k req/sec: %.2f MB/sec (%.2f GB/day)\n", float64(bandwidthPerSec)/(1024*1024), float64(bandwidthPerDay)/(1024*1024*1024))
	}
}
