#!/bin/bash
# Usage: ./run_diff.sh input_dir_1 input_dir_2 [diff_pow]
set -e

INPUT_DIR_1="$1"
INPUT_DIR_2="$2"
DIFF_POW="$3"

# Extract base names for file labeling
YEAR1=$(basename "$INPUT_DIR_1")
YEAR2=$(basename "$INPUT_DIR_2")

# Output file paths
OUTPUT_ASC_1="tests/${YEAR1}.asc"
OUTPUT_ASC_2="tests/${YEAR2}.asc"
OUTPUT_DIFF="tests/${YEAR1}_${YEAR2}_diff.png"

# Merge ASC files
go run main.go merge -input_dir "$INPUT_DIR_1" > "$OUTPUT_ASC_1" &
PID1=$!
go run main.go merge -input_dir "$INPUT_DIR_2" > "$OUTPUT_ASC_2" &
PID2=$!

wait $PID1
wait $PID2

# Generate diff PNG with optional --diff_pow
if [ -n "$DIFF_POW" ]; then
  go run main.go diffasc2png -input1 "$OUTPUT_ASC_1" -input2 "$OUTPUT_ASC_2" -diff_pow "$DIFF_POW" > "$OUTPUT_DIFF"
else
  go run main.go diffasc2png -input1 "$OUTPUT_ASC_1" -input2 "$OUTPUT_ASC_2" > "$OUTPUT_DIFF"
fi

echo "Diff image generated successfully: $OUTPUT_DIFF"