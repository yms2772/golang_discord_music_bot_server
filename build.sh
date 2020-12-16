#!/bin/bash
echo "Builing wasm..."
GOARCH=wasm GOOS=js go build -v -o ../toy/web/app.wasm
