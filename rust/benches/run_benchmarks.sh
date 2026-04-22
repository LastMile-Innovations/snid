#!/bin/bash
# Enhanced benchmark runner with human and AI readable output

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
RUST_DIR="$PROJECT_ROOT/rust"
OUTPUT_DIR="$RUST_DIR/target/criterion"

# Create output directory for reports
REPORT_DIR="$RUST_DIR/benchmark_reports"
mkdir -p "$REPORT_DIR"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/benchmark_$TIMESTAMP.json"
SUMMARY_FILE="$REPORT_DIR/summary_$TIMESTAMP.txt"

echo -e "${CYAN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║           SNID Rust Benchmark Suite - Enhanced Output          ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Timestamp:${NC} $(date)"
echo -e "${BLUE}Report Directory:${NC} $REPORT_DIR"
echo ""

# Track start time
START_TIME=$(date +%s)

# Run benchmarks and collect Criterion output.
echo -e "${YELLOW}Running benchmarks...${NC}"
echo ""

cd "$RUST_DIR"

# Run each benchmark suite
BENCHMARKS=("benchmarks" "extended_families" "ecosystem_benchmarks")
TOTAL_BENCHES=${#BENCHMARKS[@]}
CURRENT_BENCH=0

for bench in "${BENCHMARKS[@]}"; do
    CURRENT_BENCH=$((CURRENT_BENCH + 1))
    echo -e "${CYAN}[$CURRENT_BENCH/$TOTAL_BENCHES] Running: $bench${NC}"
    
    # Run with standard criterion output
    if cargo bench --bench "$bench" 2>&1 | tee -a "$REPORT_DIR/raw_output_$TIMESTAMP.log"; then
        echo -e "${GREEN}✓ $bench completed${NC}"
    else
        echo -e "${RED}✗ $bench failed${NC}"
    fi
    echo ""
done

# Calculate elapsed time
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
MINUTES=$((ELAPSED / 60))
SECONDS=$((ELAPSED % 60))

echo -e "${GREEN}All benchmarks completed in ${MINUTES}m ${SECONDS}s${NC}"
echo ""

# Parse criterion output and generate reports
echo -e "${YELLOW}Generating reports...${NC}"

# Generate human-readable summary
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}                    BENCHMARK SUMMARY                           ${NC}"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"
echo "" > "$SUMMARY_FILE"
{
    echo "Reference baseline before this pass:"
    echo "- snid_new_fast: ~6.5 ns"
    echo "- snid_to_wire_mat: ~212 ns"
    echo "- snid_to_uuid_string: ~29 ns"
    echo "- lid_from_parts: ~600 ns"
    echo ""
} >> "$SUMMARY_FILE"

# Parse the raw output to extract benchmark results
echo -e "${BLUE}Benchmark Results:${NC}"
echo "" | tee -a "$SUMMARY_FILE"

# Initialize JSON structure
echo "{" > "$REPORT_FILE"
echo "  \"timestamp\": \"$(date -Iseconds)\"," >> "$REPORT_FILE"
echo "  \"elapsed_seconds\": $ELAPSED," >> "$REPORT_FILE"
echo "  \"benchmarks\": [" >> "$REPORT_FILE"

RAW_LOG="$REPORT_DIR/raw_output_$TIMESTAMP.log"
TEMP_RESULTS="$REPORT_DIR/temp_results_$TIMESTAMP.txt"
> "$TEMP_RESULTS"

