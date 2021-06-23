package starling_goclient_public

import (
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

	defaultRefreshInterval = time.Second * 10
	defaultCacheDuration   = time.Hour * 6

	apiVersionGetPackage        = "V1"
	apiVersionSimilarFallback   = "V2"
	apiVersionGetPackageVersion = "V3"

	// apiVersion(V1|V2|V3), mode, projectName, namespace, lang
	keyFmt = "/%s/%s/%s/%s/%s"        
	// apiVersion(V1|V2|V3), mode, projectID, namespaceID, lang, version, langType
	keyFmtVersion = "/%s/%s/%s/%s/%s/%s/%s/"
)

// ModeType defines a custom type which represents the mode to retrieve process.
type ModeType string

// FallbackType defines a custom type which represents fallback language strategy.
type FallbackType uint8

const (
	// ModeNormal is the default mode which is used for online production.
	ModeNormal ModeType = "normal"
	// ModeGray is used for verification during the developing process.
	ModeGray ModeType = "gray"
	// ModeTest is to retreive the offline test data for testing.
	ModeTest ModeType = "test"

	// FallbackLangNone uses no fallback strategy.
	FallbackLangNone FallbackType = 0
	// FallbackLangDefault uses the default fallback strategy.
	FallbackLangDefault FallbackType = 1
	// FallbackLangSimilar uses the similar language fallback strategy.
	FallbackLangSimilar FallbackType = 2
	// FallbackLangCustom uses customized fallback strategy.
	FallbackLangCustom FallbackType = 3
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
