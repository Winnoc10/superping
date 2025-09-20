#!/bin/bash

echo "Building SuperPing..."
go build -o superping main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Run with: ./superping"
else
    echo "❌ Build failed"
    exit 1
fi