package i18n

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	// SDKVersion is a string which specifies the current SDK version.
	SDKVersion = "v1.0.0"

	// Domain is the starling domain which can be accessed publicly.
	Domain = "starling-public.snssdk.com"

	httpProxyMetricsKeyThroughput = "proxy.http.throughput"
	httpProxyMetricsKeyLatency    = "proxy.http.latency"
	clientRetrieveErrorMetricsKey = "client.retrieve.error"
	clientPackageEmptyMetricsKey  = "client.package.empty"
	clientKeyEmptyMetricsKey      = "client.key.empty"

	defaultLeftDelimiter   = "{"
	defaultRightDelimiter  = "}"
	defaultRefreshInterval = time.Minute
	defaultCacheDuration   = time.Hour * 6
)

const (
	// EnvNormal is the default environment which is used for online production.
	EnvNormal = "normal"
	// EnvGray is used for verification during the developing process.
	EnvGray = "gray"
	// EnvTest is used to get the offline test data for testing.
	EnvTest = "test"
)

var (
	ErrInvalidParams      = errors.New("invalid given params")
	ErrKeyNotExist        = errors.New("given key not exist")
	ErrBackToSourceFailed = errors.New("back to source to fetch data failed")
	ErrInvalidICUFormat   = errors.New("invalid ICU format string")
)

var (
	// Platform stores the runtime environment when running this SDK.
	Platform string

	// LocalIP stores the ip address of the current machine.
	LocalIP string

	// LocalLang stores the language of the current machine.
	LocalLang string

	// Set the global json library and make compatible with std json library.
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	// Option object pool for reuse.
	op = optionPool{Pool: sync.Pool{New: func() interface{} { return &option{} }}}
)

func init() {
	// Get runtime environment platform information.
	Platform = runtime.GOOS + "/" + runtime.GOARCH + "/" + runtime.Version()

	// Try to get local IP address and use 127.0.0.1 as default if failed.
	var privateNets []net.IPNet
	for _, s := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7"} {
		_, ipnet, _ := net.ParseCIDR(s)
		privateNets = append(privateNets, *ipnet)
	}
	var localIPList []net.IP
	if ifs, err := net.Interfaces(); err == nil {
		for _, netIface := range ifs {
			if netIface.Flags&net.FlagUp == 0 ||
				netIface.Flags&net.FlagLoopback != 0 ||
				netIface.Flags&net.FlagPointToPoint != 0 {
				continue
			}
			addrs, err := netIface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				ipnet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				ip := ipnet.IP
				if ip.To16() == nil || ip.IsLoopback() || ip.IsMulticast() {
					continue
				}
				var isPrivate bool
				for i := range privateNets {
					if privateNets[i].Contains(ip) {
						isPrivate = true
						break
					}
				}
				if !isPrivate {
					localIPList = append(localIPList, ip)
				}
			}
		}
	}
	if len(localIPList) != 0 {
		LocalIP = localIPList[0].String()
	}
	if len(LocalIP) == 0 {
		LocalIP = "127.0.0.1"
	}

	// Try to get local language on different systems and use Chinese as default.
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
		output, err := cmd.Output()
		if err == nil {
			langLoc := strings.Split(strings.TrimSpace(string(output)), "-")
			LocalLang = langLoc[0]
		}
	case "darwin":
		cmd := exec.Command("sh", "osascript -e 'user locale of (get system info)'")
		output, err := cmd.Output()
		if err == nil {
			langLoc := strings.Split(strings.TrimSpace(string(output)), "_")
			LocalLang = langLoc[0]
		}
	case "linux":
		if envlang, ok := os.LookupEnv("LANG"); ok {
			langLocRaw := strings.Split(strings.TrimSpace(envlang), ".")[0]
			langLoc := strings.Split(langLocRaw, "_")
			LocalLang = langLoc[0]
		}
	}
	if len(LocalLang) == 0 || LocalLang == "zh" {
		LocalLang = "zh-Hans"
	}
}
