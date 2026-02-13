package server

import (
	"net"
	"os"
	"path/filepath"
	"strconv"

	 "github.com/paregi12/torrentserver/engine/tgbot"

	 "github.com/paregi12/torrentserver/engine/log"
	 "github.com/paregi12/torrentserver/engine/settings"
	 "github.com/paregi12/torrentserver/engine/web"
)

func Start() int {
	if !settings.InitSets(settings.Args.RDB, settings.Args.SearchWA) {
		return -1
	}
	// http checks
	if settings.Args.Port == "" {
		settings.Args.Port = "8090"
	}

	log.TLogln("Check web port", settings.Args.Port)
	l, err := net.Listen("tcp", settings.Args.IP+":"+settings.Args.Port)
	if l != nil {
		l.Close()
	}
	if err != nil {
		log.TLogln("Port", settings.Args.Port, "already in use! Please set different port for HTTP. Abort")
		return -1
	}
	realPort := l.Addr().(*net.TCPAddr).Port
	settings.Args.Port = strconv.Itoa(realPort)

	// remove old disk caches
	go cleanCache()
	// set settings http and https ports. Start web server.
	settings.Port = settings.Args.Port
	settings.SslPort = settings.Args.SslPort
	settings.IP = settings.Args.IP

	if settings.Args.TGToken != "" {
		tgbot.Start(settings.Args.TGToken)
	}
	if web.Start() {
		return realPort
	} else {
		return -1
	}
}

func cleanCache() {
	if !settings.BTsets.UseDisk || settings.BTsets.TorrentsSavePath == "/" || settings.BTsets.TorrentsSavePath == "" {
		return
	}

	dirs, err := os.ReadDir(settings.BTsets.TorrentsSavePath)
	if err != nil {
		return
	}

	torrs := settings.ListTorrent()

	log.TLogln("Remove unused cache in dir:", settings.BTsets.TorrentsSavePath)
	keep := map[string]bool{}
	for _, d := range dirs {
		if len(d.Name()) != 40 {
			// Not a hash
			continue
		}

		if !settings.BTsets.RemoveCacheOnDrop {
			keep[d.Name()] = true
			for _, t := range torrs {
				if d.IsDir() && d.Name() == t.InfoHash.HexString() {
					keep[d.Name()] = false
					break
				}
			}
			for hash, del := range keep {
				if del && hash == d.Name() {
					log.TLogln("Remove unused cache:", d.Name())
					removeAllFiles(filepath.Join(settings.BTsets.TorrentsSavePath, d.Name()))
				}
			}
		} else {
			if d.IsDir() {
				log.TLogln("Remove unused cache:", d.Name())
				removeAllFiles(filepath.Join(settings.BTsets.TorrentsSavePath, d.Name()))
			}
		}
	}
}

func removeAllFiles(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, f := range files {
		name := filepath.Join(path, f.Name())
		os.Remove(name)
	}
	os.Remove(path)
}

func WaitServer() string {
	err := web.Wait()
	if err != nil {
		return err.Error()
	}
	return ""
}

func Stop() {
	web.Stop()
	settings.CloseDB()
}
