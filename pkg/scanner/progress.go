package scanner

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ProgressTracker tracks and reports scan progress
type ProgressTracker struct {
	filesProcessed   atomic.Int64
	foldersProcessed atomic.Int64
	startTime        time.Time
	lastReportTime   time.Time
	interval         time.Duration
	mu               sync.Mutex
	stopChan         chan struct{}
	stopped          bool
}

// NewProgressTracker creates a new progress tracker with the specified interval
func NewProgressTracker(interval time.Duration) *ProgressTracker {
	return &ProgressTracker{
		startTime:      time.Now(),
		lastReportTime: time.Now(),
		interval:       interval,
		stopChan:       make(chan struct{}),
	}
}

// Start begins the progress reporting goroutine
func (pt *ProgressTracker) Start() {
	go pt.reportLoop()
}

// reportLoop periodically reports progress
func (pt *ProgressTracker) reportLoop() {
	ticker := time.NewTicker(pt.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pt.report()
		case <-pt.stopChan:
			return
		}
	}
}

// report displays current progress to console
func (pt *ProgressTracker) report() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	files := pt.filesProcessed.Load()
	folders := pt.foldersProcessed.Load()
	elapsed := time.Since(pt.startTime)

	fmt.Printf("Progress: %d files, %d folders processed (%s elapsed)\n",
		files, folders, formatDuration(elapsed))
}

// IncrementFiles increments the files processed counter
func (pt *ProgressTracker) IncrementFiles() {
	pt.filesProcessed.Add(1)
}

// IncrementFolders increments the folders processed counter
func (pt *ProgressTracker) IncrementFolders() {
	pt.foldersProcessed.Add(1)
}

// Stop stops the progress reporting and displays final summary
func (pt *ProgressTracker) Stop() {
	pt.mu.Lock()
	if pt.stopped {
		pt.mu.Unlock()
		return
	}
	pt.stopped = true
	pt.mu.Unlock()

	close(pt.stopChan)
	time.Sleep(100 * time.Millisecond) // Allow goroutine to exit

	// Display final summary
	files := pt.filesProcessed.Load()
	folders := pt.foldersProcessed.Load()
	elapsed := time.Since(pt.startTime)

	fmt.Printf("\nScan complete: %d files, %d folders in %s\n",
		files, folders, formatDuration(elapsed))
}

// GetStats returns current progress statistics
func (pt *ProgressTracker) GetStats() (files int64, folders int64, elapsed time.Duration) {
	return pt.filesProcessed.Load(), pt.foldersProcessed.Load(), time.Since(pt.startTime)
}

// formatDuration formats a duration in human-readable form
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
