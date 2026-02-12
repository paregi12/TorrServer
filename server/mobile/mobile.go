package mobile

import (
	"server"
	"server/settings"
	"server/log"
	"strconv"
)

// StartTorrentServer initializes and starts the TorrServer
// Returns the port it's running on
func StartTorrentServer(path string, port int) int {
	settings.Path = path
	log.Init("", "")
	
	if port <= 0 {
		port = 8090
	}
	
	settings.Args = &settings.ExecArgs{
		Port: strconv.Itoa(port),
		Path: path,
	}
	
	go server.Start()
	return port
}

func WaitServer() string {
	return server.WaitServer()
}

func StopServer() {
	server.Stop()
}
