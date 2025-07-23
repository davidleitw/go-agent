#!/bin/bash

echo "🚀 Starting Multi-Tool Agent Demo"
echo "=================================="
echo ""

# Make sure we're in the right directory
cd "$(dirname "$0")"

# Check if binary exists, build if not
if [ ! -f "./multi-tool-demo" ]; then
    echo "🔨 Building demo..."
    go build -o multi-tool-demo
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
fi

echo "✅ Demo built successfully"
echo ""
echo "📝 Test commands available in test_commands.txt"
echo "💡 Try: 'What's the weather in Tokyo?' or 'Calculate 5 * 7'"
echo ""

# Run the demo
./multi-tool-demo