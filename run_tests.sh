#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

run_merge_test() {
    local TEMP_OUTPUT="test/temp/merged.asc"
    local EXPECTED_OUTPUT="test/merged.asc"
    local INPUT_DIR="test/split"

    mkdir -p "$(dirname "$TEMP_OUTPUT")"

    echo "Running merge test..."
    go run main.go merge -input_dir "$INPUT_DIR" > "$TEMP_OUTPUT"

    echo "Comparing merged files..."
    if diff -q "$TEMP_OUTPUT" "$EXPECTED_OUTPUT"; then
        echo "✅ Merge Test PASSED: Files are identical."
    else
        echo "❌ Merge Test FAILED: Files are different."
        diff "$TEMP_OUTPUT" "$EXPECTED_OUTPUT"
        return 1
    fi
}

run_split_test() {
    local INPUT_FILE="test/merged.asc"
    local TEMP_OUTPUT_DIR="test/temp/split"
    local EXPECTED_OUTPUT_DIR="test/split"

    rm -rf "$TEMP_OUTPUT_DIR"
    mkdir -p "$TEMP_OUTPUT_DIR"

    echo "Running split test..."
    cat "$INPUT_FILE" | go run main.go split -nrows 2 -ncols 2 -output_dir "$TEMP_OUTPUT_DIR"

    echo "Comparing split directories..."
    if diff -r -q "$TEMP_OUTPUT_DIR" "$EXPECTED_OUTPUT_DIR"; then
        echo "✅ Split Test PASSED: Directories are identical."
    else
        echo "❌ Split Test FAILED: Directories are different."
        diff -r "$TEMP_OUTPUT_DIR" "$EXPECTED_OUTPUT_DIR"
        return 1
    fi
}

run_merge_test
run_split_test