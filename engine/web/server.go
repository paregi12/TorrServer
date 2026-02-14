package web

import (
	"context"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"

	"github.com/paregi12/torrentserver/engine/dlna"
	"github.com/paregi12/torrentserver/engine/log"
	"github.com/paregi12/torrentserver/engine/settings"
	"github.com/paregi12/torrentserver/engine/torr"
	"github.com/paregi12/torrentserver/engine/version"
	"github.com/paregi12/torrentserver/engine/web/api"
	"github.com/paregi12/torrentserver/engine/web/auth"
	"github.com/paregi12/torrentserver/engine/web/blocker"
	"github.com/paregi12/torrentserver/engine/web/msx"
	"github.com/paregi12/torrentserver/engine/web/pages"
	"github.com/paregi12/torrentserver/engine/web/sslcerts"

	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

var (
	BTS      = torr.NewBTS()
	waitChan = make(chan error, 4)

	serverMu    sync.Mutex
	httpServer  *http.Server
	httpsServer *http.Server
)

func reportWait(err error) {
	select {
	case waitChan <- err:
	default:
	}
}

//	@title			Swagger Torrserver API
//	@version		{version.Version}
//	@description	Torrent streaming server.

//	@license.name	GPL 3.0

//	@BasePath	/

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func Start() bool {
	log.TLogln("Start TorrServer " + version.Version + " torrent " + version.GetTorrentVersion())
	ips := GetLocalIps()
	if len(ips) > 0 {
		log.TLogln("Local IPs:", ips)
	}
	err := BTS.Connect()
	if err != nil {
		log.TLogln("BTS.Connect() error!", err)
		return false
	}

	gin.SetMode(gin.ReleaseMode)

	corsCfg := cors.DefaultConfig()
	allowAllCORS := os.Getenv("TS_ALLOW_ALL_CORS") == "1"
	if allowAllCORS {
		corsCfg.AllowAllOrigins = true
		corsCfg.AllowPrivateNetwork = true
	} else {
		// Default to local origins only; set TS_ALLOW_ALL_CORS=1 to restore permissive behavior.
		corsCfg.AllowOrigins = []string{
			"http://127.0.0.1",
			"http://localhost",
		}
	}
	corsCfg.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "X-Requested-With", "Accept", "Authorization"}

	route := gin.New()
	route.Use(log.WebLogger(), blocker.Blocker(), gin.Recovery(), cors.New(corsCfg), location.Default())
	auth.SetupAuth(route)

	route.GET("/echo", echo)

	api.SetupRoute(route)
	msx.SetupRoute(route)
	pages.SetupRoute(route)

	if settings.BTsets.EnableDLNA {
		dlna.Start()
	}

	route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	serverMu.Lock()
	defer serverMu.Unlock()

	// check if https enabled
	if settings.Ssl {
		if settings.BTsets.SslCert == "" || settings.BTsets.SslKey == "" {
			settings.BTsets.SslCert, settings.BTsets.SslKey = sslcerts.MakeCertKeyFiles(ips)
			log.TLogln("Saving path to ssl cert and key in db", settings.BTsets.SslCert, settings.BTsets.SslKey)
			settings.SetBTSets(settings.BTsets)
		}
		err = sslcerts.VerifyCertKeyFiles(settings.BTsets.SslCert, settings.BTsets.SslKey, settings.SslPort)
		if err != nil {
			log.TLogln("Error checking certificate and private key files:", err)
			settings.BTsets.SslCert, settings.BTsets.SslKey = sslcerts.MakeCertKeyFiles(ips)
			log.TLogln("Saving path to ssl cert and key in db", settings.BTsets.SslCert, settings.BTsets.SslKey)
			settings.SetBTSets(settings.BTsets)
		}

		httpsServer = &http.Server{
			Addr:    settings.IP + ":" + settings.SslPort,
			Handler: route,
		}
		go func(server *http.Server) {
			log.TLogln("Start https server at", server.Addr)
			if runErr := server.ListenAndServeTLS(settings.BTsets.SslCert, settings.BTsets.SslKey); runErr != nil && runErr != http.ErrServerClosed {
				reportWait(runErr)
			}
		}(httpsServer)
	}

	httpServer = &http.Server{
		Addr:    settings.IP + ":" + settings.Port,
		Handler: route,
	}
	go func(server *http.Server) {
		log.TLogln("Start http server at", server.Addr)
		if runErr := server.ListenAndServe(); runErr != nil && runErr != http.ErrServerClosed {
			reportWait(runErr)
		}
	}(httpServer)

	return true
}

func Wait() error {
	return <-waitChan
}

func Stop() {
	dlna.Stop()

	serverMu.Lock()
	httpSrv := httpServer
	httpsSrv := httpsServer
	httpServer = nil
	httpsServer = nil
	serverMu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if httpsSrv != nil {
		if err := httpsSrv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			log.TLogln("HTTPS shutdown error:", err)
		}
	}

	if httpSrv != nil {
		if err := httpSrv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			log.TLogln("HTTP shutdown error:", err)
		}
	}

	BTS.Disconnect()
	reportWait(nil)
}

// echo godoc
//
//	@Summary		Tests server status
//	@Description	Tests whether server is alive or not
//
//	@Tags			API
//
//	@Produce		plain
//	@Success		200	{string}	string	"Server version"
//	@Router			/echo [get]
func echo(c *gin.Context) {
	c.String(200, "%v", version.Version)
}

func GetLocalIps() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.TLogln("Error get local IPs")
		return nil
	}
	var list []string
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		if i.Flags&net.FlagUp == net.FlagUp {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() {
					list = append(list, ip.String())
				}
			}
		}
	}
	sort.Strings(list)
	return list
}
