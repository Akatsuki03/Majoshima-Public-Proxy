#!/bin/bash

# Kill any leftover processes from previous runs
pkill -f "./new-api" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true
sleep 1

# Build Go backend if binary doesn't exist
if [ ! -f "./new-api" ]; then
  echo "Building Go backend..."
  go build -o new-api . 2>&1
  echo "Backend build complete"
fi

# Start Go backend in background
PORT=3000 ./new-api &
BACKEND_PID=$!

echo "Backend started (PID: $BACKEND_PID) on port 3000"

# Start rsbuild dev server (frontend) in foreground on port 5000
cd web && bun run dev
