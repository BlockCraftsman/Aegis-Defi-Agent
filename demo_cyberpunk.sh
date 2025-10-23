#!/bin/bash

echo "🚀 Launching Aegis Protocol - Cyberpunk Edition"
echo "=================================================="
echo ""

# Build the application
echo "🔧 Building cyberpunk interface..."
go build -o aegis-nexus ./cmd/aegis-nexus/

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "🎮 Starting cyberpunk terminal interface..."
    echo ""
    echo "💡 Features:"
    echo "   • Futuristic ASCII boot sequence"
    echo "   • Matrix-style background effects"
    echo "   • Neon cyberpunk color scheme"
    echo "   • Enhanced agent coordination interface"
    echo "   • Real-time market surveillance"
    echo ""
    echo "🎯 Controls:"
    echo "   • 1-5: Switch between views"
    echo "   • Ctrl+M: Toggle Matrix effect"
    echo "   • Ctrl+G: Toggle glitch effect"
    echo "   • Ctrl+P: Toggle pulse animation"
    echo "   • Ctrl+T: Toggle terminal mode"
    echo "   • Ctrl+C: Exit"
    echo ""
    echo "🚀 Launching in 3 seconds..."
    sleep 3
    
    # Run the application
    ./aegis-nexus
else
    echo "❌ Build failed!"
    exit 1
fi