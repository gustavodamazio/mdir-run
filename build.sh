#!/bin/bash

# Set the output directory
OUTPUT_DIR="bin"
mkdir -p $OUTPUT_DIR

# Array of target platforms
platforms=(
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "darwin/amd64"
  "darwin/arm64"
)

# Iterate over platforms and build
for platform in "${platforms[@]}"
do
  IFS="/" read -r GOOS GOARCH <<< "${platform}"
  output_name="${OUTPUT_DIR}/mdir-run-${GOOS}-${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    output_name="${output_name}.exe"
  fi

  echo "Building for $GOOS/$GOARCH..."
  env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name

  if [ $? -ne 0 ]; then
    echo "An error occurred while building for $GOOS/$GOARCH"
    exit 1
  fi
done

echo "Build complete. Binaries are located in the $OUTPUT_DIR directory."
