// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build dragonfly || freebsd || openbsd
// +build dragonfly freebsd openbsd

package ps

import (
	"bytes"
	"fmt"
	"unsafe"
)

func newUnixProcess(kp *kinfoProc) *unixProcess {
	return &unixProcess{
		pid:          int(kp.Pid),
		ppid:         int(kp.Ppid),
		uid:          int(kp.Uid),
		gid:          int(kp.Groups[0]),
		command:      string(kp.Comm[:bytes.IndexByte(kp.Comm[:], 0)]),
		creationTime: kp.CreationTime(),
	}
}

func processes() ([]Process, error) {
	b, err := sysctlProcAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	n := len(b) / sizeofKinfoProc
	procs := make([]Process, 0, n)
	for i := 0; i < n; i++ {
		kp := (*kinfoProc)(unsafe.Pointer(&b[i*sizeofKinfoProc : (i+1)*sizeofKinfoProc][0]))
		procs = append(procs, newUnixProcess(kp))
	}
	return procs, nil
}

func findProcess(pid int) (Process, error) {
	b, err := sysctlProcPID(pid)
	if err != nil {
		return nil, fmt.Errorf("no process found with PID %d: %w", pid, err)
	}

	if len(b) < sizeofKinfoProc {
		return nil, fmt.Errorf("failed to get process information for PID %d", pid)
	}
	kp := (*kinfoProc)(unsafe.Pointer(&b[:sizeofKinfoProc][0]))
	return newUnixProcess(kp), nil
}
