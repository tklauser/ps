// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/tklauser/ps"
)

func main() {
	procs, err := ps.Processes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list processes: %v\n", err)
		os.Exit(-1)
	}

	sort.Slice(procs, func(i, j int) bool {
		return procs[i].PID() < procs[j].PID()
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "PID\tPPID\tUID\tCOMMAND\n")
	for _, p := range procs {
		exeArgs := ""
		if args := p.ExecutableArgs(); len(args) > 1 {
			exeArgs = " " + strings.Join(args[1:], " ")
		}
		fmt.Fprintf(w, "%d\t%d\t%d\t%s%s\n",
			p.PID(),
			p.PPID(),
			p.UID(),
			p.ExecutablePath(), exeArgs)
	}
	w.Flush()
}
