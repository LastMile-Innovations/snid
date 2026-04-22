#!/usr/bin/env python3
"""
SNID Benchmarking Platform - FastAPI Dashboard
Web interface for running and viewing benchmark results.
Optimized for Railway deployment with volume persistence.
"""

import os
import json
import asyncio
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Optional, List, Dict, Any

from fastapi import FastAPI, HTTPException, BackgroundTasks
from fastapi.responses import HTMLResponse, FileResponse, JSONResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel

# Configuration
RESULTS_DIR = Path(os.getenv("RESULTS_DIR", "/app/results"))
RESULTS_DIR.mkdir(parents=True, exist_ok=True)
REGRESSION_THRESHOLD = float(os.getenv("REGRESSION_THRESHOLD", "10"))

app = FastAPI(title="SNID Benchmarking Platform", version="2.0.0")


# =============================================================================
# Models
# =============================================================================

class BenchmarkStatus(BaseModel):
    status: str
    message: str
    timestamp: str


class BenchmarkResult(BaseModel):
    filename: str
    timestamp: str
    size: int
    summary: Optional[Dict[str, Any]] = None


# =============================================================================
# In-memory state for running benchmarks
# =============================================================================

running_benchmarks: Dict[str, Dict[str, Any]] = {}


# =============================================================================
# Helper Functions
# =============================================================================

def get_result_files() -> List[Path]:
    """Get all result JSON files, sorted by modification time (newest first)."""
    files = list(RESULTS_DIR.glob("*.json"))
    files.sort(key=lambda f: f.stat().st_mtime, reverse=True)
    return files


def parse_result_summary(filepath: Path) -> Dict[str, Any]:
    """Parse a result file and return a summary."""
    try:
        data = json.loads(filepath.read_text())
        return {
            "timestamp": data.get("timestamp"),
            "suites": list(data.get("suites", {}).keys()),
            "languages": list(data.get("languages", {}).keys()),
            "total_benchmarks": len(data.get("suites", {})) + len(data.get("languages", {})),
        }
    except (json.JSONDecodeError, KeyError):
        return {"error": "Could not parse file"}


# =============================================================================
# Benchmark Execution
# =============================================================================

async def run_benchmark_suite(suite: str, run_id: str):
    """Run a benchmark suite in the background with pure mode isolation."""
    try:
        running_benchmarks[run_id] = {
            "status": "running",
            "started_at": datetime.now().isoformat(),
            "suite": suite,
        }

        # Run the benchmark in pure mode (isolated subprocess, no harness overhead)
        env = os.environ.copy()
        env["BENCH_PURE_MODE"] = "true"  # Ensure pure mode for zero overhead
        
        cmd = ["python3", "/app/benchmarks/runner.py", suite]
        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd="/app",
            env=env
        )
        stdout, stderr = await process.communicate()

        running_benchmarks[run_id]["status"] = "completed"
        running_benchmarks[run_id]["completed_at"] = datetime.now().isoformat()
        running_benchmarks[run_id]["returncode"] = process.returncode
        running_benchmarks[run_id]["stdout"] = stdout.decode() if stdout else ""
        running_benchmarks[run_id]["stderr"] = stderr.decode() if stderr else ""

    except Exception as e:
        running_benchmarks[run_id]["status"] = "failed"
        running_benchmarks[run_id]["error"] = str(e)


# =============================================================================
# API Endpoints
# =============================================================================

