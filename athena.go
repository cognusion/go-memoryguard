// Package memoryguard is a system to track the PSS memory usage of an os.Process
// and kill it if the usage exceeds the stated Limit.
package memoryguard

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cognusion/go-humanity"
)

// MemoryGuard is our encapsulating mechanation, and should only be acquired via a New helper.
// Member functions are goro-safe, but all struct attributes should be set immediatelyish after New(),
// and before Limit() is called.
type MemoryGuard struct {
	// Name is a name to use in lieu of PID for messaging
	Name string
	// Interval is a time.Duration to wait between checking usage
	Interval time.Duration
	// DebugOut is a logger for debug information
	DebugOut *log.Logger
	// ErrOut is a logger for StdErr coming from a process
	ErrOut *log.Logger
	// KillChan will be closed if/when the process is killed
	KillChan chan struct{}
	// StatsFrequency updates the internal frequency to which statistics are emitted to the debug logger. Default is 1 minute.
	StatsFrequency time.Duration

	cancelled chan bool
	nokill    bool        // Internal: true if the process should not be killed in overmemory cases
	running   atomic.Bool // Internal: true if the Limit goro is running.
	proc      *os.Process
	limit     atomic.Int64
	lastPss   atomic.Int64
	limiter   func()
}

// New takes an os.Process and returns a MemoryGuard for that process
func New(Process *os.Process) *MemoryGuard {
	var mg = MemoryGuard{
		proc:           Process,
		Interval:       1 * time.Second,
		KillChan:       make(chan struct{}),
		cancelled:      make(chan bool, 1),
		DebugOut:       log.New(io.Discard, "", 0),
		ErrOut:         log.New(io.Discard, "", 0),
		StatsFrequency: time.Minute,
	}
	mg.limiter = sync.OnceFunc(mg.onceLimit)

	return &mg
}

// PSS returns the last known PSS value for the watched process,
// or the current value, if there was no last value
func (m *MemoryGuard) PSS() int64 {
	if lp := m.lastPss.Load(); lp > 0 {
		return lp
	}
	pss, err := getPss(m.proc.Pid)
	if err != nil {
		return 0
	}
	return pss
}

// Cancel signals a Limit() operation to stop, returning immediately.
// After calling Cancel this MemoryGuard will be non-functional
func (m *MemoryGuard) Cancel() {
	select {
	case m.cancelled <- true:
		// cancelling
	default:
		// already cancelled
	}
}

// CancelWait signals a Limit() operation to stop, and waits to return until it is done.
// After calling CancelWait this MemoryGuard will be non-functional
func (m *MemoryGuard) CancelWait() {

	if !m.running.Load() {
		// We are already stopped.
		return
	}

	// Cancel, and poll until we're done.
	m.Cancel()
	for {
		if !m.running.Load() {
			return
		}
		time.Sleep(time.Millisecond) // too aggressive?
	}

}

// Limit takes the max usage (in Bytes) for the process and acts on the PSS.
// Returns an error if Limit is called with a zero or negative value,
// with a nil Process reference (did you use New()?),
// or if it has already been called once before, successfully.
func (m *MemoryGuard) Limit(max int64) error {
	if max <= 0 {
		return LimitZeroError
	} else if m.proc == nil {
		return LimitNilProcessError
	} else if !m.limit.CompareAndSwap(0, max) {
		return LimitOnceError
	}
	m.running.Store(true)

	go m.limiter()

	return nil
}

func (m *MemoryGuard) onceLimit() {
	defer func() {
		m.DebugOut.Print("MemoryGuard Limiter Leaving!\n")
		m.running.Store(false)
		m.lastPss.Store(0) // can't say
	}()

	var (
		name   = m.Name
		max    = m.limit.Load() // it should be impossible for this to be <= 0.
		errors int
	)
	if name == "" {
		name = fmt.Sprintf("%d", m.proc.Pid) // if proc hasn't been assigned, we panic here.
	}
	m.DebugOut.Printf("[%s] MemoryGuard Running! %v\n", name, m)

	since := time.Now()
	for {
		select {
		case <-m.cancelled:
			m.DebugOut.Printf("[%s] MemoryGuard Cancelled!\n", name)
			return
		case <-time.After(m.Interval):
			// Go for it
		}

		var (
			xss int64
			err error
		)

		xss, err = getPss(m.proc.Pid)
		if err != nil {
			errors++
			m.ErrOut.Printf("[%s] MemoryGuard getPss Error: %s (%d)\n", name, err, errors)
			continue
		} else {
			errors = 0 //reset
			m.lastPss.Store(xss)
		}

		if xss > max {
			m.ErrOut.Printf("[%s] MemoryGuard ALERT! %s Limit %s\n", name, humanity.ByteFormat(xss), humanity.ByteFormat(max))
			close(m.KillChan)
			if m.nokill {
				// don't kill it
			} else {
				// kill it
				m.proc.Kill()
			}
			m.running.Store(false)
			return
		} else if time.Since(since) >= m.StatsFrequency {
			// Belch out the stats every so often
			since = time.Now()
			m.DebugOut.Printf("[%s] MemoryGuard: %s Limit %s Consecutive errors: %d\n", name, humanity.ByteFormat(xss), humanity.ByteFormat(max), errors)
		}
	}
}

// getPss takes a pid, and returns the sum of PSS page sizes in Bytes, or an error
func getPss(pid int) (int64, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var (
		res int64
		pfx = []byte("Pss:")
	)

	r := bufio.NewScanner(f)
	for r.Scan() {
		line := r.Bytes()
		if bytes.HasPrefix(line, pfx) {
			var size int64
			_, err := fmt.Sscanf(string(line[4:]), "%d", &size)
			if err != nil {
				return 0, err
			}
			res += size
		}
	}
	if err := r.Err(); err != nil {
		return 0, err
	}

	return res * 1024, nil
}
