package api

import (
	"github.com/paregi12/torrentserver/engine/web/auth"

	"github.com/gin-gonic/gin"
)

type requestI struct {
	Action string `json:"action,omitempty"`
}

func SetupRoute(route gin.IRouter) {
	authorized := route.Group("/", auth.CheckAuth())

	authorized.GET("/shutdown", shutdown)
	authorized.GET("/shutdown/*reason", shutdown)

	authorized.POST("/settings", settings)

	authorized.POST("/torrents", torrents)

	authorized.POST("/torrent/upload", torrentUpload)

	authorized.POST("/cache", cache)

	authorized.HEAD("/stream", stream)
	authorized.GET("/stream", stream)

	authorized.HEAD("/stream/*fname", stream)
	authorized.GET("/stream/*fname", stream)

	authorized.HEAD("/play/:hash/:id", play)
	authorized.GET("/play/:hash/:id", play)

	authorized.POST("/viewed", viewed)

	authorized.GET("/playlistall/all.m3u", allPlayList)

	authorized.GET("/playlist", playList)
	authorized.GET("/playlist/*fname", playList)

	authorized.GET("/download/:size", download)

	// Add storage settings endpoints
	authorized.GET("/storage/settings", GetStorageSettings)
	authorized.POST("/storage/settings", UpdateStorageSettings)

	// Add TMDB settings endpoint
	authorized.GET("/tmdb/settings", tmdbSettings)

	authorized.GET("/ffp/:hash/:id", ffp)
}
