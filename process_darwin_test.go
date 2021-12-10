// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ps_test

import "golang.org/x/sys/unix"

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
