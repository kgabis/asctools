#!/bin/bash

set -e

echo "Building asctools..."
go build -o ./asctools ./cmd

echo "Cleaning temp directory"
rm -rf test/temp

run_merge_test() {
    local TEMP_OUTPUT="test/temp/merged.asc"
    local EXPECTED_OUTPUT="test/merged.asc"
    local INPUT_DIR="test/split"

    mkdir -p "$(dirname "$TEMP_OUTPUT")"

    echo "Running merge test..."
    ./asctools merge -input_dir "$INPUT_DIR" > "$TEMP_OUTPUT"

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
    ./asctools split -nrows 2 -ncols 2 -output_dir "$TEMP_OUTPUT_DIR" < "$INPUT_FILE" 

    echo "Comparing split directories..."
    if diff -r -q "$TEMP_OUTPUT_DIR" "$EXPECTED_OUTPUT_DIR"; then
        echo "✅ Split Test PASSED: Directories are identical."
    else
        echo "❌ Split Test FAILED: Directories are different."
        diff -r "$TEMP_OUTPUT_DIR" "$EXPECTED_OUTPUT_DIR"
        return 1
    fi
}

run_asc2png_test() {
    local TEMP_OUTPUT="test/temp/1to9.png"
    local EXPECTED_OUTPUT="test/1to9.png"
    local INPUT_FILE="test/1to9.asc"

    mkdir -p "$(dirname "$TEMP_OUTPUT")"

    echo "Running asc2png test..."
    ./asctools asc2png -scaling_operation up -scale 100 < "$INPUT_FILE" > "$TEMP_OUTPUT"

    echo "Comparing asc2png output files..."
    if diff -q "$TEMP_OUTPUT" "$EXPECTED_OUTPUT"; then
        echo "✅ asc2png Test PASSED: Files are identical."
    else
        echo "❌ asc2png Test FAILED: Files are different."
        # diff will just report that binary files differ, which is enough.
        diff "$TEMP_OUTPUT" "$EXPECTED_OUTPUT" || true
        return 1
    fi
}

run_asc2stl_test() {
    local TEMP_OUTPUT="test/temp/1to9.stl"
    local EXPECTED_OUTPUT="test/1to9.stl"
    local INPUT_FILE="test/1to9.asc"

    mkdir -p "$(dirname "$TEMP_OUTPUT")"

    echo "Running asc2stl test..."
    ./asctools asc2stl -scale 100 < "$INPUT_FILE" > "$TEMP_OUTPUT"

    echo "Comparing asc2stl output files..."
    if diff -q "$TEMP_OUTPUT" "$EXPECTED_OUTPUT"; then
        echo "✅ asc2stl Test PASSED: Files are identical."
    else
        echo "❌ asc2stl Test FAILED: Files are different."
        # diff will just report that binary files differ, which is enough.
        diff "$TEMP_OUTPUT" "$EXPECTED_OUTPUT" || true
        return 1
    fi
}

run_crop_test() {
    local TEMP_OUTPUT="test/temp/1to9_cropped.asc"
    local EXPECTED_OUTPUT="test/1to9_cropped.asc"
    local INPUT_FILE="test/1to9.asc"

    mkdir -p "$(dirname "$TEMP_OUTPUT")"

    echo "Running crop test..."
    ./asctools crop -start_x 2 -start_y 2 -end_x 3 -end_y 3 < "$INPUT_FILE" > "$TEMP_OUTPUT"

    echo "Comparing crop output files..."
    if diff -q "$TEMP_OUTPUT" "$EXPECTED_OUTPUT"; then
        echo "✅ crop Test PASSED: Files are identical."
    else
        echo "❌ crop Test FAILED: Files are different."
        # diff will just report that binary files differ, which is enough.
        diff "$TEMP_OUTPUT" "$EXPECTED_OUTPUT" || true
        return 1
    fi
}

run_merge_test
run_split_test
run_asc2png_test
run_asc2stl_test
run_crop_test