# Parse benchmark results from the raw log to temp file
# Criterion format: benchmark name on its own line OR same line as "time:"
CURRENT_BENCH=""
CURRENT_CHANGE=""
CURRENT_THRPT=""
while IFS= read -r line; do
    # Skip "Benchmarking" lines with colons (Warming up, Collecting, Analyzing)
    if [[ $line =~ ^Benchmarking\ .+: ]]; then
        continue
    fi
    
    # Extract benchmark name - Criterion prints group/function on a line by itself.
    if [[ $line =~ ^[A-Za-z0-9_]+/[A-Za-z0-9_]+$ ]]; then
        CURRENT_BENCH="$line"
    fi
    
    # Extract change detection
    if [[ $line =~ Performance\ has\ improved\. ]]; then
        CURRENT_CHANGE="improved"
    elif [[ $line =~ Performance\ has\ regressed\. ]]; then
        CURRENT_CHANGE="regressed"
    elif [[ $line =~ No\ change\ in\ performance\ detected\. ]]; then
        CURRENT_CHANGE="no_change"
    elif [[ $line =~ Change\ within\ noise\ threshold\. ]]; then
        CURRENT_CHANGE="noise"
    fi
    
    # Extract throughput
    if [[ $line =~ ^\ *thrpt:\ +\[([0-9.]+)\ ([a-zμMGT]+/s)\ +([0-9.]+)\ ([a-zμMGT]+/s)\ +([0-9.]+)\ ([a-zμMGT]+/s)\] ]]; then
        THRPT_MEAN="${BASH_REMATCH[3]}"
        THRPT_UNIT="${BASH_REMATCH[4]}"
        CURRENT_THRPT="${THRPT_MEAN} ${THRPT_UNIT}"
    fi
    
    # Extract timing statistics - line starts with whitespace then "time:"
    if [[ $line =~ ^\ *time:\ +\[([0-9.]+)\ ([a-zμs]+)\ +([0-9.]+)\ ([a-zμs]+)\ +([0-9.]+)\ ([a-zμs]+)\] ]]; then
        MIN_VAL="${BASH_REMATCH[1]}"
        MIN_UNIT="${BASH_REMATCH[2]}"
        MEAN_VAL="${BASH_REMATCH[3]}"
        MEAN_UNIT="${BASH_REMATCH[4]}"
        MAX_VAL="${BASH_REMATCH[5]}"
        MAX_UNIT="${BASH_REMATCH[6]}"
        
        # Use mean value for display
        VALUE=$MEAN_VAL
        UNIT=$MEAN_UNIT
        
        # Convert to seconds for JSON
        case "$UNIT" in
            "ps")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000000000000" | bc)
                DISPLAY="${VALUE} ps"
                ;;
            "ns")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000000000" | bc)
                DISPLAY="${VALUE} ns"
                ;;
            "μs"|"us")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000000" | bc)
                DISPLAY="${VALUE} μs"
                ;;
            "ms")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000" | bc)
                DISPLAY="${VALUE} ms"
                ;;
            "s")
                SECONDS_VAL=$VALUE
                DISPLAY="${VALUE} s"
                ;;
            *)
                SECONDS_VAL=$VALUE
                DISPLAY="${VALUE} ${UNIT}"
                ;;
        esac
        
        if [ -n "$CURRENT_BENCH" ]; then
            # Format change indicator
            CHANGE_IND=""
            if [ "$CURRENT_CHANGE" = "improved" ]; then
                CHANGE_IND="${GREEN}↑${NC}"
            elif [ "$CURRENT_CHANGE" = "regressed" ]; then
                CHANGE_IND="${RED}↓${NC}"
            elif [ "$CURRENT_CHANGE" = "no_change" ]; then
                CHANGE_IND="${YELLOW}→${NC}"
            elif [ "$CURRENT_CHANGE" = "noise" ]; then
                CHANGE_IND="${CYAN}~${NC}"
            else
                CHANGE_IND="${CYAN}?${NC}"
            fi
            
            echo -e "  ${GREEN}✓${NC} $CURRENT_BENCH: ${DISPLAY} (mean) ${CHANGE_IND} ${CURRENT_THRPT}" | tee -a "$SUMMARY_FILE"
            
            # Save to temp file for JSON generation
            echo "$CURRENT_BENCH|$VALUE|$UNIT|$SECONDS_VAL|$CURRENT_CHANGE|$CURRENT_THRPT" >> "$TEMP_RESULTS"
            CURRENT_BENCH=""
            CURRENT_CHANGE=""
            CURRENT_THRPT=""
        fi
    fi
    
    # Handle case where benchmark name is on same line as time:
    if [[ $line =~ ^([A-Za-z0-9_]+/[A-Za-z0-9_]+)\ +time:\ +\[([0-9.]+)\ ([a-zμs]+)\ +([0-9.]+)\ ([a-zμs]+)\ +([0-9.]+)\ ([a-zμs]+)\] ]]; then
        CURRENT_BENCH="${BASH_REMATCH[1]}"
        MIN_VAL="${BASH_REMATCH[2]}"
        MIN_UNIT="${BASH_REMATCH[3]}"
        MEAN_VAL="${BASH_REMATCH[4]}"
        MEAN_UNIT="${BASH_REMATCH[5]}"
        MAX_VAL="${BASH_REMATCH[6]}"
        MAX_UNIT="${BASH_REMATCH[7]}"
        
        VALUE=$MEAN_VAL
        UNIT=$MEAN_UNIT
        
        # Convert to seconds for JSON
        case "$UNIT" in
            "ps")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000000000000" | bc)
                DISPLAY="${VALUE} ps"
                ;;
            "ns")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000000000" | bc)
                DISPLAY="${VALUE} ns"
                ;;
            "μs"|"us")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000000" | bc)
                DISPLAY="${VALUE} μs"
                ;;
            "ms")
                SECONDS_VAL=$(echo "scale=15; $VALUE / 1000" | bc)
                DISPLAY="${VALUE} ms"
                ;;
            "s")
                SECONDS_VAL=$VALUE
                DISPLAY="${VALUE} s"
                ;;
            *)
                SECONDS_VAL=$VALUE
                DISPLAY="${VALUE} ${UNIT}"
                ;;
        esac
        
        CHANGE_IND="${CYAN}?${NC}"
        echo -e "  ${GREEN}✓${NC} $CURRENT_BENCH: ${DISPLAY} (mean) ${CHANGE_IND}" | tee -a "$SUMMARY_FILE"
        
        # Save to temp file for JSON generation
        echo "$CURRENT_BENCH|$VALUE|$UNIT|$SECONDS_VAL|unknown|" >> "$TEMP_RESULTS"
        CURRENT_BENCH=""
    fi
