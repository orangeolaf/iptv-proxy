package routes

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/jamesnetherton/m3u"

	"github.com/pierre-emmanuelJ/iptv-proxy/pkg/config"
	proxyM3U "github.com/pierre-emmanuelJ/iptv-proxy/pkg/m3u"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type proxy struct {
	*config.ProxyConfig
	*m3u.Track
}

// Serve the pfinder api
func Serve(proxyConfig *config.ProxyConfig) error {
	router := gin.Default()
	router.Use(cors.Default())
	Routes(proxyConfig, router.Group("/"))

	return router.Run(fmt.Sprintf(":%d", proxyConfig.HostConfig.Port))
}

// Routes adds the routes for the app to the RouterGroup r
func Routes(proxyConfig *config.ProxyConfig, r *gin.RouterGroup) {

	p := &proxy{
		proxyConfig,
		nil,
	}

	r.GET("/iptv.m3u", p.getM3U)

	for i, track := range proxyConfig.Playlist.Tracks {
		oriURL, err := url.Parse(track.URI)
		if err != nil {
			return
		}
		tmp := &proxy{
			nil,
			&proxyConfig.Playlist.Tracks[i],
		}
		r.GET(oriURL.RequestURI(), tmp.reverseProxy)
	}
}

func (p *proxy) reverseProxy(c *gin.Context) {
	rpURL, err := url.Parse(p.Track.URI)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get(rpURL.String())
	if err != nil {
		log.Fatal(err)
	}

	c.Status(resp.StatusCode)
	c.Stream(func(w io.Writer) bool {
		io.Copy(w, resp.Body)
		return false
	})
}

func (p *proxy) getM3U(c *gin.Context) {
	result, err := proxyM3U.Marshall(p.Playlist, p.HostConfig)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "text/plain", []byte(result))
}