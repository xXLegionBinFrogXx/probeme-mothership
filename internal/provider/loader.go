package provider

import (
	"fmt"
	"log/slog"
	"path/filepath"
)

// Options configures Load.
type Options struct {
	Explicit  []string
	Dir       string
	InitFlags uint64
	Logger    *slog.Logger
}

// Loaded pairs a provider with the capabilities it owns after dedupe
// (first provider to claim a capability wins).
type Loaded struct {
	Provider *Provider
	Caps     uint64
}

func Load(opts Options) ([]*Loaded, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	var paths []string
	if len(opts.Explicit) > 0 {
		paths = opts.Explicit
	} else {
		glob, err := filepath.Glob(filepath.Join(opts.Dir, "libprobeme_*.so"))
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", opts.Dir, err)
		}
		paths = glob
		if len(paths) == 0 {
			logger.Warn("no providers found", "dir", opts.Dir)
			return nil, nil
		}
	}

	var (
		claimed uint64
		loaded  = make([]*Loaded, 0, len(paths))
	)
	for _, path := range paths {
		p, err := Open(path)
		if err != nil {
			if len(opts.Explicit) > 0 {
				return nil, err
			}
			logger.Warn("skipping provider", "path", path, "err", err)
			continue
		}
		if err := p.Init(opts.InitFlags); err != nil {
			p.Close()
			if len(opts.Explicit) > 0 {
				return nil, err
			}
			logger.Warn("skipping provider: init failed", "provider", p.Name(), "err", err)
			continue
		}

		caps := p.Capabilities()
		if dup := caps & claimed; dup != 0 {
			logger.Info("capabilities already claimed, masking",
				"provider", p.Name(), "duplicates", CapNamesFor(dup))
			caps &^= dup
		}
		if caps == 0 {
			logger.Info("provider owns no capabilities, skipping", "provider", p.Name())
			p.Close()
			continue
		}
		claimed |= caps
		loaded = append(loaded, &Loaded{Provider: p, Caps: caps})

		abi := p.ABIVersion()
		logger.Info("provider loaded",
			"provider", p.Name(),
			"path", p.Path(),
			"abi", fmt.Sprintf("%d.%d", abi>>16, abi&0xffff),
			"capabilities", CapNamesFor(caps))
	}
	return loaded, nil
}
