package compile

import (
	"fmt"
	"net"
	"net/url"

	"github.com/goccy/go-json"
)

// URL 支持url的序列化
type URL url.URL

func (u URL) String() string {
	v := (*url.URL)(&u)
	return v.String()
}

func (u *URL) UnmarshalText(data []byte) error {
	n, err := url.Parse(string(data))
	if err != nil {
		return err
	}
	*u = (URL)(*n)
	return nil
}

func (u URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

type NodeInfo struct {
	Id        string            `json:"id"` // 唯一性
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []URL             `json:"endpoints"`
}

func NewNodeInfo() *NodeInfo {
	return &NodeInfo{
		Metadata: make(map[string]string),
	}
}

// AddVersion 添加版本
func (n *NodeInfo) AddVersion(version string) *NodeInfo {
	n.Version = version
	return n
}

// AddMetadata 添加元数据
func (n *NodeInfo) AddMetadata(key, value string) *NodeInfo {
	n.Metadata[key] = value
	return n
}

// AddHttPort 添加http端点
func (n *NodeInfo) AddHttpPort(port string) *NodeInfo {
	var nurl url.URL
	nurl.Scheme = "http"
	nurl.Host = net.JoinHostPort(IpAddr, port)
	n.Endpoints = append(n.Endpoints, (URL)(nurl))
	return n
}

// AddGrpcPort 添加grpc端点
func (n *NodeInfo) AddGrpcPort(port string) *NodeInfo {
	var nurl url.URL
	nurl.Scheme = "grpc"
	nurl.Host = net.JoinHostPort(IpAddr, port)
	n.Endpoints = append(n.Endpoints, (URL)(nurl))
	return n
}

// AddName 添加名称
func (n *NodeInfo) AddName(name string) *NodeInfo {
	n.Name = name
	return n
}

// AddId 添加id
func (n *NodeInfo) AddId(id string) *NodeInfo {
	n.Id = id
	return n
}

// Key 获取键
func (n *NodeInfo) Key() string {
	return fmt.Sprintf("services:%s:%s", n.Name, n.Id)
}

// Value 获取值
func (n *NodeInfo) Value() string {
	value, _ := json.Marshal(n)
	return string(value)
}

func (n *NodeInfo) GetEndpoints() []string {
	var endpoints []string
	for _, u := range n.Endpoints {
		endpoints = append(endpoints, u.String())
	}
	return endpoints
}

var Node = NewNodeInfo()
