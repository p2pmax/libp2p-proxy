package main

import (
	"net"
	"fmt"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (p *ProxyService) Serve(ln net.Listener, remotePeer peer.ID) error {

	p.socks = ln
	go p.Wait(ln.Close)

	for {
		conn, err := ln.Accept()
		select {
		case <-p.stop:
			// Listener was closed, exit the loop
			return nil
		default:
			if err != nil {
				// If the listener is closed, log and exit the loop
				if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
					fmt.Println("Listener closed, stopping accept loop")
					return nil
				}
				fmt.Println("Error accepting connection:", err)
				continue
			}
			// Handle the connection
			go p.sideHandler(conn, remotePeer)
		}

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
