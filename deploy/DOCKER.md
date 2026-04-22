# Docker Deployment Guide

This guide covers deploying the SNID Benchmarking Platform using Docker Compose or bare Docker.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+ (for compose deployment)
- 4GB+ RAM available
- 10GB+ disk space

## Quick Start with Docker Compose

### 1. Build and Start

```bash
cd /path/to/snid/benchmarks

# Create results directory
mkdir -p results

# Start in web dashboard mode
docker-compose up -d

# Or start in CLI mode for one-off run
docker-compose --profile cli up snid-benchmarks-cli
```

### 2. Access Dashboard

Open http://localhost:8080 in your browser.

### 3. View Logs

```bash
docker-compose logs -f
```

### 4. Stop Services

```bash
docker-compose down
```

## Configuration

### Environment Variables

Create a `.env` file in the `benchmarks/` directory:

```bash
# Mode: web (dashboard) or cli (one-off)
BENCH_MODE=web

# Pure mode: run benchmarks in isolated subprocess (zero harness overhead)
BENCH_PURE_MODE=true

# Results directory (volume mount)
RESULTS_DIR=/app/results

# Port for web dashboard
PORT=8080

# Regression threshold percentage
REGRESSION_THRESHOLD=10

# Suites to run in CLI mode (comma-separated)
BENCH_SUITES=go,rust,python
```

### Pure Mode (Zero Overhead)

The platform uses **pure mode** by default to ensure the benchmarking harness does not affect results:

- Benchmarks run in an isolated subprocess with no FastAPI or dashboard code loaded
- Result files are written only after benchmark completion
- No logging or metrics collection during measurement
- Set `BENCH_PURE_MODE=false` to disable (not recommended for accurate results)

### Custom Port

```bash
# Edit docker-compose.yml or use env var
PORT=9000 docker-compose up
```

## Bare Docker Deployment

### Build Image

```bash
cd /path/to/snid
docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .
```

### Run Web Dashboard

```bash
docker run -d \
  --name snid-benchmarks \
  -p 8080:8080 \
  -v $(pwd)/benchmarks/results:/app/results \
  -e BENCH_MODE=web \
  -e RESULTS_DIR=/app/results \
  snid-benchmarks:latest
```

### Run CLI Benchmark

```bash
docker run --rm \
  -v $(pwd)/benchmarks/results:/app/results \
  -e BENCH_MODE=cli \
  -e BENCH_SUITES=all \
  snid-benchmarks:latest
```

### Run Specific Suite

```bash
docker run --rm \
  -v $(pwd)/benchmarks/results:/app/results \
  -e BENCH_MODE=cli \
  snid-benchmarks:latest \
  python benchmarks/runner.py go
```

## Volume Persistence

Results are persisted via Docker volume mounts:

```bash
# Local directory mount (recommended)
-v $(pwd)/benchmarks/results:/app/results

# Named volume
-v snid-results:/app/results
```

## Health Checks

The container includes a built-in health check:

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' snid-benchmarks

# View health logs
docker inspect --format='{{json .State.Health}}' snid-benchmarks | jq
```

## Resource Limits

Set resource limits for heavy benchmark workloads:

```bash
docker run -d \
  --name snid-benchmarks \
  --cpus="2.0" \
  --memory="4g" \
  -p 8080:8080 \
  -v $(pwd)/benchmarks/results:/app/results \
  snid-benchmarks:latest
```

## Multi-Host Deployment

### Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c benchmarks/docker-compose.yml snid-benchmarks

# Scale services
docker service scale snid-benchmarks_snid-benchmarks=3
```

### Kubernetes

See `deploy/KUBERNETES.md` for K8s deployment instructions.

## Troubleshooting

### Build Failures

```bash
# Clear Docker cache
docker system prune -a

# Build with no cache
docker build --no-cache -f benchmarks/Dockerfile -t snid-benchmarks .
```

### Permission Issues

```bash
# Fix results directory permissions
sudo chown -R $USER:$USER benchmarks/results
chmod 755 benchmarks/results
```

### Container Won't Start

```bash
# Check logs
docker logs snid-benchmarks

# Run in foreground to see errors
docker run --rm -it \
  -v $(pwd)/benchmarks/results:/app/results \
  -e BENCH_MODE=cli \
  snid-benchmarks:latest \
  /bin/bash
```

## Performance Tuning

### Build Optimization

The Dockerfile uses multi-stage builds for faster rebuilds:

```bash
# Rebuild only changed layers
docker build --cache-from snid-benchmarks:latest -f benchmarks/Dockerfile -t snid-benchmarks .
```

### Runtime Performance

For maximum benchmark performance:

1. Use `--cpus` to allocate dedicated cores
2. Use `--memory` to prevent swapping
3. Mount results on fast storage (SSD/NVMe)
4. Disable Docker's logging driver for benchmarks:

```bash
docker run --log-driver=none ...
```

## Backup Results

```bash
# Backup results directory
tar -czf snid-benchmark-results-$(date +%Y%m%d).tar.gz benchmarks/results/

# Restore
tar -xzf snid-benchmark-results-YYYYMMDD.tar.gz
```

## Updating

```bash
# Pull latest code
git pull

# Rebuild image
docker build -f benchmarks/Dockerfile -t snid-benchmarks:latest .

# Restart service
docker-compose up -d --force-recreate
```
