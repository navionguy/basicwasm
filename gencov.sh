#!/usr/bin/env bash
#set -e

# Define the output file for the combined coverage profile
COVERAGE_FILE="coverage.all"

# Start with a clean file and add the header
echo "mode: set" > "$COVERAGE_FILE"

# Iterate over each package and append coverage data
for pkg in *.out; do

    # Append the new coverage data, skipping the header line
    sed '1d' $pkg >> "$COVERAGE_FILE"
done

echo "Coverage report created: $COVERAGE_FILE"

# Optional: Display coverage and generate an HTML report
go tool cover -func="$COVERAGE_FILE"
go tool cover -html="$COVERAGE_FILE" -o coverage.html

