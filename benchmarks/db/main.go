package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LastMile-Innovations/snid"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	_ "github.com/lib/pq"
)

var (
	benchmark  = flag.String("benchmark", "all", "Benchmark to run: insert-throughput, index-size, query-performance, network-overhead, all")
	rows       = flag.Int("rows", 10000000, "Number of rows to insert")
	workers    = flag.Int("workers", 10, "Number of concurrent workers")
	runs       = flag.Int("runs", 10, "Number of benchmark runs")
	batchSize  = flag.Int("batch", 1000, "Batch size for inserts")
	dbURL      = flag.String("db", "postgres://benchmark:benchmark@localhost:5432/snid_bench?sslmode=disable", "Database connection string")
	resultsDir = flag.String("results", "../results", "Directory to save results")
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
		runInsertThroughputBenchmark(db, *rows, *workers, *batchSize, *runs)
	case "index-size":
		runIndexSizeBenchmark(db)
	case "query-performance":
		runQueryPerformanceBenchmark(db)
	case "network-overhead":
		runNetworkOverheadBenchmark()
	case "all":
		runInsertThroughputBenchmark(db, *rows, *workers, *batchSize, *runs)
		runIndexSizeBenchmark(db)
		runQueryPerformanceBenchmark(db)
		runNetworkOverheadBenchmark()
	default:
		log.Fatalf("Unknown benchmark: %s", *benchmark)
	}
}

func runInsertThroughputBenchmark(db *sql.DB, rows, workers, batchSize, runs int) {
	fmt.Printf("Running insert throughput benchmark (%d rows, %d workers, batch %d, %d runs)\n", rows, workers, batchSize, runs)

	idTypes := []string{"snid_binary", "snid_uuid", "uuidv7", "uuidv4", "ulid", "sequential"}
	results := make(map[string]InsertBenchmarkResult)

	for _, idType := range idTypes {
		fmt.Printf("\nBenchmarking %s...\n", idType)

		var totalElapsed time.Duration
		var throughputValues []float64

		for run := 0; run < runs; run++ {
			// Clear table
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", idType))

			// Warmup
			insertRows(db, idType, rows/10, workers, batchSize)
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", idType))

			// Actual benchmark
			elapsed := insertRows(db, idType, rows, workers, batchSize)
			totalElapsed += elapsed
			throughput := float64(rows) / elapsed.Seconds()
			throughputValues = append(throughputValues, throughput)

			fmt.Printf("  Run %d: %v (%.2f rows/sec)\n", run+1, elapsed, throughput)
		}

		avgTime := totalElapsed / time.Duration(runs)
		avgThroughput := float64(rows) / avgTime.Seconds()

		// Calculate stddev
		var sum, sumSq float64
		for _, v := range throughputValues {
			sum += v
			sumSq += v * v
		}
		mean := sum / float64(len(throughputValues))
		variance := (sumSq / float64(len(throughputValues))) - (mean * mean)
		stddev := 0.0
		if variance > 0 {
			stddev = variance
		}

		fmt.Printf("  Average: %v (%.2f rows/sec ± %.2f)\n", avgTime, avgThroughput, stddev)

		results[idType] = InsertBenchmarkResult{
			IDType:       idType,
			Rows:         rows,
			Workers:      workers,
			BatchSize:    batchSize,
			Runs:         runs,
			AvgTime:      avgTime.String(),
			AvgThroughput: avgThroughput,
			Stddev:       stddev,
			Throughputs:  throughputValues,
		}
	}

	// Save results
	saveInsertResults(results)
}

func insertRows(db *sql.DB, table string, totalRows, workers, batchSize int) time.Duration {
	start := time.Now()

	// Pre-generate IDs to isolate insert performance
	ids := preGenerateIDs(table, totalRows)

	// Worker pool
	var wg sync.WaitGroup
	rowsPerWorker := totalRows / workers
	var completed atomic.Int64

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerIndex int) {
			defer wg.Done()

			startIdx := workerIndex * rowsPerWorker
			endIdx := startIdx + rowsPerWorker
			if workerIndex == workers-1 {
				endIdx = totalRows
			}

			// Process in batches
			for batchStart := startIdx; batchStart < endIdx; batchStart += batchSize {
				batchEnd := batchStart + batchSize
				if batchEnd > endIdx {
					batchEnd = endIdx
				}

				tx, err := db.Begin()
				if err != nil {
					log.Printf("Worker %d: Failed to begin transaction: %v", workerIndex, err)
					return
				}

				stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (id, data) VALUES ($1, $2)", table))
				if err != nil {
					log.Printf("Worker %d: Failed to prepare statement: %v", workerIndex, err)
					tx.Rollback()
					return
				}

				for i := batchStart; i < batchEnd; i++ {
					_, err = stmt.Exec(ids[i], "test data")
					if err != nil {
						log.Printf("Worker %d: Failed to insert row %d: %v", workerIndex, i, err)
						tx.Rollback()
						stmt.Close()
						return
					}
				}

				stmt.Close()
				if err := tx.Commit(); err != nil {
					log.Printf("Worker %d: Failed to commit transaction: %v", workerIndex, err)
					return
				}

				completed.Add(int64(batchEnd - batchStart))
			}
		}
	}(w)
	}

	wg.Wait()
	return time.Since(start)
}

