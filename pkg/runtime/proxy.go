package runtime

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// MountReverseProxy 将 pathPrefix 前缀的请求反代到 targetBase（如 http://aiops-platform:8081）。
func MountReverseProxy(r *gin.Engine, pathPrefix, targetBase string) {
	target, err := url.Parse(strings.TrimRight(targetBase, "/"))
	if err != nil {
		panic("invalid proxy target: " + targetBase)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = target.Host
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = rw.Write([]byte(`{"error":"upstream unavailable"}`))
	}

	handler := func(c *gin.Context) {
		req := c.Request
		if rid := RequestIDFromContext(c); rid != "" {
			req.Header.Set(HeaderRequestID, rid)
		}
		for _, h := range []string{HeaderUserID, HeaderRole, HeaderEnvironment, "Authorization"} {
			if v := c.GetHeader(h); v != "" {
				req.Header.Set(h, v)
			}
		}
		proxy.ServeHTTP(c.Writer, req)
	}

	r.Any(pathPrefix, handler)
	r.Any(pathPrefix+"/*path", handler)
}

// EnvOr 读取环境变量，缺省用 defaultVal。
func EnvOr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
