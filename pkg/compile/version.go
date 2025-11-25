package compile

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/rs/zerolog/log"
)

var (
	Name     = "" // 项目名
	Id       = "" // 项目ID, Hostname.Name
	Hostname = ""
	IpAddr   = "" // 内网IP地址

	Version   = ""
	GoVersion = runtime.Version()
	GoOs      = runtime.GOOS
	GoArch    = runtime.GOARCH
	GitCommit = ""
	BuildTime = ""
)

func init() {
	GoVersion = runtime.Version()

	Hostname, _ = os.Hostname()
	addrs, _ := net.InterfaceAddrs()
	Id = fmt.Sprintf("%s.%s", Hostname, Name)

	// 获取内网IP地址
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsPrivate() {
			if ipnet.IP.To4() != nil {
				IpAddr = ipnet.IP.String()
				break
			}
		}
	}

	Node.AddId(Id).AddName(Name).AddVersion(Version)
}

func Os() string {
	return fmt.Sprintf("%s/%s", GoOs, GoArch)
}

func Print() {
	fmt.Printf("Id: %s\nVersion: %s\nGo Version: %s\nOS: %s\nGit Commit: %s\nBuild Time: %s\n", Id, Version, GoVersion, Os(), GitCommit, BuildTime)
}

func Log() {
	log.Info().Str("id", Id).Str("version", Version).Str("go_version", GoVersion).Str("os", Os()).Str("commit", GitCommit).Str("build_time", BuildTime).Msg("build info")
}
