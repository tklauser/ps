// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ps

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

func getDarwinVersion() int {
	osrel, err := unix.Sysctl("kern.osrelease")
	if err != nil {
		return 0
	}
	ver := 0
	for i := 0; i < len(osrel) && '0' <= osrel[i] && osrel[i] <= '9'; i++ {
		ver *= 10
		ver += int(osrel[i] - '0')
	}
	return ver
}

func getExePathAndArgs(pid int) (string, []string) {
	// TODO(tk): figure out why the format returned by the kern.procargs2 sysctl is different on
	// macOS 10.15 and earlier.
	if getDarwinVersion() <= 19 {
		// 19.x.y is macOS 10.15 (Catalina), see
		// https://en.wikipedia.org/wiki/Darwin_(operating_system)#Release_history
		return "", nil
	}

	// See function getproclline() in adv_cmds/ps/print.c
	// The format of KERN_PROCARGS2 is a C int (argc) followed by the executable’s string area.
	// The string area consists of NUL-terminated strings, beginning with the executable path,
	// and then starting on an aligned boundary, all of the elements of argv, envp, and applev.
	procArgs, err := unix.SysctlRaw("kern.procargs2", pid)
	if err != nil {
		return "", nil
	}
	return parseProcArgs(procArgs)
}

func parseProcArgs(procArgs []byte) (string, []string) {
	var argc int32 // C.int
	if err := binary.Read(bytes.NewReader(procArgs), binary.LittleEndian, &argc); err != nil {
		return "", nil
	}
	if argc < 1 {
		return "", nil
	}

	procArgs = procArgs[4:]
	nulPos := bytes.IndexByte(procArgs, 0)
	exe := string(procArgs[:nulPos])
	if argc == 1 {
		return exe, []string{exe}
	}

	nulPos++
	for nulPos < len(procArgs) && procArgs[nulPos] == 0 {
		nulPos++
	}

	procArgs = procArgs[nulPos:]
	args := make([]string, 0, argc)
	for i := 0; i < int(argc) && len(procArgs) > 0; i++ {
		arg := string(bytes.Trim(procArgs[:bytes.IndexByte(procArgs, 0)], "\x00"))
		args = append(args, arg)
		procArgs = procArgs[len(arg)+1:]
	}
	return exe, args
}

func newUnixProcess(kp *unix.KinfoProc) *unixProcess {
	pid := int(kp.Proc.P_pid)
	exePath, exeArgs := getExePathAndArgs(pid)
	return &unixProcess{
		pid:            pid,
		ppid:           int(kp.Eproc.Ppid),
		uid:            int(kp.Eproc.Ucred.Uid),
		gid:            int(kp.Eproc.Ucred.Groups[0]),
		command:        string(kp.Proc.P_comm[:bytes.IndexByte(kp.Proc.P_comm[:], 0)]),
		executablePath: exePath,
		executableArgs: exeArgs,
		creationTime:   time.Unix(kp.Proc.P_starttime.Sec, int64(kp.Proc.P_starttime.Usec)*1000),
	}
}

func processes() ([]Process, error) {
	kinfoProcs, err := unix.SysctlKinfoProcSlice("kern.proc.all")
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	procs := make([]Process, 0, len(kinfoProcs))
	for _, kp := range kinfoProcs {
		procs = append(procs, newUnixProcess(&kp))
	}
	return procs, nil
}

func findProcess(pid int) (Process, error) {
	kp, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return nil, fmt.Errorf("no process found with PID %d: %w", pid, err)
	}

	if kpid := int(kp.Proc.P_pid); kpid != pid {
		return nil, fmt.Errorf("kernel info PID %d doesn't match requested PID %d", kpid, pid)
	}
	return newUnixProcess(kp), nil
}
