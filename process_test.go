// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ps_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tklauser/ps"
)

var startTime = time.Now()

func getExeName() string {
	if runtime.GOOS == "windows" {
		return "ps.test.exe"
	}
	return "ps.test"
}

func logProcess(t *testing.T, p ps.Process) {
	t.Helper()

	t.Logf("process %d (%s) ppid %d uid %d gid %d ctime %v",
		p.PID(), p.Command(), p.PPID(), p.UID(), p.GID(), p.CreationTime())

	exeArgs := ""
	if args := p.ExecutableArgs(); len(args) > 1 {
		exeArgs = " " + strings.Join(args[1:], " ")
	}
	t.Logf("  $ %s%s", p.ExecutablePath(), exeArgs)
}

func checkOwnProcess(t *testing.T, p ps.Process) {
	t.Helper()

	logProcess(t, p)

	if got, want := p.PID(), os.Getpid(); got != want {
		t.Errorf("PID: got %v, want %v", got, want)
	}
	if got, want := p.PPID(), os.Getppid(); got != want {
		t.Errorf("PPID: got %v, want %v", got, want)
	}
	if got, want := p.UID(), os.Getuid(); got != want {
		t.Errorf("UID: got %v, want %v", got, want)
	}
	if got, want := p.GID(), os.Getgid(); got != want {
		t.Errorf("GID: got %v, want %v", got, want)
	}
	if got, want := p.Command(), getExeName(); got != want {
		t.Errorf("Command: got %v, want %v", got, want)
	}
	if got, want := filepath.Base(p.ExecutablePath()), getExeName(); got != want {
		t.Errorf("ExecutablePath: got command %q, want %q", got, want)
	}

	slack := 2 * time.Minute
	if diff := p.CreationTime().Sub(startTime); diff > slack {
		t.Errorf("process created %v after tests started", diff)
	} else if diff < -slack {
		t.Errorf("process created %v before tests started", -diff)
	}
}

func TestProcesses(t *testing.T) {
	procs, err := ps.Processes()
	if err != nil {
		t.Fatalf("Processes: %v", err)
	}
	if len(procs) == 0 {
		t.Errorf("no processes returned")
	}

	for _, p := range procs {
		if p.Command() == getExeName() {
			checkOwnProcess(t, p)
			return
		}
	}
	t.Errorf("didn't find process with command name %q", getExeName())
}

func getInitName() string {
	switch runtime.GOOS {
	case "darwin":
		return "launchd"
	case "dragonfly", "freebsd", "netbsd", "openbsd",
		"illumos", "solaris":
		return "init"
	case "linux":
		// might be systemd, sysv init, openrc, ...
		b, err := ioutil.ReadFile("/proc/1/comm")
		if err == nil {
			return string(bytes.Trim(b, " \r\n"))
		}
	case "windows":
		return "System"
	}
	return ""
}

func TestFindProcessOwn(t *testing.T) {
	proc, err := ps.FindProcess(-1)
	if err == nil {
		t.Fatal("FindProcess: got nil error, want an error")
	}
	if proc != nil {
		t.Fatalf("FindProcess: got process %v, want nil", proc)
	}
	proc, err = ps.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if proc == nil {
		t.Fatal("FindProcess: got nil process")
	}

	checkOwnProcess(t, proc)
}

func TestFindProcessInit(t *testing.T) {
	initPID := 1
	wantUID, wantGID := 0, 0
	if runtime.GOOS == "windows" {
		// see https://devblogs.microsoft.com/oldnewthing/?p=23283
		initPID = 4
		wantUID, wantGID = -1, -1
	}

	proc, err := ps.FindProcess(initPID)
	if os.IsPermission(err) {
		t.Skipf("no permission to read init process")
	} else if err != nil {
		t.Fatalf("FindProcess(%d): %v", initPID, err)
	}

	logProcess(t, proc)

	if got, want := proc.PID(), initPID; got != want {
		t.Errorf("PID: got %v, want %v", got, want)
	}
	if got, want := proc.PPID(), 0; got != want {
		t.Errorf("Parent PID: got %v, want %v", got, want)
	}
	if uid := proc.UID(); uid != wantUID {
		t.Errorf("UID: got %v, want %v", uid, wantUID)
	}
	if gid := proc.GID(); gid != wantGID {
		t.Errorf("GID: got %v, want %v", gid, wantGID)
	}
	if initName := getInitName(); initName != "" {
		cmd := proc.Command()
		if (runtime.GOOS == "illumos" || runtime.GOOS == "solaris") && cmd == "" {
			t.Skipf("command: empty; might lack permissions to read /proc/1 on %s, skipping", runtime.GOOS)
		} else if cmd != initName {
			t.Errorf("command: got %q, want %q", cmd, initName)
		}
	} else {
		t.Skipf("init process name not defined on %s", runtime.GOOS)
	}
}

func BenchmarkProcesses(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := ps.Processes()
		if err != nil {
			b.Fatalf("Processes: %v", err)
		}
	}
}

func BenchmarkFindProcess(b *testing.B) {
	pid := os.Getpid()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ps.FindProcess(pid)
		if err != nil {
			b.Fatalf("FindProcess: %v", err)
		}
	}
}
