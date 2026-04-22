# Railway Deployment Guide

This guide covers deploying the SNID Benchmarking Platform to Railway with persistent volume storage.

## Prerequisites

- Railway account (https://railway.app)
- Railway CLI installed: `npm install -g @railway/cli`
- Git repository with SNID code

## Quick Start (5 minutes)

### 1. Initialize Railway Project

```bash
# Login to Railway
railway login

# Initialize project in SNID repository
cd /path/to/snid
railway init
```

### 2. Create Volume for Results Persistence

```bash
# Create a volume named "benchmark-results"
railway volume add benchmark-results
```

### 3. Deploy Service

```bash
# Build and deploy from Dockerfile
railway up
```

### 4. Configure Environment Variables

In Railway dashboard, set these variables:

```bash
BENCH_MODE=web
RESULTS_DIR=/app/results
RAILWAY_VOLUME=true
PORT=8080
REGRESSION_THRESHOLD=10
```

### 5. Mount Volume

In Railway service settings:
- Go to "Volumes" tab
- Mount `benchmark-results` volume to `/app/results`

### 6. Access Dashboard

Railway will provide a public URL. Access it to view the benchmark dashboard.

## Railway-Specific Configuration

### Environment Variables

| Variable | Value | Purpose |
|----------|-------|---------|
| `BENCH_MODE` | `web` | Run web dashboard mode |
| `BENCH_PURE_MODE` | `true` | Run benchmarks in isolated mode (zero harness overhead) |
| `RESULTS_DIR` | `/app/results` | Volume mount path |
| `RAILWAY_VOLUME` | `true` | Enable Railway volume features |
| `PORT` | `8080` | Railway-exposed port |
| `REGRESSION_THRESHOLD` | `10` | Alert threshold (%) |

### Pure Mode (Zero Overhead)

The platform uses **pure mode** by default to ensure the benchmarking harness does not affect results:

- Benchmarks run in an isolated subprocess with no FastAPI or dashboard code loaded
- Result files are written only after benchmark completion
- No logging or metrics collection during measurement
- Set `BENCH_PURE_MODE=false` to disable (not recommended for accurate results)

### Volume Mounting

Railway volumes provide persistent storage across deployments:

```yaml
# In Railway service settings (via UI or CLI)
Volume: benchmark-results
Mount Path: /app/results
```

### Health Checks

Railway automatically uses the health check endpoint defined in the Dockerfile:

```
GET /health
```

## One-Off Benchmark Runs

For running benchmarks without persistent dashboard:

```bash
# Run Go benchmarks only
railway run --env BENCH_MODE=cli --env BENCH_SUITES=go python benchmarks/runner.py go

# Run all suites
railway run --env BENCH_MODE=cli --env BENCH_SUITES=all python benchmarks/runner.py all
```

## Viewing Logs

```bash
# View real-time logs
railway logs

# View logs for specific deployment
railway logs --deployment <deployment-id>
```

## Scaling

Railway automatically scales based on traffic. For heavy benchmark workloads:

1. Go to service settings
2. Adjust "CPU" and "RAM" allocations
3. Consider using "Background Worker" for long-running benchmarks

## Cost Optimization

- Use `railway run` for one-off benchmarks (billed per execution)
- Keep dashboard service in standby when not actively benchmarking
- Set up scheduled triggers instead of always-on service

## Troubleshooting

### Volume Not Mounting

Ensure:
1. Volume is created: `railway volume`
2. Mount path matches `RESULTS_DIR` env var
3. Service is restarted after volume configuration

### Build Failures

Check:
1. Go version compatibility (requires 1.24)
2. Rust toolchain availability
3. Python version (3.12)

### Out of Memory

Increase RAM allocation in service settings or run suites individually.

## Integration with CI

See `deploy/CI.md` for GitHub Actions integration with Railway.

## Monitoring

Railway provides built-in metrics:
- CPU usage
- Memory usage
- Network traffic
- Request latency

Access via Railway dashboard "Metrics" tab.
