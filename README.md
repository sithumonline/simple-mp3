# Simple MP3 Player

A cross-platform MP3 player built in Go featuring real-time audio visualization and generative art rendering driven by audio levels.

## Demo

https://github.com/user-attachments/assets/00e70f26-e377-4d07-8a5d-18095f0de1c4

## Architecture

This application implements a modular architecture separating audio processing, GUI rendering, and visualization components:

- **Audio Engine**: Built on the `faiface/beep` library for MP3 decoding and playback
- **GUI Framework**: Utilizes `hajimehoshi/ebiten/v2` for cross-platform windowing and `guigui` for widget management
- **Visualization**: Real-time generative art using `jdxyw/generativeart` synchronized with audio levels
- **Level Metering**: Custom RMS-based audio level detection with exponential moving average smoothing

## Features

### Audio Playback

- Multi-file playlist support with sequential playback
- Play/pause/next/previous track controls
- Volume adjustment with ±0.5 increment steps
- Automatic track advancement
- Robust error handling for file access and format validation

### Real-time Visualization

- Audio-reactive generative art rendering at ~14 FPS
- RMS-based level metering with configurable smoothing (α = 0.2)
- Dynamic visual patterns synchronized to audio amplitude
- Optimized rendering pipeline with throttled updates

### User Interface

- Minimal, responsive GUI with fixed 350x300 window dimensions
- Unicode symbol-based controls (⏸ ▶ ⏮ ⏭)
- Real-time track information with middle ellipsis truncation
- Custom font loading support (Noto Sans Symbols 2)

## Technical Specifications

### Dependencies

```
go 1.24.3
github.com/faiface/beep v1.1.0         // Audio processing
github.com/hajimehoshi/ebiten/v2        // Graphics engine
github.com/hajimehoshi/guigui           // GUI framework
github.com/jdxyw/generativeart         // Visualization engine
golang.org/x/text                       // Font handling
```

### Audio Processing

- **Sample Rate**: 44.1 kHz
- **Buffer Size**: 4410 samples (~100ms latency)
- **Format Support**: MP3 via go-mp3 decoder
- **Level Detection**: Real-time RMS calculation with EMA smoothing

### Performance Characteristics

- **Memory**: Streaming audio processing (no full file buffering)
- **CPU**: Optimized with 70ms visualization update intervals
- **Threading**: Concurrent audio playback and GUI rendering

## Installation

### Prerequisites

```bash
go version  # Requires Go 1.24.3+
```

### Build from Source

```bash
git clone https://github.com/sithumonline/simple-mp3.git
cd simple-mp3
go mod download
go build -o mp3player main.go
```

## Usage

### Basic Execution

```bash
go run main.go <file1.mp3> [file2.mp3] [file3.mp3...]
```

### Example

```bash
go run main.go ~/Downloads/No\ promises.mp3 ~/Downloads/Celine_Dion_-_My_Heart_Will_Go_On_Titanic_Song_Inpetto_Mix_djgeru.prv.pl_\(ge.mp3 ~/Downloads/Pihitak_Nathi_-_Api_Asarana_Wela_Wage_Gunadasa_Kapuge_Sarigama_lk.mp3
```

### Controls

- **Space**: Play/Pause toggle
- **←/→**: Previous/Next track navigation
- **+/-**: Volume adjustment
- **Mouse**: Interactive GUI controls

## Implementation Details

### Audio Pipeline

```
MP3 File → Decoder → Volume Control → Level Meter → Speaker Output
                                   ↓
                            Visualization Engine
```

### Level Metering Algorithm

The application implements real-time audio level detection using:

1. **RMS Calculation**: Root Mean Square across stereo channels
2. **Exponential Moving Average**: Smoothing with α=0.2 for stable visualization
3. **Normalized Output**: 0.0-1.0 range for consistent visual mapping

### Generative Art Engine

- **Canvas**: 640x390 pixel rendering surface
- **Update Rate**: 70ms intervals (14.3 FPS)
- **Color Palette**: Dynamic generation based on audio levels
- **Algorithms**: Configurable generative art patterns from the generativeart library

## File Structure

```
├── main.go                 # Application entry point and GUI setup
├── go.mod                  # Go module dependencies
├── NotoSansSymbols2-Regular.ttf  # Unicode symbol font
├── pkg/
│   ├── beeeep.go          # Audio playback engine
│   ├── levelmeter.go      # Real-time audio level detection
│   ├── utiles.go          # Utility functions
│   └── art/
│       └── ga_art.go      # Generative art visualization
```

## License

This project is open source. Check the repository for license details.

## Contributing

Contributions are welcome. Please ensure:

- Code follows Go conventions and formatting
- Audio processing changes maintain real-time performance
- GUI modifications preserve cross-platform compatibility
- Include performance benchmarks for audio pipeline changes
