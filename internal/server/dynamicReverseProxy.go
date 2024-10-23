package server

import (
	"fmt"
	netHttp "net/http"
	"net/http/httputil"
	"net/url"
	"nps-auth/pkg/sql"

	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru/v2"
)

var (
	lruCache *lru.Cache[string, int]
)

func initLru() {
	var err error
	lruCache, err = lru.New[string, int](2000)
	if err != nil {
		log.Panic().Err(err).Msg("init lru error")
	}

}

// 创建一个反向代理
func dynamicReverseProxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取路径中的动态部分，即 xxx
		channelId := c.Param("channel")    // 获取 /proxy/xxx 部分
		pathParts := c.Param("proxyParts") // 获取 /index.html 部分

		// 查找对应的端口号
		var port int
		var ok bool
		port, ok = lruCache.Get(channelId)
		if !ok {
			var channel sql.Channel
			if err := sql.GetDB().First(&channel, "channel_id = ?", channelId).Error; err != nil {
				log.Error().Err(err).Msg("query channel error")
				c.JSON(netHttp.StatusNotFound, gin.H{"error": "未知路径"})
				return
			} else {
				port = channel.NpsTunnelPort
				lruCache.Add(channelId, port)
			}
		}

		// 生成代理目标URL
		target := fmt.Sprintf("http://127.0.0.1:%d/%s", port, pathParts)
		targetURL, err := url.Parse(target)
		if err != nil {
			log.Error().Err(err).Str("target", target).Msg("query channel error")
			c.JSON(netHttp.StatusNotFound, gin.H{"error": "未知路径"})
			return
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// 修改请求的Host为目标地址
		c.Request.Host = targetURL.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
