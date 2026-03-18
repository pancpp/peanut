package conf

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	gVersion     string
	gBuildTime   string
	gGitHash     string
	gBuildNumber string
)

var (
	gConf *viper.Viper
)

func init() {
	// flags
	var (
		showVersion bool
		confDirPath string
	)

	pflag.BoolVarP(&showVersion, "version", "V", false, "Show version information.")
	pflag.StringVarP(&confDirPath, "config", "c", "/etc/peanut/peanut.yaml", "Config file path.")
	pflag.Parse()

	if showVersion {
		fmt.Println("Version     :", gVersion)
		fmt.Println("BuildTime   :", gBuildTime)
		fmt.Println("GitHash     :", gGitHash)
		fmt.Println("Build number:", gBuildNumber)
		os.Exit(0)
	}

	// viper
	conf := viper.New()

	// set default values
	conf.SetDefault("log.enable_console_log", false)
	conf.SetDefault("log.path", "/var/log/peanut/peanut.log")
	conf.SetDefault("log.max_size", 500)
	conf.SetDefault("log.max_backups", 3)
	conf.SetDefault("log.local_time", true)
	conf.SetDefault("log.compress", true)

	conf.SetDefault("p2p.private_key_path", "/etc/peanut/peanut-private-key.b64")
	conf.SetDefault("p2p.pnet_psk_path", "")
	conf.SetDefault("p2p.listen_multiaddrs", []string{
		"/ip4/0.0.0.0/udp/19882/quic-v1",
	})
	conf.SetDefault("p2p.discovery_multiaddrs", []string{
		"/dns4/disc.cpplint.com/udp/19880/quic-v1/p2p/12D3KooWEJifdgZLnVo87Pres6UqT3wZTGm3T6sVPFxnBQdszvEP",
	})
	conf.SetDefault("p2p.relay_multiaddrs", []string{
		"/dns4/relay1.cpplint.com/udp/19881/quic-v1/p2p/12D3KooWCGzYPZzNror5zms8bP2QciBqs7BTc6GFThDGyXyR1ghZ",
	})
	conf.SetDefault("p2p.allowlist_path", "/etc/peanut/peanut-allowlist.yaml")
	conf.SetDefault("p2p.enable_holepunch_tracer", false)
	conf.SetDefault("p2p.enable_conn_notifier", false)

	conf.SetDefault("vpn.tun_name", "peanut0")

	// set file path
	conf.SetConfigFile(confDirPath)
	gConf = conf
}

func Init() error {
	confFilePath := gConf.ConfigFileUsed()

	// check config directory path
	confDir := path.Dir(confFilePath)
	if _, err := os.Stat(confDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(confDir, 0755); err != nil {
			return err
		}
	}

	// check config file path
	if f, err := os.OpenFile(confFilePath, os.O_CREATE|os.O_RDONLY, 0600); err != nil {
		return err
	} else {
		f.Close()
	}

	// read config file
	if err := gConf.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func GetBool(k string) bool {
	return gConf.GetBool(k)
}

func GetInt(k string) int {
	return gConf.GetInt(k)
}

func GetString(k string) string {
	return gConf.GetString(k)
}

func GetFloat64(k string) float64 {
	return gConf.GetFloat64(k)
}

func GetStringSlice(k string) []string {
	return gConf.GetStringSlice(k)
}
