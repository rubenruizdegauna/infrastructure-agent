// Copyright 2021 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
//go:build linux

package v4

import "path/filepath"

const (
	// defaults for td-agent-bit (<=1.9)
	defaultLoggingBinDir1 = "/opt/td-agent-bit/bin"
	defaultFluentBitExe1  = "td-agent-bit"
	// defaults for fluent-bit (>=2.0)
	defaultLoggingBinDir2 = "/opt/fluent-bit/bin"
	defaultFluentBitExe2  = "fluent-bit"
)

func defaultLoggingBinDir(ffExists bool, ffEnabled bool, _ string) string {
	if !ffExists || !ffEnabled {
		return defaultLoggingBinDir2
	}
	return defaultLoggingBinDir1
}

func defaultFluentBitExePath(ffExists bool, ffEnabled bool, loggingBinDir string) string {
	defaultFluentBitExe := defaultFluentBitExe2
	if ffExists && ffEnabled {
		defaultFluentBitExe = defaultFluentBitExe1
	}

	return filepath.Join(loggingBinDir, defaultFluentBitExe)
}
