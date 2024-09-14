#!/bin/bash

OUTPUT_DIR="bin"
PACKAGE_DIR="dist"
mkdir -p $PACKAGE_DIR

for file in $OUTPUT_DIR/*
do
  filename=$(basename "$file")
  if [[ "$filename" == *.exe ]]; then
    zip -j "$PACKAGE_DIR/${filename%.exe}.zip" "$file"
  else
    tar -czf "$PACKAGE_DIR/${filename}.tar.gz" -C "$OUTPUT_DIR" "$filename"
  fi
done

echo "Packaging complete. Archives are located in the $PACKAGE_DIR directory."
