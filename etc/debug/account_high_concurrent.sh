#!/usr/bin/env bash
#
# Reproduce the "create_drop_account_high" CI test locally.
#
# CI config: 50 CREATE threads + 20 DROP threads, 5 min, index ∈ [1,10000]
#
# Usage:
#   ./etc/debug/account_high_concurrent.sh              # defaults
#   ./etc/debug/account_high_concurrent.sh -d 60        # run 60 seconds
#   ./etc/debug/account_high_concurrent.sh -p 6001      # custom port
#
# Prerequisites: mysql client, local MO running on 127.0.0.1:36001
#
# Safety: the script monitors system memory and worker errors in real time.
#   - If memory usage exceeds MEM_LIMIT (default 92%), test stops immediately.
#   - If any worker hits an error, test stops immediately.
# This avoids swap thrashing that would distort timing measurements.

set -uo pipefail

HOST="127.0.0.1"
PORT=36001
USER="dump"
PASS="111"
CREATE_VUSERS=50
DROP_VUSERS=20
DURATION=300          # seconds (5 min)
INDEX_MAX=10000
MEM_LIMIT=92          # percent, abort if exceeded

while getopts "h:p:u:w:c:r:d:m:M:" opt; do
    case $opt in
        h) HOST="$OPTARG" ;;
        p) PORT="$OPTARG" ;;
        u) USER="$OPTARG" ;;
        w) PASS="$OPTARG" ;;
        c) CREATE_VUSERS="$OPTARG" ;;
        r) DROP_VUSERS="$OPTARG" ;;
        d) DURATION="$OPTARG" ;;
        m) INDEX_MAX="$OPTARG" ;;
        M) MEM_LIMIT="$OPTARG" ;;
        *) echo "unknown option -$opt"; exit 1 ;;
    esac
done

TOTAL_VUSERS=$((CREATE_VUSERS + DROP_VUSERS))
TMPDIR=$(mktemp -d /tmp/mo-acct-test.XXXXXX)
trap 'rm -rf "$TMPDIR"' EXIT

# signal file: workers check this to know when to stop early
STOP_FILE="$TMPDIR/.stop"

echo "=== Account High-Concurrent Test ==="
echo "  host=$HOST  port=$PORT  user=$USER"
echo "  create_vusers=$CREATE_VUSERS  drop_vusers=$DROP_VUSERS"
echo "  duration=${DURATION}s  index_range=[1,$INDEX_MAX]"
echo "  mem_limit=${MEM_LIMIT}%"
echo "  tmpdir=$TMPDIR"
echo ""

MYSQL="mysql -h $HOST -P $PORT -u $USER -p$PASS --skip-ssl -N -s"

# verify connectivity
if ! $MYSQL -e "SELECT 1" >/dev/null 2>&1; then
    echo "ERROR: cannot connect to MO at $HOST:$PORT" >&2
    exit 1
fi
echo "Connection OK."

get_mem_pct() {
    awk '/MemTotal/{t=$2} /MemAvailable/{a=$2} END{printf "%d", (t-a)*100/t}' /proc/meminfo
}

DEADLINE=$(($(date +%s) + DURATION))

run_create() {
    local id=$1
    local ok=0 fail=0
    local logfile="$TMPDIR/create_${id}.log"
    while [ "$(date +%s)" -lt "$DEADLINE" ] && [ ! -f "$STOP_FILE" ]; do
        idx=$(( (RANDOM % INDEX_MAX) + 1 ))
        if $MYSQL -e "create account if not exists acount_new_${idx} admin_name 'admin' identified by '111';" 2>>"$logfile"; then
            ok=$((ok + 1))
        else
            fail=$((fail + 1))
            echo "[$(date '+%H:%M:%S')] create#$id FAIL idx=$idx" >> "$logfile"
            touch "$STOP_FILE"
            break
        fi
    done
    echo "$ok $fail" > "$TMPDIR/create_result_${id}"
}

run_drop() {
    local id=$1
    local ok=0 fail=0
    local logfile="$TMPDIR/drop_${id}.log"
    while [ "$(date +%s)" -lt "$DEADLINE" ] && [ ! -f "$STOP_FILE" ]; do
        idx=$(( (RANDOM % INDEX_MAX) + 1 ))
        if $MYSQL -e "drop account if exists acount_new_${idx};" 2>>"$logfile"; then
            ok=$((ok + 1))
        else
            fail=$((fail + 1))
            echo "[$(date '+%H:%M:%S')] drop#$id FAIL idx=$idx" >> "$logfile"
            touch "$STOP_FILE"
            break
        fi
    done
    echo "$ok $fail" > "$TMPDIR/drop_result_${id}"
}

