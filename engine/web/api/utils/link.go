package utils

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime"
	     "github.com/paregi12/torrentserver/engine/torrshash"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

func ParseFile(file multipart.File) (*torrent.TorrentSpec, error) {
	minfo, err := metainfo.Load(file)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	// mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	mag := minfo.Magnet(nil, &info)
	spec := new(torrent.TorrentSpec)
	spec.Trackers = [][]string{mag.Trackers}
	spec.DisplayName = info.Name
	spec.InfoHash = minfo.HashInfoBytes()
	return spec, nil
}

func ParseLink(link string) (*torrent.TorrentSpec, error) {
	urlLink, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(urlLink.Scheme) {
	case "magnet":
		return fromMagnet(urlLink.String())
	case "http", "https":
		return fromHttp(urlLink.String())
	case "":
		return fromMagnet("magnet:?xt=urn:btih:" + urlLink.Path)
	case "file":
		return fromFile(urlLink.Path)
	default:
		err = fmt.Errorf("unknown scheme:", urlLink, urlLink.Scheme)
	}
	return nil, err
}

func fromMagnet(link string) (*torrent.TorrentSpec, error) {
	mag, err := metainfo.ParseMagnetUri(link)
	if err != nil {
		return nil, err
	}

	var trackers [][]string
	if len(mag.Trackers) > 0 {
		trackers = [][]string{mag.Trackers}
	}

	spec := new(torrent.TorrentSpec)
	spec.Trackers = trackers
	spec.DisplayName = mag.DisplayName
	spec.InfoHash = mag.InfoHash
	return spec, nil
}

func ParseTorrsHash(token string) (*torrent.TorrentSpec, *torrshash.TorrsHash, error) {
	if strings.HasPrefix(token, "torrs://") {
		token = strings.TrimPrefix(token, "torrs://")
	}
	th, err := torrshash.Unpack(token)
	if err != nil {
		return nil, nil, err
	}

	var trackers [][]string
	if len(th.Trackers()) > 0 {
		trackers = [][]string{th.Trackers()}
	}

	spec := new(torrent.TorrentSpec)
	spec.Trackers = trackers
	spec.DisplayName = th.Title()
	spec.InfoHash = metainfo.NewHashFromHex(th.Hash)
	return spec, th, nil
}

func fromHttp(link string) (*torrent.TorrentSpec, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	client := new(http.Client)
	client.Timeout = time.Duration(time.Second * 60)
	req.Header.Set("User-Agent", "DWL/1.1.1 (Torrent)")

	resp, err := client.Do(req)
	if er, ok := err.(*url.Error); ok {
		if strings.HasPrefix(er.URL, "magnet:") {
			return fromMagnet(er.URL)
		}
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	minfo, err := metainfo.Load(resp.Body)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}
	// mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	mag := minfo.Magnet(nil, &info)

	spec := new(torrent.TorrentSpec)
	spec.Trackers = [][]string{mag.Trackers}
	spec.DisplayName = info.Name
	spec.InfoHash = minfo.HashInfoBytes()
	return spec, nil
}

func fromFile(path string) (*torrent.TorrentSpec, error) {
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	minfo, err := metainfo.LoadFromFile(path)
	if err != nil {
		return nil, err
	}
	info, err := minfo.UnmarshalInfo()
	if err != nil {
		return nil, err
	}

	// mag := minfo.Magnet(info.Name, minfo.HashInfoBytes())
	mag := minfo.Magnet(nil, &info)
	spec := new(torrent.TorrentSpec)
	spec.Trackers = [][]string{mag.Trackers}
	spec.DisplayName = info.Name
	spec.InfoHash = minfo.HashInfoBytes()
	return spec, nil
}
