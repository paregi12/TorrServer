package nuvioServer

import (
	"github.com/paregi12/torrentserver/engine"
	"github.com/paregi12/torrentserver/engine/settings"
	"github.com/paregi12/torrentserver/engine/torr/utils"
	"strconv"
	"strings"
)

func StartTorrentServer(pathdb string, port int64) int64 {
	settings.Path = pathdb
	settings.Args = &settings.ExecArgs{
		Path: pathdb,
		Port: strconv.FormatInt(port, 10),
	}
	return int64(engine.Start())
}

func WaitTorrentServer() {
	engine.WaitServer()
}

func StopTorrentServer() {
	engine.Stop()
}

func AddTrackers(trackers string) {
	tracks := strings.Split(trackers, ",\n")
	utils.SetDefTrackers(tracks)
}
