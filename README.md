# Capture Clam 🦪

Lightweight tool for streaming video and audio from capture cards with ultra-low latency.

Perfect for playing games directly on your PC screen through a capture card while simultaneously streaming the gameplay via Discord or other screen-sharing apps.

## Features

- **Zero-latency preview** - Play games directly with no perceptible input lag
- **Stream-ready** - Share the preview window with Discord, OBS, or any screen-sharing app
- **Capture card support** - Works with any V4L2-compatible capture device
- **Clean architecture** - Thin CLI commands, pure business logic in service layer
- **Simple device selection** - List and select audio/video devices by index or name
- **Automatic cleanup** - Remove orphaned PulseAudio loopback modules

## Requirements

- Go 1.16 or later
- GStreamer 1.0 with plugins:
  - `gst-plugins-base`
  - `gst-plugins-good`
  - `gst-plugins-bad`
- PulseAudio
- V4L2 (Video4Linux2)

### Installing Dependencies

**Arch Linux:**
```bash
sudo pacman -S gstreamer gst-plugins-base gst-plugins-good gst-plugins-bad pulseaudio v4l-utils
```

**Ubuntu/Debian:**
```bash
sudo apt install gstreamer1.0-tools gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-pulseaudio pulseaudio v4l-utils
```

## Installation

```bash
git clone https://github.com/Yeti47/capture-clam.git
cd capture-clam
go build -o capture-clam ./cmd/capture-clam
```

## Usage

### List Available Devices

**Audio sources:**
```bash
./capture-clam list-audio
```

**Video devices:**
```bash
./capture-clam list-video
```

### Start Preview

**By device index:**
```bash
./capture-clam run --audio-source 6
```

**By device name:**
```bash
./capture-clam run --audio-source alsa_input.usb-Device.analog-stereo
```

**Specify both audio and video:**
```bash
./capture-clam run --audio-source 6 --video-source 1
```

### Cleanup

Remove orphaned PulseAudio loopback modules:
```bash
./capture-clam cleanup
```

## Architecture

```
.
├── cmd/                    # CLI commands (thin layer)
│   ├── cleanup.go         # Cleanup command
│   ├── list_audio.go      # List audio devices
│   ├── list_video.go      # List video devices
│   ├── root.go            # Root command
│   └── run.go             # Run capture preview
└── internal/
    └── capture/
        └── service.go     # Core business logic
```

**Design principles:**
- UI output lives in `cmd/` layer only
- Pure business logic in `internal/capture`
- Single service struct handles all capture operations

## Technical Details

**GStreamer Pipeline:**
- Video: `v4l2src` → MJPEG (60fps) → `jpegdec` → `videoconvert` → leaky queue → `autovideosink`
- Audio: `pulsesrc` → S16LE 48kHz stereo → `audioconvert` → `audioresample` → leaky queue → `autoaudiosink`
- Both sinks run with `sync=false` for zero-latency operation

## License

[MIT](LICENSE.md)
