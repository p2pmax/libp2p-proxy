package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"syscall"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	P2PHttpID   protocol.ID = "/http"
	ID          protocol.ID = "/p2pdao/libp2p-proxy/1.0.0"
	ServiceName string      = "p2pdao.libp2p-proxy"
)

var Log = logging.Logger("libp2p-proxy")

type ProxyService struct {
	ctx     context.Context
	host    host.Host
	http    *http.Server
	p2pHost string
}

func NewProxyService(ctx context.Context, h host.Host, p2pHost string) *ProxyService {
	ps := &ProxyService{ctx, h, nil, p2pHost}
	h.SetStreamHandler(ID, ps.Handler)
	return ps
}

// Close terminates this listener. It will no longer handle any
// incoming streams
func (p *ProxyService) Close() error {
	if s := p.http; s != nil {
		c, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		p.http = nil
		s.Shutdown(c)
	}
	return p.host.Close()
}

func (p *ProxyService) Wait(fn func() error) error {
	<-p.ctx.Done()
	defer p.Close()

	if fn != nil {
		if err := fn(); err != nil {
			return err
		}
	}
	return p.ctx.Err()
}

func (p *ProxyService) Handler(s network.Stream) {
	if err := s.Scope().SetService(ServiceName); err != nil {
		Log.Errorf("error attaching stream to service: %s", err)
		s.Reset()
		return
	}

	p.handler(NewBufReaderStream(s))
}

func (p *ProxyService) handler(bs *BufReaderStream) {
	defer bs.Close()

}

func shouldLogError(err error) bool {
	return err != nil && err != io.EOF &&
		err != io.ErrUnexpectedEOF && err != syscall.ECONNRESET &&
		!strings.Contains(err.Error(), "timeout") &&
		!strings.Contains(err.Error(), "reset") &&
		!strings.Contains(err.Error(), "closed")
}
