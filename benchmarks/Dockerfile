# Multi-stage Dockerfile for SNID Benchmarking Platform
# Supports Go 1.24, Rust 2021 edition, Python 3.12
# Optimized for Railway deployment with volume persistence

# =============================================================================
# Stage 1: Base with all build tools
# =============================================================================
FROM golang:1.24 AS base

# Install Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup default stable

# Install Python 3.12
RUN apt-get update && apt-get install -y \
    python3.12 \
    python3.12-venv \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Set Python 3.12 as default
RUN update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1
RUN update-alternatives --install /usr/bin/pip3 pip3 /usr/bin/pip3.12 1

# =============================================================================
# Stage 2: Build dependencies
# =============================================================================
FROM base AS builder

WORKDIR /app

# Copy source code
COPY go /app/go
COPY rust /app/rust
COPY python /app/python
COPY benchmarks /app/benchmarks

# Build Go dependencies
WORKDIR /app/go
RUN go mod download

# Build Rust dependencies
WORKDIR /app/rust
RUN cargo fetch

# Build Python package
WORKDIR /app/python
RUN pip3 install --upgrade pip maturin
RUN maturin build --release

# Install Python benchmark dependencies
RUN pip3 install pytest pytest-benchmark hypothesis numpy matplotlib jinja2

# =============================================================================
# Stage 3: Runtime image
# =============================================================================
FROM base AS runtime

WORKDIR /app

# Install Python runtime dependencies
RUN pip3 install --no-cache-dir \
    pytest \
    pytest-benchmark \
    hypothesis \
    numpy \
    matplotlib \
    jinja2 \
    fastapi \
    uvicorn[standard] \
    python-multipart

# Copy built artifacts
COPY --from=builder /app/python/target/wheels/*.whl /tmp/
RUN pip3 install /tmp/*.whl && rm /tmp/*.whl

# Copy source code
COPY go /app/go
COPY rust /app/rust
COPY python /app/python
COPY benchmarks /app/benchmarks

# Pre-build Rust (for faster benchmark startup)
WORKDIR /app/rust
RUN cargo build --release

# Create results directory with proper permissions
RUN mkdir -p /app/results && chmod 777 /app/results

# Set environment variables
ENV BENCH_MODE=web
ENV BENCH_PURE_MODE=true
ENV RESULTS_DIR=/app/results
ENV PORT=8080
ENV REGRESSION_THRESHOLD=10

# Expose port for web dashboard
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python3 -c "import requests; requests.get('http://localhost:8080/health')" || exit 1

# Entrypoint script
COPY benchmarks/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
