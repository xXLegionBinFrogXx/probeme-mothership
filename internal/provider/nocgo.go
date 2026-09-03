//go:build !cgo

// Stubs so downstream packages type-check with CGO_ENABLED=0; the go tool
// excludes the cgo files in that mode.
package provider

import (
	"errors"
	"time"
)

const ABIMajor = 1

var (
	ErrStructTooSmall = errors.New("provider struct smaller than expected")
	ErrABIMismatch    = errors.New("provider ABI major mismatch")
	ErrNotSupported   = errors.New("operation not supported")
	ErrIO             = errors.New("I/O error")
	ErrInvalid        = errors.New("invalid argument")
	ErrNotInit        = errors.New("provider not initialized")
	ErrCgoDisabled    = errors.New("probeme: built with CGO_ENABLED=0; providers unavailable")
)

type Provider struct{ name string }

func (p *Provider) Name() string              { return p.name }
func (p *Provider) Path() string              { return "" }
func (p *Provider) Capabilities() uint64      { return 0 }
func (p *Provider) ABIVersion() uint32        { return 0 }
func (p *Provider) Init(uint64) error         { return ErrCgoDisabled }
func (p *Provider) CollectAll(*CBuffer) error { return ErrCgoDisabled }
func (p *Provider) Destroy()                  {}
func (p *Provider) Close()                    {}

type CBuffer struct{}

func Open(string) (*Provider, error) { return nil, ErrCgoDisabled }

func NewCBuffer() (*CBuffer, error) { return nil, ErrCgoDisabled }
func (b *CBuffer) Reset()           {}
func (b *CBuffer) Snapshot() *Snapshot {
	return &Snapshot{PublishedAt: time.Now()}
}
