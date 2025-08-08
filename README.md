

# memoryguard
`import "github.com/cognusion/go-memoryguard"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Examples](#pkg-examples)

## <a name="pkg-overview">Overview</a>
Package memoryguard is a system to track the PSS memory usage of an os.Process
and kill it if the usage exceeds the stated Limit.




## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [type Error](#Error)
  * [func (e Error) Error() string](#Error.Error)
* [type MemoryGuard](#MemoryGuard)
  * [func New(Process *os.Process) *MemoryGuard](#New)
  * [func (m *MemoryGuard) Cancel()](#MemoryGuard.Cancel)
  * [func (m *MemoryGuard) CancelWait()](#MemoryGuard.CancelWait)
  * [func (m *MemoryGuard) Limit(max int64) error](#MemoryGuard.Limit)
  * [func (m *MemoryGuard) PSS() int64](#MemoryGuard.PSS)

#### <a name="pkg-examples">Examples</a>
* [MemoryGuard](#example-memoryguard)
* [MemoryGuard.CancelWait](#example-memoryguard_cancelwait)

#### <a name="pkg-files">Package files</a>
[athena.go](https://github.com/cognusion/go-memoryguard/tree/master/athena.go) [errors.go](https://github.com/cognusion/go-memoryguard/tree/master/errors.go)


## <a name="pkg-constants">Constants</a>
``` go
const (
    // LimitZeroError is returned by Limit(int64) when the passed variable is <= 0.
    LimitZeroError = Error("please call Limit(int64) with a value greater than zero")
    // LimitNilProcessError is returned by Limit(int64) when the referenced *os.Process is nil.
    LimitNilProcessError = Error("a Process has not been created and assigned, or is nil")
    // LimitOnceError is returned by Limit(int64) if it has been called without error previously.
    LimitOnceError = Error("Limit(int64) already called once")
)
```




## <a name="Error">type</a> [Error](https://github.com/cognusion/go-memoryguard/tree/master/errors.go?s=558:575#L13)
``` go
type Error string
```
Error is an error type










### <a name="Error.Error">func</a> (Error) [Error](https://github.com/cognusion/go-memoryguard/tree/master/errors.go?s=627:656#L16)
``` go
func (e Error) Error() string
```
Error returns the stringified version of Error




## <a name="MemoryGuard">type</a> [MemoryGuard](https://github.com/cognusion/go-memoryguard/tree/master/athena.go?s=512:1354#L22)
``` go
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
    // contains filtered or unexported fields
}

```
MemoryGuard is our encapsulating mechanation, and should only be acquired via a New helper.
Member functions are goro-safe, but all struct attributes should be set immediatelyish after New(),
and before Limit() is called.



##### Example MemoryGuard:
``` go
// Get a handle on our process
us, _ := os.FindProcess(os.Getpid())

// Create a new MemoryGuard around the process
mg := New(us)
mg.Limit(512 * 1024 * 1024) // Set the HWM memory limit. You can change this at any time

// Do stuff that is memory-hungry

// Stop guarding. After this, if you want to guard the process again,
// Make a New() guard.
mg.Cancel()
// Cancel returns immediately, goros will end eventually.
```





### <a name="New">func</a> [New](https://github.com/cognusion/go-memoryguard/tree/master/athena.go?s=1426:1468#L46)
``` go
func New(Process *os.Process) *MemoryGuard
```
New takes an os.Process and returns a MemoryGuard for that process





### <a name="MemoryGuard.Cancel">func</a> (\*MemoryGuard) [Cancel](https://github.com/cognusion/go-memoryguard/tree/master/athena.go?s=2234:2264#L76)
``` go
func (m *MemoryGuard) Cancel()
```
Cancel signals a Limit() operation to stop, returning immediately.
After calling Cancel this MemoryGuard will be non-functional




### <a name="MemoryGuard.CancelWait">func</a> (\*MemoryGuard) [CancelWait](https://github.com/cognusion/go-memoryguard/tree/master/athena.go?s=2516:2550#L87)
``` go
func (m *MemoryGuard) CancelWait()
```
CancelWait signals a Limit() operation to stop, and waits to return until it is done.
After calling CancelWait this MemoryGuard will be non-functional


##### Example MemoryGuard_CancelWait:
``` go
// Get a handle on our process
us, _ := os.FindProcess(os.Getpid())

// Create a new MemoryGuard around the process
mg := New(us)
mg.Limit(512 * 1024 * 1024) // Set the HWM memory limit. You can change this at any time

// Do stuff that is memory-hungry

// Stop guarding. After this, if you want to guard the process again,
// Make a New() guard.
mg.CancelWait()
// CancelWait pauses until the goros are all done.
```



### <a name="MemoryGuard.Limit">func</a> (\*MemoryGuard) [Limit](https://github.com/cognusion/go-memoryguard/tree/master/athena.go?s=3038:3082#L109)
``` go
func (m *MemoryGuard) Limit(max int64) error
```
Limit takes the max usage (in Bytes) for the process and acts on the PSS.
Returns an error if Limit is called with a zero or negative value,
with a nil Process reference (did you use New()?),
or if it has already been called once before, successfully.




### <a name="MemoryGuard.PSS">func</a> (\*MemoryGuard) [PSS](https://github.com/cognusion/go-memoryguard/tree/master/athena.go?s=1934:1967#L63)
``` go
func (m *MemoryGuard) PSS() int64
```
PSS returns the last known PSS value for the watched process,
or the current value, if there was no last value








- - -
Generated by [godoc2md](http://github.com/cognusion/godoc2md)
