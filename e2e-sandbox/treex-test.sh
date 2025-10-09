#!/bin/bash

echo "=== Testing Treex E2E Functionality ==="
echo "Current directory: $(pwd)"
echo "Files created:"
ls -la

echo ""
echo "=== Running treex on created filesystem ==="
treex

echo ""
echo "=== Testing treex with depth limit ==="
treex -l 2

echo ""
echo "=== Testing treex directories only ==="
treex -d

echo ""
echo "=== Testing treex with hidden files ==="
treex -h

echo ""
echo "=== Test completed successfully ==="