echo "Starting $CREATE_VUSERS create workers + $DROP_VUSERS drop workers at $(date '+%H:%M:%S') ..."
echo "Will run until $(date -d @$DEADLINE '+%H:%M:%S') (${DURATION}s)"
echo ""

PIDS=()
for i in $(seq 1 $CREATE_VUSERS); do
    run_create "$i" &
    PIDS+=($!)
done
for i in $(seq 1 $DROP_VUSERS); do
    run_drop "$i" &
    PIDS+=($!)
done

STOP_REASON=""

# progress + memory guard + error guard
while [ "$(date +%s)" -lt "$DEADLINE" ] && [ -z "$STOP_REASON" ]; do
    alive=0
    for pid in "${PIDS[@]}"; do
        kill -0 "$pid" 2>/dev/null && alive=$((alive + 1)) || true
    done
    elapsed=$(( $(date +%s) - DEADLINE + DURATION ))
    mem_pct=$(get_mem_pct)
    echo -ne "\r  elapsed=${elapsed}s / ${DURATION}s  workers=${alive}/${TOTAL_VUSERS}  mem=${mem_pct}%  "

    if [ "$mem_pct" -ge "$MEM_LIMIT" ]; then
        STOP_REASON="MEMORY(${mem_pct}% >= ${MEM_LIMIT}%)"
    fi

    if [ -f "$STOP_FILE" ]; then
        STOP_REASON="WORKER_ERROR"
    fi

    sleep 3
done
echo ""

if [ -n "$STOP_REASON" ]; then
    echo "*** EARLY STOP: $STOP_REASON at $(date '+%H:%M:%S') ***"
    touch "$STOP_FILE"
    # give workers a moment to notice the stop signal
    sleep 2
fi

echo "Waiting for workers to finish ..."
for pid in "${PIDS[@]}"; do
    wait "$pid" 2>/dev/null || true
done

# aggregate
create_ok=0; create_fail=0
drop_ok=0; drop_fail=0
for f in "$TMPDIR"/create_result_*; do
    [ -f "$f" ] || continue
    read ok fail < "$f"
    create_ok=$((create_ok + ok))
    create_fail=$((create_fail + fail))
done
for f in "$TMPDIR"/drop_result_*; do
    [ -f "$f" ] || continue
    read ok fail < "$f"
    drop_ok=$((drop_ok + ok))
    drop_fail=$((drop_fail + fail))
done

create_total=$((create_ok + create_fail))
drop_total=$((drop_ok + drop_fail))
total=$((create_total + drop_total))
actual_dur=$(( $(date +%s) - DEADLINE + DURATION ))

echo ""
echo "=== Results ==="
if [ -n "$STOP_REASON" ]; then
    echo "  (stopped early: $STOP_REASON)"
fi
printf "  %-20s  %8s  %8s  %8s  %s\n" "Transaction" "Success" "Errors" "Total" "Rate"
if [ "$create_total" -gt 0 ]; then
    rate=$(echo "scale=4; $create_ok / $create_total" | bc)
else
    rate="N/A"
fi
printf "  %-20s  %8d  %8d  %8d  %s\n" "create_account" "$create_ok" "$create_fail" "$create_total" "$rate"
if [ "$drop_total" -gt 0 ]; then
    rate=$(echo "scale=4; $drop_ok / $drop_total" | bc)
else
    rate="N/A"
fi
printf "  %-20s  %8d  %8d  %8d  %s\n" "drop_account" "$drop_ok" "$drop_fail" "$drop_total" "$rate"
echo ""
echo "  Total operations: $total"
if [ "$actual_dur" -gt 0 ]; then
    echo "  Duration: ${actual_dur}s"
    echo "  TPS (approx): $(echo "scale=1; $total / $actual_dur" | bc)"
fi
echo ""

# show error samples
errors=$(cat "$TMPDIR"/*.log 2>/dev/null | head -20)
if [ -n "$errors" ]; then
    echo "=== Error samples (first 20 lines) ==="
    echo "$errors"
fi

echo ""
echo "Full logs in: $TMPDIR"
# keep tmpdir alive if there are errors or early stop
if [ "$((create_fail + drop_fail))" -gt 0 ] || [ -n "$STOP_REASON" ]; then
    trap - EXIT
    echo "(tmpdir preserved)"
fi
