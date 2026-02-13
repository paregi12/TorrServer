package nuvioServer

import (
	"github.com/paregi12/torrentserver/server"
	"github.com/paregi12/torrentserver/server/settings"
	"github.com/paregi12/torrentserver/server/torr/utils"
	"strconv"
	"strings"
)

func StartTorrentServer(pathdb string, port int64) int64 {
	settings.Args = &settings.ExecArgs{
		Path: pathdb,
		Port: strconv.FormatInt(port, 10),
	}
	return int64(server.Start())
}

func WaitTorrentServer() {
	server.WaitServer()
}

func StopTorrentServer() {
	server.Stop()
}

func AddTrackers(trackers string) {
	tracks := strings.Split(trackers, ",\n")
	utils.SetDefTrackers(tracks)
}
