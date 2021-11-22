// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build dragonfly || freebsd
// +build dragonfly freebsd

package ps

import (
	"time"

	"golang.org/x/sys/unix"
)

func (kp *kinfoProc) CreationTime() time.Time {
	return time.Unix(kp.Start.Sec, int64(kp.Start.Usec)*1000)
}

func sysctlProcAll() ([]byte, error) {
	return unix.SysctlRaw("kern.proc.all")
}

func sysctlProcPID(pid int) ([]byte, error) {
	return unix.SysctlRaw("kern.proc.pid", pid)
}