@app.get("/", response_class=HTMLResponse)
async def dashboard():
    """Render the dashboard UI."""
    html = """
    <!DOCTYPE html>
    <html>
    <head>
        <title>SNID Benchmarking Platform</title>
        <style>
            body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; background: #f5f5f5; }
            .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
            h1 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
            .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
            .stat-card { background: #f8f9fa; padding: 20px; border-radius: 6px; border-left: 4px solid #007bff; }
            .stat-value { font-size: 2em; font-weight: bold; color: #007bff; }
            .stat-label { color: #666; font-size: 0.9em; }
            .suites { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin: 20px 0; }
            .suite-btn { padding: 15px; background: #007bff; color: white; border: none; border-radius: 6px; cursor: pointer; font-size: 1em; transition: background 0.2s; }
            .suite-btn:hover { background: #0056b3; }
            .suite-btn:disabled { background: #ccc; cursor: not-allowed; }
            .results { margin-top: 30px; }
            .result-item { padding: 15px; background: #f8f9fa; margin: 10px 0; border-radius: 6px; display: flex; justify-content: space-between; align-items: center; }
            .result-info { flex: 1; }
            .result-actions { display: flex; gap: 10px; }
            .btn { padding: 8px 16px; border-radius: 4px; text-decoration: none; font-size: 0.9em; }
            .btn-primary { background: #007bff; color: white; }
            .btn-secondary { background: #6c757d; color: white; }
            .status { padding: 4px 8px; border-radius: 4px; font-size: 0.8em; }
            .status-running { background: #ffc107; color: #333; }
            .status-completed { background: #28a745; color: white; }
            .status-failed { background: #dc3545; color: white; }
        </style>
    </head>
    <body>
        <div class="container">
            <h1>🚀 SNID Benchmarking Platform</h1>
            
            <div class="stats">
                <div class="stat-card">
                    <div class="stat-value" id="total-results">-</div>
                    <div class="stat-label">Total Results</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value" id="running-count">-</div>
                    <div class="stat-label">Running</div>
                </div>
            </div>

            <h2>Run Benchmarks</h2>
            <div class="suites">
                <button class="suite-btn" onclick="runBenchmark('go')">Go</button>
                <button class="suite-btn" onclick="runBenchmark('rust')">Rust</button>
                <button class="suite-btn" onclick="runBenchmark('python')">Python</button>
                <button class="suite-btn" onclick="runBenchmark('comparison')">Comparison</button>
                <button class="suite-btn" onclick="runBenchmark('llm')">LLM</button>
                <button class="suite-btn" onclick="runBenchmark('ecosystem')">Ecosystem</button>
                <button class="suite-btn" onclick="runBenchmark('all')">All Suites</button>
            </div>

            <h2>Recent Results</h2>
            <div class="results" id="results-list">
                <p>Loading...</p>
            </div>
        </div>

        <script>
            async function loadResults() {
                const response = await fetch('/results');
                const data = await response.json();
                
                document.getElementById('total-results').textContent = data.length;
                
                const resultsList = document.getElementById('results-list');
                if (data.length === 0) {
                    resultsList.innerHTML = '<p>No results yet. Run a benchmark to get started.</p>';
                    return;
                }
                
                resultsList.innerHTML = data.map(r => `
                    <div class="result-item">
                        <div class="result-info">
                            <strong>${r.filename}</strong><br>
                            <small>${r.timestamp}</small>
                        </div>
                        <div class="result-actions">
                            <a href="/results/${r.filename}" class="btn btn-primary">View</a>
                            <a href="/download/${r.filename}" class="btn btn-secondary">Download</a>
                        </div>
                    </div>
                `).join('');
            }

            async function runBenchmark(suite) {
                const runId = Date.now().toString();
                const response = await fetch(`/run/${suite}?run_id=${runId}`, { method: 'POST' });
                const data = await response.json();
                
                if (data.status === 'started') {
                    alert('Benchmark started! Check back in a few minutes.');
                    setTimeout(loadResults, 5000);
                } else {
                    alert('Error: ' + data.message);
                }
            }

            loadResults();
            setInterval(loadResults, 30000);
        </script>
    </body>
    </html>
    """
    return html


@app.get("/health")
async def health_check():
    """Health check endpoint for container orchestration."""
    return {"status": "healthy", "timestamp": datetime.now().isoformat()}


@app.get("/results")
async def list_results():
    """List all benchmark result files."""
    files = get_result_files()
    results = []
    for f in files:
        results.append({
            "filename": f.name,
            "timestamp": datetime.fromtimestamp(f.stat().st_mtime).isoformat(),
            "size": f.stat().st_size,
            "summary": parse_result_summary(f),
        })
    return results


@app.get("/results/{filename}")
async def get_result(filename: str):
    """Get a specific result file."""
    filepath = RESULTS_DIR / filename
    if not filepath.exists():
        raise HTTPException(status_code=404, detail="Result file not found")
    return JSONResponse(content=json.loads(filepath.read_text()))


@app.get("/download/{filename}")
async def download_result(filename: str):
    """Download a specific result file."""
    filepath = RESULTS_DIR / filename
    if not filepath.exists():
        raise HTTPException(status_code=404, detail="Result file not found")
    return FileResponse(filepath, media_type="application/json", filename=filename)


@app.post("/run/{suite}")
async def run_benchmark(suite: str, background_tasks: BackgroundTasks, run_id: Optional[str] = None):
    """Trigger a benchmark run."""
    valid_suites = ["go", "rust", "python", "comparison", "llm", "ecosystem", "all"]
    if suite not in valid_suites:
        raise HTTPException(status_code=400, detail=f"Invalid suite. Valid: {valid_suites}")

    if run_id is None:
        run_id = f"{suite}_{datetime.now().strftime('%Y%m%d_%H%M%S')}"

    # Start benchmark in background
    background_tasks.add_task(run_benchmark_suite, suite, run_id)

    return {
        "status": "started",
        "run_id": run_id,
        "suite": suite,
        "message": f"Benchmark suite '{suite}' started",
    }


@app.get("/status/{run_id}")
async def get_run_status(run_id: str):
    """Get the status of a running benchmark."""
    if run_id not in running_benchmarks:
        raise HTTPException(status_code=404, detail="Run ID not found")
    return running_benchmarks[run_id]


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=int(os.getenv("PORT", 8080)))
