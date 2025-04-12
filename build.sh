#!/bin/bash

# Set the output directory
OUTPUT_DIR="bin"
mkdir -p $OUTPUT_DIR

echo "====== Building for all platforms ======"

# 1. Build for macOS natively
echo "Building for macOS..."
mac_platforms=(
  "darwin/amd64"
  "darwin/arm64"
)

for platform in "${mac_platforms[@]}"
do
  IFS="/" read -r GOOS GOARCH <<< "${platform}"
  output_name="${OUTPUT_DIR}/mdir-run-${GOOS}-${GOARCH}"
  
  echo "Building for $GOOS/$GOARCH..."
  
  # Build using CGO for macOS
  env CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name

  if [ $? -ne 0 ]; then
    echo "An error occurred while building for $GOOS/$GOARCH"
    exit 1
  fi
done

# 2. Build for Linux using fyne-cross
echo ""
echo "==== Building for Linux using fyne-cross... ===="
fyne-cross linux -arch=amd64 -app-id github.com/gustavodamazio/mdir-run
fyne-cross linux -arch=arm64 -app-id github.com/gustavodamazio/mdir-run

# 3. Build for Windows using fyne-cross
echo ""
echo "==== Building for Windows using fyne-cross... ===="
fyne-cross windows -arch=amd64 -app-id github.com/gustavodamazio/mdir-run
fyne-cross windows -arch=arm64 -app-id github.com/gustavodamazio/mdir-run

# 4. Copy files from fyne-cross to bin folder
echo ""
echo "==== Moving files to bin folder... ===="

# Extract and copy Linux files
if [ -f "fyne-cross/dist/linux-amd64/mdir-run.tar.xz" ]; then
  mkdir -p temp_extract
  tar -xf fyne-cross/dist/linux-amd64/mdir-run.tar.xz -C temp_extract
  cp -f temp_extract/usr/local/bin/mdir-run $OUTPUT_DIR/mdir-run-linux-amd64
  rm -rf temp_extract
  echo "✅ Linux amd64 extracted and copied successfully"
else
  echo "❌ Error: Linux amd64 not found"
fi

if [ -f "fyne-cross/dist/linux-arm64/mdir-run.tar.xz" ]; then
  mkdir -p temp_extract
  tar -xf fyne-cross/dist/linux-arm64/mdir-run.tar.xz -C temp_extract
  cp -f temp_extract/usr/local/bin/mdir-run $OUTPUT_DIR/mdir-run-linux-arm64
  rm -rf temp_extract
  echo "✅ Linux arm64 extracted and copied successfully"
else
  echo "❌ Error: Linux arm64 not found"
fi

# Extract and copy Windows files
if [ -f "fyne-cross/dist/windows-amd64/mdir-run.exe.zip" ]; then
  mkdir -p temp_extract
  unzip -o fyne-cross/dist/windows-amd64/mdir-run.exe.zip -d temp_extract
  cp -f temp_extract/mdir-run.exe $OUTPUT_DIR/mdir-run-windows-amd64.exe
  rm -rf temp_extract
  echo "✅ Windows amd64 extracted and copied successfully"
else
  echo "❌ Error: Windows amd64 not found"
fi

if [ -f "fyne-cross/dist/windows-arm64/mdir-run.exe.zip" ]; then
  mkdir -p temp_extract
  unzip -o fyne-cross/dist/windows-arm64/mdir-run.exe.zip -d temp_extract
  cp -f temp_extract/mdir-run.exe $OUTPUT_DIR/mdir-run-windows-arm64.exe
  rm -rf temp_extract
  echo "✅ Windows arm64 extracted and copied successfully"
else
  echo "❌ Error: Windows arm64 not found"
fi

echo ""
echo "Build complete. Binaries available in folder $OUTPUT_DIR:"
ls -la $OUTPUT_DIR

echo ""
echo "Compilation for all platforms completed successfully!"
