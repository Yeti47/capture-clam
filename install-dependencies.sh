#!/bin/bash

# Setup script for capture-clam dependencies
# This script installs the required GStreamer libraries and tools

set -e

echo "🔧 Installing capture-clam dependencies..."

# Detect Linux distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "❌ Cannot detect Linux distribution"
    exit 1
fi

case $OS in
    ubuntu|debian|pop|linuxmint)
        echo "📦 Detected Debian/Ubuntu-based system"
        sudo apt-get update
        sudo apt-get install -y \
            gstreamer1.0-tools \
            gstreamer1.0-plugins-base \
            gstreamer1.0-plugins-good \
            gstreamer1.0-plugins-bad \
            gstreamer1.0-plugins-ugly \
            gstreamer1.0-pulseaudio \
            gstreamer1.0-alsa \
            gstreamer1.0-libav
        ;;
    fedora|rhel|centos)
        echo "📦 Detected Fedora/RHEL-based system"
        sudo dnf install -y \
            gstreamer1 \
            gstreamer1-plugins-base \
            gstreamer1-plugins-good \
            gstreamer1-plugins-bad-free \
            gstreamer1-plugins-ugly-free \
            gstreamer1-libav
        ;;
    arch|manjaro)
        echo "📦 Detected Arch-based system"
        sudo pacman -S --needed --noconfirm \
            gstreamer \
            gst-plugins-base \
            gst-plugins-good \
            gst-plugins-bad \
            gst-plugins-ugly \
            gst-libav
        ;;
    opensuse*|sles)
        echo "📦 Detected openSUSE/SLES system"
        sudo zypper install -y \
            gstreamer \
            gstreamer-plugins-base \
            gstreamer-plugins-good \
            gstreamer-plugins-bad \
            gstreamer-plugins-ugly \
            gstreamer-plugins-libav
        ;;
    *)
        echo "❌ Unsupported distribution: $OS"
        echo "Please install GStreamer manually:"
        echo "  - gstreamer1.0-tools"
        echo "  - gstreamer1.0-plugins-base"
        echo "  - gstreamer1.0-plugins-good"
        echo "  - gstreamer1.0-plugins-bad"
        echo "  - gstreamer1.0-plugins-ugly"
        echo "  - gstreamer1.0-pulseaudio or gstreamer1.0-alsa"
        exit 1
        ;;
esac

echo "✅ Dependencies installed successfully!"
echo ""
echo "You can now run capture-clam:"
echo "  ./capture-clam list-audio"
echo "  ./capture-clam list-video"
echo "  ./capture-clam run --audio-source <id>"
