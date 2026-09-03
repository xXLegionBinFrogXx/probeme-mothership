// Package config parses flags and PME_* environment variables.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Default filter lists, mirroring node_exporter.
const (
	DefaultMountPointsExclude = "/dev,/proc,/sys,/run,/var/lib/docker"
	DefaultFSTypesExclude     = "autofs,binfmt_misc,bpf,cgroup,cgroup2,configfs,debugfs,devpts," +
		"devtmpfs,fusectl,hugetlbfs,mqueue,nsfs,overlay,proc,procfs,pstore," +
		"rpc_pipefs,securityfs,selinuxfs,squashfs,sysfs,tracefs"
	DefaultNetDevicesExclude  = "lo,veth,docker,br-,virbr"
	DefaultDiskDevicesExclude = "loop,ram,dm-"
)

type Filters struct {
	MountPointsExclude []string
	FSTypesExclude     []string
	NetDevicesExclude  []string
	DiskDevicesExclude []string
}

type Config struct {
	Providers       []string
	ProviderDir     string
	ProbeInterval   time.Duration
	ProbeTimeout    time.Duration
	ListenAddress   string
	LogLevel        string
	RequireProvider bool
	Filters         Filters
}

func Load(args []string) (*Config, bool, error) {
	fs := flag.NewFlagSet("probeme-mothership", flag.ContinueOnError)

	var (
		versionFlag     bool
		providers       string
		providerDir     string
		interval        = fs.String("probe.interval", envOr("PME_PROBE_INTERVAL", "5s"), "collect interval (env PME_PROBE_INTERVAL)")
		timeout         = fs.String("probe.timeout", envOr("PME_PROBE_TIMEOUT", "3s"), "collect timeout (env PME_PROBE_TIMEOUT)")
		listen          = fs.String("web.listen-address", envOr("PME_WEB_LISTEN_ADDRESS", "127.0.0.1:9167"), "listen address (env PME_WEB_LISTEN_ADDRESS)")
		logLevel        = fs.String("log.level", envOr("PME_LOG_LEVEL", "info"), "log level: debug, info, warn, error (env PME_LOG_LEVEL)")
		requireProvider bool
		mountExcludes   = fs.String("collector.filesystem.mountpoints-exclude", envOr("PME_FS_MOUNTPOINTS_EXCLUDE", DefaultMountPointsExclude), "prefix exclude list for mountpoints")
		fstypeExcludes  = fs.String("collector.filesystem.fstypes-exclude", envOr("PME_FS_TYPES_EXCLUDE", DefaultFSTypesExclude), "prefix exclude list for filesystem types")
		netdevExcludes  = fs.String("collector.netdev.devices-exclude", envOr("PME_NETDEV_DEVICES_EXCLUDE", DefaultNetDevicesExclude), "prefix exclude list for net devices")
		diskExcludes    = fs.String("collector.disk.devices-exclude", envOr("PME_DISK_DEVICES_EXCLUDE", DefaultDiskDevicesExclude), "prefix exclude list for disk devices")
	)

	fs.BoolVar(&versionFlag, "version", false, "print version and exit")
	fs.StringVar(&providers, "providers", envOr("PME_PROVIDERS", ""), "comma-separated provider .so paths; overrides provider.dir discovery")
	fs.StringVar(&providerDir, "provider.dir", envOr("PME_PROVIDER_DIR", "/usr/local/lib"), "directory globbed for libprobeme_*.so (env PME_PROVIDER_DIR)")
	fs.BoolVar(&requireProvider, "require-provider", envBool("PME_REQUIRE_PROVIDER", false), "exit 2 when no usable provider is found")

	if err := fs.Parse(args); err != nil {
		return nil, false, err
	}

	cfg := &Config{
		Providers:       splitList(providers),
		ProviderDir:     providerDir,
		ListenAddress:   *listen,
		LogLevel:        *logLevel,
		RequireProvider: requireProvider,
		Filters: Filters{
			MountPointsExclude: splitList(*mountExcludes),
			FSTypesExclude:     splitList(*fstypeExcludes),
			NetDevicesExclude:  splitList(*netdevExcludes),
			DiskDevicesExclude: splitList(*diskExcludes),
		},
	}

	var err error
	if cfg.ProbeInterval, err = time.ParseDuration(*interval); err != nil {
		return nil, false, fmt.Errorf("--probe.interval: %w", err)
	}
	if cfg.ProbeTimeout, err = time.ParseDuration(*timeout); err != nil {
		return nil, false, fmt.Errorf("--probe.timeout: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, false, err
	}
	return cfg, versionFlag, nil
}

func (c *Config) validate() error {
	if c.ProbeInterval <= 0 {
		return fmt.Errorf("--probe.interval must be > 0, got %s", c.ProbeInterval)
	}
	if c.ProbeTimeout <= 0 {
		return fmt.Errorf("--probe.timeout must be > 0, got %s", c.ProbeTimeout)
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("--log.level must be one of debug, info, warn, error; got %q", c.LogLevel)
	}
	if c.ProviderDir == "" && len(c.Providers) == 0 {
		return fmt.Errorf("no providers given and --provider.dir empty")
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func splitList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
