package pkg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Track struct {
	Path     string
	Stream   beep.StreamSeekCloser
	Format   beep.Format
	Ctrl     *beep.Ctrl
	Volume   *effects.Volume
	Meter    *LevelMeter
	Finished chan struct{}
}

type Player struct {
	mu        sync.Mutex
	Playlist  []string
	Idx       int
	Cur       *Track
	inited    bool
	batchSize int // speaker buffer size (samples)
}

func NewPlayer(paths []string) *Player {
	return &Player{
		Playlist:  paths,
		Idx:       0,
		batchSize: 44100 / 10, // ~100ms
	}
}

func (p *Player) openTrack(i int) (*Track, error) {
	if i < 0 || i >= len(p.Playlist) {
		return nil, fmt.Errorf("index out of bounds")
	}
	f, err := os.Open(p.Playlist[i])
	if err != nil {
		return nil, err
	}
	stream, format, err := mp3.Decode(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	// Initialize speaker once with the first file’s format.
	if !p.inited {
		if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second)/10); err != nil {
			stream.Close()
			f.Close()
			return nil, err
		}
		p.inited = true
	}
	ctrl := &beep.Ctrl{Streamer: stream, Paused: false}
	vol := &effects.Volume{Streamer: ctrl, Base: 2, Volume: 0.0} // 0dB
	meter := NewLevelMeter(vol, format.SampleRate)
	done := make(chan struct{}, 1)

	// Play and notify when finished.
	speaker.Play(beep.Seq(vol, beep.Callback(func() { close(done) })))

	return &Track{
		Path:     p.Playlist[i],
		Stream:   stream,
		Format:   format,
		Ctrl:     ctrl,
		Volume:   vol,
		Meter:    meter,
		Finished: done,
	}, nil
}

func (p *Player) closeCurrent() {
	if p.Cur != nil {
		_ = p.Cur.Stream.Close()
		p.Cur = nil
	}
}

func (p *Player) Play(i int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closeCurrent()

	t, err := p.openTrack(i)
	if err != nil {
		return err
	}
	p.Idx = i
	p.Cur = t
	fmt.Printf("▶️  Playing: %s\n", filepath.Base(p.Cur.Path))
	return nil
}

func (p *Player) PauseToggle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Cur == nil {
		return
	}
	speaker.Lock()
	p.Cur.Ctrl.Paused = !p.Cur.Ctrl.Paused
	state := "⏸️ Paused"
	if !p.Cur.Ctrl.Paused {
		state = "▶️ Resumed"
	}
	speaker.Unlock()
	fmt.Println(state)
}

func (p *Player) VolUp(db float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Cur == nil {
		return
	}
	speaker.Lock()
	p.Cur.Volume.Volume += db // in “log units” with Base=2 ≈ ~6dB per +1
	cur := p.Cur.Volume.Volume
	speaker.Unlock()
	fmt.Printf("🔊 Volume: %+0.2f\n", cur)
}

func (p *Player) VolDown(db float64) {
	p.VolUp(-db)
}

func (p *Player) Next() error {
	if len(p.Playlist) == 0 {
		return nil
	}
	n := (p.Idx + 1) % len(p.Playlist)
	return p.Play(n)
}

func (p *Player) Prev() error {
	if len(p.Playlist) == 0 {
		return nil
	}
	n := (p.Idx - 1 + len(p.Playlist)) % len(p.Playlist)
	return p.Play(n)
}

func (p *Player) RunAutoAdvance() {
	for {
		p.mu.Lock()
		t := p.Cur
		p.mu.Unlock()
		if t == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		// Wait for track to finish
		select {
		case <-t.Finished:
			// Autonext
			if err := p.Next(); err != nil {
				log.Println("advance error:", err)
			}
		default:
			time.Sleep(150 * time.Millisecond)
		}
	}
}

// pkg/player.go
func (p *Player) CurrentLevel() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Cur == nil || p.Cur.Meter == nil || (p.Cur.Ctrl != nil && p.Cur.Ctrl.Paused) {
		return 0
	}
	return p.Cur.Meter.Level()
}

func usage() {
	fmt.Println(`Commands:
  p            - toggle pause/play
  n            - Next track
  b            - previous track
  + / -        - volume up/down (step ~0.5)
  ls           - list playlist
  now          - show current track
  help         - show this help
  q            - quit`)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mp3player <file1.mp3> [file2.mp3 ...]")
		os.Exit(1)
	}
	// Basic existence check
	var files []string
	for _, a := range os.Args[1:] {
		if _, err := os.Stat(a); err != nil {
			log.Fatalf("cannot open %s: %v", a, err)
		}
		files = append(files, a)
	}

	player := NewPlayer(files)
	if err := player.Play(0); err != nil {
		log.Fatal(err)
	}
	go player.RunAutoAdvance()

	usage()
	sc := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		cmd := strings.TrimSpace(sc.Text())
		switch cmd {
		case "p":
			player.PauseToggle()
		case "n":
			if err := player.Next(); err != nil {
				fmt.Println("error:", err)
			}
		case "b":
			if err := player.Prev(); err != nil {
				fmt.Println("error:", err)
			}
		case "+":
			player.VolUp(0.5)
		case "-":
			player.VolDown(0.5)
		case "ls":
			for i, p := range player.Playlist {
				cur := " "
				if i == player.Idx {
					cur = "→"
				}
				fmt.Printf("%s %2d. %s\n", cur, i+1, filepath.Base(p))
			}
		case "now":
			player.mu.Lock()
			if player.Cur != nil {
				fmt.Println("Now:", filepath.Base(player.Cur.Path))
			} else {
				fmt.Println("Now: <none>")
			}
			player.mu.Unlock()
		case "help":
			usage()
		case "q", "quit", "exit":
			fmt.Println("bye!")
			return
		default:
			if cmd != "" {
				fmt.Println("unknown command. type 'help'.")
			}
		}
	}
}
