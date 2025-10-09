#!/bin/bash

echo "=== Testing E2E Sandbox ==="
echo "Current directory: $(pwd)"
echo "Home directory: $HOME" 
echo "Treex binary location: $(which treex)"
echo ""

echo "Files in current directory:"
ls -la

echo ""
echo "Testing treex command:"
treex --help | head -3

echo ""
echo "=== Test completed successfully ==="