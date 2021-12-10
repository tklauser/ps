// Copyright 2021 Tobias Klauser. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !darwin
// +build !darwin

package ps_test

func getDarwinVersion() int {
	return 0
}
