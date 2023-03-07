// Copyright 2021 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//go:build !linux && !windows

package v4

func defaultLoggingBinDir(_ bool, _ bool, _ string) string {
	return ""
}

func defaultFluentBitExePath(_ bool, _ bool, _ string) string {
	return ""
}
