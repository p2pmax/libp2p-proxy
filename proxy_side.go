package main

import (
	"net"

	"github.com/libp2p/go-libp2p/core/peer"
)

func (p *ProxyService) Serve(proxyAddr string, remotePeer peer.ID) error {
	ln, err := net.Listen("tcp", proxyAddr)
	if err != nil {
		return err
	}
	p.socks = ln
	go p.Wait(ln.Close)

	for {
		conn, err := ln.Accept()
		if err := p.ctx.Err(); err != nil {
			return err
		}

		if err != nil {
			return err
		}
		go p.sideHandler(conn, remotePeer)
	}
}

func (p *ProxyService) sideHandler(conn net.Conn, remotePeer peer.ID) {
	defer conn.Close()

	s, err := p.host.NewStream(p.ctx, remotePeer, ID)
	if err != nil {
		Log.Errorf("creating stream to %s error: %v", remotePeer, err)
		return
	}

	defer s.Close()
	if err := tunneling(s, conn); shouldLogError(err) {
		Log.Warn(err)
	}
}
