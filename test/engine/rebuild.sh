#!/bin/bash

# Game Engine Editor Build Script

echo "Building game editor..."
go build -o game-editor ./cmd/editor/

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Run with: ./game-editor"
else
    echo "❌ Build failed!"
    exit 1
fi