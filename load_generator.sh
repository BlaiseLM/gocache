#!/bin/bash

# Populates cache and generates load for Prometheus/Grafana monitoring

set -e

ADDR="${CACHE_ADDR:-localhost:8080}"
NUM_KEYS="${NUM_KEYS:-1000}"
NUM_REQUESTS="${NUM_REQUESTS:-0}"
HIT_RATIO="${HIT_RATIO:-0.8}"
WORKERS="${WORKERS:-10}"
DURATION="${DURATION:-0}"

print_usage() {
    cat << EOF
Usage: $0 [options]

This tool populates the cache and generates load for monitoring via Prometheus/Grafana.

Options:
  -a, --addr       Server address (default: localhost:8080)
  -k, --keys       Number of keys to populate (default: 1000)
  -r, --requests   Number of GET requests per worker (default: 0 = continuous)
  -h, --hit-ratio  Hit ratio 0.0-1.0 (default: 0.8)
  -w, --workers    Number of concurrent workers (default: 10)
  -d, --duration   Run for N seconds (default: 0 = continuous)
  --help           Show this help message

Examples:
  $0                              # Continuous mode
  $0 -d 300                       # Run for 5 minutes (300 seconds)
  $0 -r 10000                     # 10k requests per worker
  $0 -k 5000 -h 0.9 -w 20         # 5k keys, 90% hit ratio, 20 workers

Monitor metrics at http://localhost:9090/metrics

Press Ctrl+C to stop
EOF
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--addr)
            ADDR="$2"
            shift 2
            ;;
        -k|--keys)
            NUM_KEYS="$2"
            shift 2
            ;;
        -r|--requests)
            NUM_REQUESTS="$2"
            shift 2
            ;;
        -h|--hit-ratio)
            HIT_RATIO="$2"
            shift 2
            ;;
        -w|--workers)
            WORKERS="$2"
            shift 2
            ;;
        -d|--duration)
            DURATION="$2"
            shift 2
            ;;
        --help)
            print_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

IFS=':' read -r HOST PORT <<< "$ADDR"

echo "==================================="
echo "Cache Load Generator"
echo "==================================="
echo ""
echo "Server: $ADDR"
echo "Populating $NUM_KEYS keys"
echo "Workers: $WORKERS"
echo "Hit ratio: $HIT_RATIO"

if [ "$DURATION" -gt 0 ]; then
    echo "Mode: Time-based ($DURATION seconds)"
elif [ "$NUM_REQUESTS" -gt 0 ]; then
    echo "Mode: Request-based ($NUM_REQUESTS requests per worker)"
else
    echo "Mode: Continuous (Ctrl+C to stop)"
fi
echo ""

send_command() {
    local host=$1
    local port=$2
    local command=$3
    echo "$command" | nc -N "$host" "$port" 2>/dev/null || echo "ERROR"
}

populate_cache() {
    echo "Populating cache..."
    
    local count=0
    local progress_interval=$((NUM_KEYS / 10))
    if [ $progress_interval -lt 1 ]; then
        progress_interval=1
    fi
    
    for i in $(seq 0 $((NUM_KEYS - 1))); do
        key="key_$i"
        value="value_$i"
        
        result=$(send_command "$HOST" "$PORT" "SET $key $value")
        
        if [[ ! "$result" =~ "OK" ]]; then
            echo "ERROR: Failed to set $key"
            exit 1
        fi
        
        count=$((count + 1))
        if [ $((count % progress_interval)) -eq 0 ] || [ $count -eq $NUM_KEYS ]; then
            echo "  $count/$NUM_KEYS keys"
        fi
    done
    
    echo "✓ Populated $NUM_KEYS keys"
    echo ""
}

random_number() {
    local max=$1
    echo $((RANDOM % max))
}

should_hit() {
    local threshold=$(echo "$HIT_RATIO * 100" | bc | cut -d'.' -f1)
    local rand=$((RANDOM % 100))
    
    if [ $rand -lt $threshold ]; then
        return 0  # hit
    else
        return 1  # miss
    fi
}

worker() {
    local worker_id=$1
    local requests_sent=0
    local start_time=$(date +%s)
    
    while true; do
        if [ "$DURATION" -gt 0 ]; then
            current_time=$(date +%s)
            elapsed=$((current_time - start_time))
            if [ $elapsed -ge $DURATION ]; then
                break
            fi
        fi
        
        if [ "$NUM_REQUESTS" -gt 0 ] && [ $requests_sent -ge $NUM_REQUESTS ]; then
            break
        fi
        
        if should_hit; then
            key_num=$(random_number $NUM_KEYS)
            key="key_$key_num"
        else
            key_num=$(random_number $((NUM_KEYS * 10)))
            key="missing_key_$key_num"
        fi
        
        send_command "$HOST" "$PORT" "GET $key" > /dev/null
        
        requests_sent=$((requests_sent + 1))
    done
    
    echo "Worker $worker_id completed: $requests_sent requests"
}

cleanup() {
    echo ""
    echo "Stopping workers..."
    kill $(jobs -p) 2>/dev/null || true
    wait 2>/dev/null || true
    echo "✓ Load generation stopped"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Main execution
populate_cache

echo "Sending GET requests..."
echo ""
echo "Press Ctrl+C to stop"
echo ""

for i in $(seq 1 $WORKERS); do
    worker $i &
done

# Wait for all workers to complete
wait

echo ""
echo "✓ Load generation complete"
