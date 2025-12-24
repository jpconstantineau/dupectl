package scanner

import (
	"fmt"
	"sync"
	"time"
)

// ProgressIndicator displays real-time scan progress
type ProgressIndicator struct {
	startTime      time.Time
	foldersScanned int64
	filesScanned   int64
	filesHashed    int64
	mu             sync.RWMutex
	spinner        *Spinner
	enabled        bool
	interval       time.Duration
	ticker         *time.Ticker
	done           chan struct{}
}

// Spinner provides braille spinner animation
type Spinner struct {
	frames []rune
	index  int
}

var brailleSpinner = []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// NewProgressIndicator creates a progress indicator
func NewProgressIndicator(enabled bool, interval time.Duration) *ProgressIndicator {
	if interval <= 0 {
		interval = 10 * time.Second // Default to 10 seconds
	}
	return &ProgressIndicator{
		startTime: time.Now(),
		spinner:   &Spinner{frames: brailleSpinner},
		enabled:   enabled,
		interval:  interval,
		done:      make(chan struct{}),
	}
}

// Start begins displaying progress updates at configured interval
func (p *ProgressIndicator) Start() {
	if !p.enabled {
		return
	}

	p.ticker = time.NewTicker(p.interval)
	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.Display()
			case <-p.done:
				return
			}
		}
	}()
}

// Stop stops progress updates
func (p *ProgressIndicator) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.done)
}

// IncrementFolders increments folder count
func (p *ProgressIndicator) IncrementFolders() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.foldersScanned++
}

// IncrementFiles increments file count
func (p *ProgressIndicator) IncrementFiles() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.filesScanned++
}

// IncrementFilesHashed increments hashed file count
func (p *ProgressIndicator) IncrementFilesHashed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.filesHashed++
}

// Display prints current progress
func (p *ProgressIndicator) Display() {
	if !p.enabled {
		return
	}

	p.mu.RLock()
	folders := p.foldersScanned
	files := p.filesScanned
	hashed := p.filesHashed
	elapsed := time.Since(p.startTime)
	p.mu.RUnlock()

	frame := p.spinner.Next()
	if hashed > 0 {
		fmt.Printf("\r%c Folders: %d | Files: %d | Hashed: %d/%d | Elapsed: %s",
			frame, folders, files, hashed, files, formatDuration(elapsed))
	} else {
		fmt.Printf("\r%c Folders: %d | Files: %d | Elapsed: %s",
			frame, folders, files, formatDuration(elapsed))
	}
}

// Summary returns final statistics
func (p *ProgressIndicator) Summary() (folders, files int64, duration time.Duration) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.foldersScanned, p.filesScanned, time.Since(p.startTime)
}

// Next returns next spinner frame
func (s *Spinner) Next() rune {
	frame := s.frames[s.index]
	s.index = (s.index + 1) % len(s.frames)
	return frame
}

// formatDuration formats duration as "1m 30s"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
