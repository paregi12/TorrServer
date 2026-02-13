//go:build windows
// +build windows

package fuse

import (
	 "github.com/paregi12/torrentserver/engine/log"
	 "github.com/paregi12/torrentserver/engine/settings"
)

func FuseAutoMount() {
	if settings.Args.FusePath != "" {
		log.TLogln("Windows not support FUSE")
	}
}

func FuseCleanup() {
}
