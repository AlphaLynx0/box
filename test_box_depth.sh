#!/usr/bin/env bash
#
# test_box_depth.sh
#
# Usage: ./test_box_with_timing.sh [-n START_DEPTH]
#
#   -n START_DEPTH   Number of boxes to apply before the first test.
#                    (Default: 1)
#
# This script will:
#   1. Build a pipeline of “echo 123 | ./box | ./box | …” with START_DEPTH boxes.
#   2. Time how long it takes to run that pipeline (construction + printing).
#   3. Print the boxed output and “(took X.XXXs)” at the footer.
#   4. Append one more “| ./box” each iteration, test again, measure time, etc.
#   5. Stop as soon as the pipeline fails (exit status ≠ 0).
#
# Author: AlphaLynx <alphalynx@protonmail.com>
# Date: June 4, 2025
# License: MIT

# -------------------------------------------------------------------
# Parse arguments
# -------------------------------------------------------------------
start_depth=1
while getopts ":n:" opt; do
  case $opt in
    n)
      if [[ "$OPTARG" =~ ^[0-9]+$ ]]; then
        start_depth="$OPTARG"
      else
        echo "Invalid value for -n: must be a non-negative integer."
        exit 1
      fi
      ;;
    \?)
      echo "Usage: $0 [-n START_DEPTH]"
      exit 1
      ;;
  esac
done

# -------------------------------------------------------------------
# Build initial pipeline string with start_depth copies of “| ./box”
# -------------------------------------------------------------------
pipeline="echo 123"
depth=0

# If start_depth is zero, we’ll still test “echo 123” (depth=0).
# If start_depth ≥ 1, we wrap “echo 123” in N boxes before testing.
for ((i = 0; i < start_depth; i++)); do
  pipeline="$pipeline | ./box"
  depth=$((depth + 1))
done

# -------------------------------------------------------------------
# Main loop: test “depth” boxes, then append one more each iteration.
# -------------------------------------------------------------------
while true; do
  echo
  echo "=== Output with $depth box(es) ==="

  # Record start time (seconds.nanoseconds)
  t_start=$(date +%s.%N)

  # Run the pipeline (lets stdout/stderr print to your terminal)
  if eval "$pipeline"; then
    t_end=$(date +%s.%N)
    # Calculate elapsed = t_end - t_start (using bc for floating arithmetic)
    elapsed=$(echo "$t_end - $t_start" | bc)

    echo "=== End of output for $depth box(es) (took ${elapsed}s) ==="
  else
    t_end=$(date +%s.%N)
    elapsed=$(echo "$t_end - $t_start" | bc)
    echo
    echo "!!! FAILURE at $depth box(es) (attempt took ${elapsed}s) !!!"
    break
  fi

  # Prepare for next iteration: add one more box
  pipeline="$pipeline | ./box"
  depth=$((depth + 1))
done

# The last successful depth is one less than the iteration that failed.
max_success=$((depth - 1))
echo
echo "→ Maximum successful boxes: $max_success"
exit 0

