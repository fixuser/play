package meta

import (
	"cmp"
	"context"
	"net"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	HeaderUserAgent     = "x-meta-user-agent"
	HeaderRequestMethod = "x-meta-request-method"
	HeaderRequestPath   = "x-meta-request-path"
	HeaderRequestHost   = "x-meta-request-host"
	HeaderToken         = "x-meta-token"
	MetaUserId          = "x-meta-user-id"
	MetaUserType        = "x-meta-user-type"
	MetaUser            = "x-meta-user"
	MetaUserIp          = "x-meta-user-ip"
	MetaUserCountry     = "x-meta-user-country"
	MetaDeviceId        = "x-meta-device-id"
	MetaLanguage        = "x-meta-language"
	MetaVersion         = "x-meta-version"
	MetaPlatform        = "x-meta-platform"
)

// MetadataAnnotator 是一个中间件，用于从 HTTP 请求中提取元数据并将其添加到 gRPC 元数据中
func MetadataAnnotator(ctx context.Context, req *http.Request) (md metadata.MD) {
	data := make(map[string]string, 15)

	// 从请求头中提取元数据
	data[HeaderUserAgent] = req.Header.Get(HeaderUserAgent)
	data[HeaderRequestMethod] = req.Method
	data[HeaderRequestPath] = req.URL.Path
	data[HeaderRequestHost] = req.Host

	data[MetaLanguage] = cmp.Or(req.Header.Get("Accept-Language"), req.Header.Get(MetaLanguage))
	data[MetaVersion] = req.Header.Get(MetaVersion)
	data[MetaPlatform] = req.Header.Get(MetaPlatform)
	data[MetaDeviceId] = req.Header.Get(MetaDeviceId)

	// 获取用户登陆令牌
	token := req.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		token = req.Header.Get(HeaderToken)
	}
	data[HeaderToken] = token

	// 获取用户IP
	ip := cmp.Or(req.Header.Get("Cf-Connecting-Ip"), req.Header.Get("CloudFront-Viewer-Address"), req.Header.Get("X-Real-Ip"))
	data[MetaUserIp], _, _ = net.SplitHostPort(ip)

	// 获取用户国家
	country := cmp.Or(req.Header.Get("Cf-Ipcountry"), req.Header.Get("CloudFront-Viewer-Country"))
	data[MetaUserCountry] = strings.ToUpper(country)

	// 删除value为空的key
	for k, v := range data {
		if v == "" {
			delete(data, k)
		}
	}
	return metadata.New(data)
}
