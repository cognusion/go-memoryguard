package memoryguard

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	. "github.com/smartystreets/goconvey/convey"
)

func ExampleMemoryGuard() {
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
}

func ExampleMemoryGuard_CancelWait() {
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
}

func Test_MemoryGuardOnUsPSSRapid(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.Limit(400 * 1024 * 1024) // we won't actually hit this, right?
		defer mg.Cancel()

		Convey("and we spam the PSS() function, we don't get killed, and a PSS is returned", func() {
			for range 1000 {
				So(mg.PSS(), ShouldBeGreaterThan, 0)
			}
			So(mg.running.Load(), ShouldBeTrue)
		})
	})
}

func Test_MemoryGuardOnUsPSS(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.Name = "bob"
		mg.Limit(400 * 1024 * 1024) // we won't actually hit this, right?
		defer mg.Cancel()

		Convey("we don't get killed, and a PSS is returned", func() {
			So(mg.running.Load(), ShouldBeTrue)
			So(mg.PSS(), ShouldBeGreaterThan, 0)
		})

		Convey("if we call Limit() again it refuses", func() {
			So(mg.Limit(400*1024*1024), ShouldBeError)
		})
	})
}

func Test_MemoryGuardLimitZero(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is created with a Limit of 0, it refuses", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.Name = "bob"
		So(mg.Limit(0), ShouldBeError)
	})
}

func Test_MemoryGuardOnUsDelay(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		var limit = int64(400 * 1024 * 1024)

		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.StatsFrequency = time.Second
		mg.Limit(limit) // we won't actually hit this, right?
		// deliberately NOT defering a cancel, because we are going to CancelWait later.
		//defer mg.Cancel()

		Convey("After 2 seconds", func() {
			time.Sleep(2 * time.Second)

			Convey("we don't get killed, and a PSS is returned", func() {
				So(mg.running.Load(), ShouldBeTrue)
				So(mg.PSS(), ShouldBeGreaterThan, 0)
			})

			Convey("and we can span the Limit function, and the goro count is stable. (HINT: LeakTest will bomb too, if this failed)", func() {
				count := runtime.NumGoroutine()
				for range 1000 {
					mg.Limit(limit)
				}
				So(runtime.NumGoroutine(), ShouldEqual, count)
			})

			mg.CancelWait()
			So(mg.running.Load(), ShouldBeFalse)
		})

	})
}

func Test_MemoryGuardSmapsBadPid(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard checks SMAPS for an invalid pid", t, func() {
		_, e := getPss(-10)
		Convey("it returns an error", func() {
			So(e, ShouldNotBeNil)
		})
	})
}

func Test_MemoryGuardGetPss(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard checks SMAPS with a valid pid for PSS", t, func() {
		_, e := getPss(os.Getpid())
		Convey("it doesn't return an error", func() {
			So(e, ShouldBeNil)
		})
	})
}

func Test_MemoryGuardGetPssBadPid(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.proc.Pid = -10
		mg.Limit(400 * 1024 * 1024) // we won't actually hit this, right?
		defer mg.Cancel()

		// Pause so the limiter can run a cycle or two on the bad PID, possibly
		// dying.
		time.Sleep(2 * time.Second)

		Convey("and we have a bad pid, we don't get killed, and a PSS of 0 is returned", func() {
			So(mg.running.Load(), ShouldBeTrue)
			So(mg.PSS(), ShouldEqual, 0)
		})
	})
}

func Test_MemoryGuardCancelSpam(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.Limit(400 * 1024 * 1024) // we won't actually hit this, right?
		defer mg.Cancel()

		Convey("and we spam the cancel function, we don't get blocked", func() {
			for range 1000 {
				mg.Cancel()
			}
			mg.CancelWait() // for latency
			So(mg.running.Load(), ShouldBeFalse)
		})
	})
}

func Test_MemoryGuardCancelWaitSpam(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.Limit(400 * 1024 * 1024) // we won't actually hit this, right?
		defer mg.Cancel()

		Convey("and we spam the cancel function, we don't get blocked", func() {
			for range 1000 {
				mg.CancelWait()
			}
			So(mg.running.Load(), ShouldBeFalse)
		})
	})
}

func Test_MemoryGuardKillPSS(t *testing.T) {
	defer leaktest.Check(t)()

	Convey("When a MemoryGuard is running on us", t, func() {
		us, _ := os.FindProcess(os.Getpid())
		mg := New(us)
		mg.Interval = time.Second
		mg.nokill = true

		Convey("and set a really low threshold, we'll get killed", func() {
			defer mg.Cancel()
			mg.Limit(1024) // 1KB

			<-mg.KillChan // wait for the kill
			So(mg.running.Load(), ShouldBeFalse)
		})
	})
}

func Test_MemoryGuardMaxPSS(t *testing.T) {
	// t.Skip("Skipping, as this runs and external command without consent, that could chew up memory if MG doesn't work.\n")
	defer leaktest.Check(t)()

	Convey("When an external command runs", t, func() {
		cmd := exec.Command("tests/mem.sh")
		err := cmd.Start()
		So(err, ShouldBeNil)
		mg := New(cmd.Process)
		mg.Interval = time.Second
		mg.Limit(1024 * 1024) // 1MB

		Convey("and memory grows above mss, it should be killed promptly.", func() {
			defer mg.Cancel()
			start := time.Now()
			err := cmd.Wait()
			<-mg.KillChan // wait for the kill
			stop := time.Now()
			So(err, ShouldNotBeNil)
			So(stop.Sub(start), ShouldBeLessThanOrEqualTo, 3*time.Second)
			So(mg.running.Load(), ShouldBeFalse)
		})

	})
}
