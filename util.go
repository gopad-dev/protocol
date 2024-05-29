// SPDX-FileCopyrightText: 2019 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"log/slog"
)

// NewVersion returns the int32 pointer converted i.
func NewVersion(i int32) *int32 {
	return &i
}

func Error(err error) slog.Attr {
	return slog.Any("err", err)
}
