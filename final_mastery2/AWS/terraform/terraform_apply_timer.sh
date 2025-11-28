#!/usr/bin/env zsh
# terraform_apply_timer.sh
# Wrapper to run `terraform apply` and log duration and status to a CSV file.

set -euo pipefail

# Config
TF_CMD=${TF_CMD:-"terraform apply -auto-approve"}
LOG_FILE="tf_apply_log.csv"

# If in a subdirectory, put the log file relative to this script's dir
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
LOG_PATH="$SCRIPT_DIR/$LOG_FILE"

timestamp() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

human_duration() {
  # seconds to human HH:MM:SS
  local secs=$1
  printf '%02d:%02d:%02d' $((secs/3600)) $((secs%3600/60)) $((secs%60))
}

echo "Starting: $(timestamp)" >&2
START_NS=$(date +%s%N)
START_EPOCH=$(date +%s)

# Run terraform apply (allow user to override TF_CMD env var)
echo "Running: $TF_CMD" >&2
set +e
eval $TF_CMD
EXIT_CODE=$?
set -e

END_NS=$(date +%s%N)
END_EPOCH=$(date +%s)

# Compute durations
NS_DIFF=$((END_NS - START_NS))
# duration in seconds floating
DURATION_SEC=$(awk -v n=$NS_DIFF 'BEGIN{printf "%.3f", n/1e9}')
INT_SECS=$((END_EPOCH - START_EPOCH))
HUMAN=$(human_duration $INT_SECS)

echo "Finished: $(timestamp)" >&2
echo "Duration: ${DURATION_SEC}s (approx ${HUMAN})" >&2
echo "Exit code: ${EXIT_CODE}" >&2

# Ensure log header exists
if [ ! -f "$LOG_PATH" ]; then
  echo "timestamp,dir,command,duration_seconds,duration_hms,exit_code" > "$LOG_PATH"
fi

PWD_DIR=$(pwd -P)
echo "$(timestamp),${PWD_DIR},'${TF_CMD}',${DURATION_SEC},${HUMAN},${EXIT_CODE}" >> "$LOG_PATH"

if [ "$EXIT_CODE" -ne 0 ]; then
  echo "terraform apply failed (exit $EXIT_CODE). Log written to $LOG_PATH" >&2
  exit $EXIT_CODE
fi

echo "Log appended to: $LOG_PATH" >&2
exit 0