func preGenerateIDs(table string, count int) []interface{} {
	ids := make([]interface{}, count)
	entropy := ulid.Monotonic(time.Now(), 0)

	for i := 0; i < count; i++ {
		switch table {
		case "snid_binary":
			id := snid.NewFast()
			ids[i] = id[:]
		case "snid_uuid":
			id := snid.NewUUIDv7()
			ids[i] = id.UUIDString()
		case "uuidv7":
			id := uuid.Must(uuid.NewV7())
			ids[i] = id.String()
		case "uuidv4":
			id := uuid.Must(uuid.NewRandom())
			ids[i] = id.String()
		case "ulid":
			id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
			ids[i] = id.String()
		case "sequential":
			ids[i] = nil // Auto-increment
		}
	}

	return ids
}

func runIndexSizeBenchmark(db *sql.DB) {
	fmt.Println("Running index size benchmark")

	idTypes := []string{"snid_binary", "snid_uuid", "uuidv7", "uuidv4", "ulid", "sequential"}
	results := make(map[string]IndexSizeResult)

	for _, idType := range idTypes {
		fmt.Printf("\nAnalyzing %s...\n", idType)

		var tableSize, indexSize, totalSize int64
		var leafFragmentation float64

		// Get table size
		db.QueryRow(fmt.Sprintf("SELECT pg_relation_size('%s')", idType)).Scan(&tableSize)

		// Get index size
		db.QueryRow(fmt.Sprintf("SELECT pg_indexes_size('%s')", idType)).Scan(&indexSize)

		// Get total size
		db.QueryRow(fmt.Sprintf("SELECT pg_total_relation_size('%s')", idType)).Scan(&totalSize)

		// Get leaf page fragmentation (requires pgstattuple)
		err := db.QueryRow(fmt.Sprintf("SELECT leaf_fragmentation FROM pgstattuple('%s')", idType)).Scan(&leafFragmentation)
		if err != nil {
			leafFragmentation = 0
		}

		leafDensity := 100.0 - leafFragmentation

		fmt.Printf("  Table size: %d bytes (%.2f MB)\n", tableSize, float64(tableSize)/(1024*1024))
		fmt.Printf("  Index size: %d bytes (%.2f MB)\n", indexSize, float64(indexSize)/(1024*1024))
		fmt.Printf("  Total size: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/(1024*1024))
		fmt.Printf("  Leaf page density: %.2f%%\n", leafDensity)

		results[idType] = IndexSizeResult{
			IDType:          idType,
			TableSizeBytes:  tableSize,
			IndexSizeBytes:  indexSize,
			TotalSizeBytes:  totalSize,
			LeafDensity:     leafDensity,
		}
	}

	// Save results
	saveIndexSizeResults(results)
}

func runQueryPerformanceBenchmark(db *sql.DB) {
	fmt.Println("Running query performance benchmark")

	idTypes := []string{"snid_binary", "snid_uuid", "uuidv7", "uuidv4", "ulid", "sequential"}
	results := make(map[string]QueryPerformanceResult)

	for _, idType := range idTypes {
		fmt.Printf("\nBenchmarking %s...\n", idType)

		// Point lookup benchmark
		var id interface{}
		db.QueryRow(fmt.Sprintf("SELECT id FROM %s LIMIT 1", idType)).Scan(&id)

		iterations := 10000
		start := time.Now()
		for i := 0; i < iterations; i++ {
			var data string
			db.QueryRow(fmt.Sprintf("SELECT data FROM %s WHERE id = $1", idType), id).Scan(&data)
		}
		elapsed := time.Since(start)

		avgLatency := elapsed / time.Duration(iterations)
		pointOpsPerSec := float64(iterations) / elapsed.Seconds()
		fmt.Printf("  Point lookup: %v avg latency (%.2f ops/sec)\n", avgLatency, pointOpsPerSec)

		// Range scan benchmark
		start = time.Now()
		for i := 0; i < iterations; i++ {
			rows, _ := db.Query(fmt.Sprintf("SELECT * FROM %s WHERE id > $1 ORDER BY id LIMIT 1000", idType), id)
			rows.Close()
		}
		elapsed = time.Since(start)

		avgLatency = elapsed / time.Duration(iterations)
		rangeOpsPerSec := float64(iterations) / elapsed.Seconds()
		fmt.Printf("  Range scan: %v avg latency (%.2f ops/sec)\n", avgLatency, rangeOpsPerSec)

		results[idType] = QueryPerformanceResult{
			IDType:              idType,
			PointAvgLatency:     avgLatency.String(),
			PointOpsPerSec:      pointOpsPerSec,
			RangeAvgLatency:     avgLatency.String(),
			RangeOpsPerSec:      rangeOpsPerSec,
			Iterations:          iterations,
		}
	}

	// Save results
	saveQueryPerformanceResults(results)
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
	results := make(map[string]NetworkOverheadResult)

	for idType, length := range idTypes {
		fmt.Printf("\n%s: %s chars\n", idType, length)

		// Calculate JSON payload size
		jsonSize := len(fmt.Sprintf(`{"id":"%s","data":"test"}`, idType))
		fmt.Printf("  JSON payload size: %d bytes\n", jsonSize)

		// Calculate bandwidth at 10k req/sec
		bandwidthPerSec := jsonSize * 10000
		bandwidthPerDay := bandwidthPerSec * 86400
		fmt.Printf("  Bandwidth at 10k req/sec: %.2f MB/sec (%.2f GB/day)\n", float64(bandwidthPerSec)/(1024*1024), float64(bandwidthPerDay)/(1024*1024*1024))

		results[idType] = NetworkOverheadResult{
			IDType:           idType,
			StringLength:     length,
			JSONPayloadBytes: jsonSize,
			BandwidthPerSec:  bandwidthPerSec,
			BandwidthPerDay:  bandwidthPerDay,
		}
	}

	// Save results
	saveNetworkOverheadResults(results)
}

// Result types for JSON serialization
type InsertBenchmarkResult struct {
	IDType        string    `json:"id_type"`
	Rows          int       `json:"rows"`
	Workers       int       `json:"workers"`
	BatchSize     int       `json:"batch_size"`
	Runs          int       `json:"runs"`
	AvgTime       string    `json:"avg_time"`
	AvgThroughput float64   `json:"avg_throughput"`
	Stddev        float64   `json:"stddev"`
	Throughputs   []float64 `json:"throughputs"`
}

type IndexSizeResult struct {
	IDType          string  `json:"id_type"`
	TableSizeBytes  int64   `json:"table_size_bytes"`
	IndexSizeBytes  int64   `json:"index_size_bytes"`
	TotalSizeBytes  int64   `json:"total_size_bytes"`
	LeafDensity     float64 `json:"leaf_density"`
}

type QueryPerformanceResult struct {
	IDType          string  `json:"id_type"`
	PointAvgLatency string  `json:"point_avg_latency"`
	PointOpsPerSec  float64 `json:"point_ops_per_sec"`
	RangeAvgLatency string  `json:"range_avg_latency"`
	RangeOpsPerSec  float64 `json:"range_ops_per_sec"`
	Iterations      int     `json:"iterations"`
}

type NetworkOverheadResult struct {
	IDType           string `json:"id_type"`
	StringLength     string `json:"string_length"`
	JSONPayloadBytes int    `json:"json_payload_bytes"`
	BandwidthPerSec  int    `json:"bandwidth_per_sec"`
	BandwidthPerDay  int    `json:"bandwidth_per_day"`
}

func saveInsertResults(results map[string]InsertBenchmarkResult) {
	os.MkdirAll(*resultsDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s/insert-throughput-%s.json", *resultsDir, timestamp)

	data, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(filename, data, 0644)
	fmt.Printf("  Results saved to: %s\n", filename)
}

func saveIndexSizeResults(results map[string]IndexSizeResult) {
	os.MkdirAll(*resultsDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s/index-size-%s.json", *resultsDir, timestamp)

	data, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(filename, data, 0644)
	fmt.Printf("  Results saved to: %s\n", filename)
}

func saveQueryPerformanceResults(results map[string]QueryPerformanceResult) {
	os.MkdirAll(*resultsDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s/query-performance-%s.json", *resultsDir, timestamp)

	data, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(filename, data, 0644)
	fmt.Printf("  Results saved to: %s\n", filename)
}

func saveNetworkOverheadResults(results map[string]NetworkOverheadResult) {
	os.MkdirAll(*resultsDir, 0755)
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s/network-overhead-%s.json", *resultsDir, timestamp)

	data, _ := json.MarshalIndent(results, "", "  ")
	os.WriteFile(filename, data, 0644)
	fmt.Printf("  Results saved to: %s\n", filename)
}