done < "$RAW_LOG"

# Generate JSON from temp file
FIRST=true
while IFS='|' read -r NAME VALUE UNIT SECONDS_VAL CHANGE THRPT; do
    if [ "$FIRST" = true ]; then
        FIRST=false
    else
        echo "," >> "$REPORT_FILE"
    fi
    echo "    {" >> "$REPORT_FILE"
    echo "      \"name\": \"$NAME\"," >> "$REPORT_FILE"
    echo "      \"value\": $VALUE," >> "$REPORT_FILE"
    echo "      \"unit\": \"$UNIT\"," >> "$REPORT_FILE"
    echo "      \"mean_seconds\": $SECONDS_VAL," >> "$REPORT_FILE"
    echo "      \"change\": \"$CHANGE\"," >> "$REPORT_FILE"
    echo "      \"throughput\": \"$THRPT\"" >> "$REPORT_FILE"
    echo "    }" >> "$REPORT_FILE"
done < "$TEMP_RESULTS"

echo "" >> "$REPORT_FILE"
echo "  ]" >> "$REPORT_FILE"
echo "}" >> "$REPORT_FILE"

echo "" | tee -a "$SUMMARY_FILE"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}" | tee -a "$SUMMARY_FILE"
echo -e "${BLUE}Performance Summary Table:${NC}" | tee -a "$SUMMARY_FILE"
echo "" | tee -a "$SUMMARY_FILE"
printf "%-45s %12s %12s %10s\n" "Benchmark" "Mean Time" "Throughput" "Change" | tee -a "$SUMMARY_FILE"
printf "%-45s %12s %12s %10s\n" "---------" "---------" "---------" "------" | tee -a "$SUMMARY_FILE"

# Generate table from temp file
while IFS='|' read -r NAME VALUE UNIT SECONDS_VAL CHANGE THRPT; do
    # Format change for table
    case "$CHANGE" in
        "improved") CHANGE_SYM="↑" ;;
        "regressed") CHANGE_SYM="↓" ;;
        "no_change") CHANGE_SYM="→" ;;
        "noise") CHANGE_SYM="~" ;;
        *) CHANGE_SYM="?" ;;
    esac
    
    # Format throughput (empty if not available)
    if [ -z "$THRPT" ]; then
        THRPT_DISPLAY="N/A"
    else
        THRPT_DISPLAY="$THRPT"
    fi
    
    printf "%-45s %10s %2s %12s %10s\n" "$NAME" "$VALUE" "$UNIT" "$THRPT_DISPLAY" "$CHANGE_SYM" | tee -a "$SUMMARY_FILE"
done < "$TEMP_RESULTS"

# Clean up temp file
rm -f "$TEMP_RESULTS"

echo "" | tee -a "$SUMMARY_FILE"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}" | tee -a "$SUMMARY_FILE"
echo -e "${BLUE}Reports generated:${NC}" | tee -a "$SUMMARY_FILE"
echo -e "  • JSON: $REPORT_FILE" | tee -a "$SUMMARY_FILE"
echo -e "  • Summary: $SUMMARY_FILE" | tee -a "$SUMMARY_FILE"
echo -e "  • Raw log: $REPORT_DIR/raw_output_$TIMESTAMP.log" | tee -a "$SUMMARY_FILE"
echo -e "${CYAN}════════════════════════════════════════════════════════════════${NC}"

# Display JSON summary if available
if [ -f "$REPORT_FILE" ] && command -v jq &> /dev/null; then
    echo ""
    echo -e "${BLUE}JSON Summary (AI-readable):${NC}"
    jq '.' "$REPORT_FILE"
fi

echo ""
echo -e "${GREEN}Benchmark run complete!${NC}"
