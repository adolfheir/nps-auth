package server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	netHttp "net/http"
	"net/http/httputil"
	"net/url"
	"nps-auth/pkg/sql"
	"strings"

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
		target := fmt.Sprintf("http://127.0.0.1:%d", port)

		targetURL, err := url.Parse(target)
		if err != nil {
			log.Error().Err(err).Str("target", target).Msg("query channel error")
			c.JSON(netHttp.StatusNotFound, gin.H{"error": "未知路径"})
			return
		}

		// 打印原地址和新地址
		originalURL := c.Request.URL.String()
		newURL := fmt.Sprintf("%s%s", target, pathParts)
		log.Info().Str("original_url", originalURL).Str("new_url", newURL).Msg("dynamicReverseProxy")

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// 修改请求地址
		c.Request.Host = targetURL.Host
		c.Request.URL.Path = pathParts

		// 修改html内容
		proxy.ModifyResponse = func(resp *http.Response) error {
			if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
				var bodyBytes []byte
				var err error

				if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
					// 解压缩 Gzip 响应体
					gr, err := gzip.NewReader(resp.Body)
					if err != nil {
						return err
					}
					defer gr.Close()

					bodyBytes, err = io.ReadAll(gr)
					if err != nil {
						return err
					}
				} else {
					// 读取未压缩的响应体
					bodyBytes, err = io.ReadAll(resp.Body)
					if err != nil {
						return err
					}
				}

				bodyStr := string(bodyBytes)

				// 替换 window.__dynamic_base__ 的值
				newBaseUrl := fmt.Sprintf("window.__dynamic_base__ = \"/proxy/%s/\"", channelId)
				modifiedBodyStr := strings.ReplaceAll(bodyStr, "window.__dynamic_base__ = \"/\"", newBaseUrl)
				modifiedBody := []byte(modifiedBodyStr)

				// 重新压缩修改后的响应体
				var buf bytes.Buffer
				w := gzip.NewWriter(&buf)
				if _, err := w.Write(modifiedBody); err != nil {
					return err
				}
				if err := w.Close(); err != nil {
					return err
				}

				// 更新响应体
				resp.Body = io.NopCloser(&buf)
				resp.ContentLength = int64(buf.Len())
				resp.Header.Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
				resp.Header.Set("Content-Encoding", "gzip") // 设置压缩标识
			}
			return nil
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
