package main

import (
	"C"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/zalando/go-keyring"
)

var proxy1 *ProxyService

func main() {}

//export RunMain
func RunMain(input *C.char) {
	fmt.Println("proxy1:", proxy1) // check if it's nil or has a value
	if proxy1 != nil {
		proxy1.Close()
	}
	serverPeerID := C.GoString(input)
	fmt.Println("Received string from C:", serverPeerID)
	service := "p2pmax"
	user := "user"
	// Retrieve the secret
	secret, err := keyring.Get(service, user)
	var privk crypto.PrivKey
	if err != nil {
		fmt.Println("Error retrieving secret:", err)
		privKey, _, err := GeneratePeerKey()
		if err != nil {
			fmt.Println("Error GeneratePeerKey", err)
			return
		}
		// save it into keychain
		keyring.Set(service, user, privKey)
		privk, _ = ReadPeerKey(privKey)
	}else {
		privk, _ = ReadPeerKey(secret)
	}

	ctx := ContextWithSignal(context.Background())

	var opts []libp2p.Option = []libp2p.Option{
		libp2p.Identity(privk),
		libp2p.UserAgent(ServiceName),
		libp2p.EnableRelay(),
		libp2p.EnableHolePunching(),
		libp2p.WithDialTimeout(time.Second * 60),
	}

	opts = append(opts,
		libp2p.NoListenAddrs,
	)
	host, err := libp2p.New(opts...)
	if err != nil {
		Log.Fatal(err)
	}

	fmt.Printf("Peer ID: %s\n", host.ID())
	serverPeer := &peer.AddrInfo{ID: host.ID()}

	serverPeer, err = peer.AddrInfoFromString(serverPeerID)
	if err != nil {
		Log.Fatal(err)
	}

	// host.Peerstore().AddAddrs(serverPeer.ID, serverPeer.Addrs, peerstore.PermanentAddrTTL)
	ctxt, cancel := context.WithTimeout(ctx, time.Second*5)
	if err = host.Connect(ctxt, *serverPeer); err != nil {
		Log.Fatal(err)
	}
	res := <-ping.Ping(ctxt, host, serverPeer.ID)
	if res.Error != nil {
		Log.Fatalf("ping error: %v", res.Error)
	} else {
		Log.Infof("ping RTT: %s", res.RTT)
	}
	cancel()
	host.ConnManager().Protect(serverPeer.ID, "proxy")
	proxy1 = NewProxyService(ctx, host, "")
	go func() {
		if err := proxy1.Serve("localhost:1082", serverPeer.ID); err != nil {
			Log.Fatal(err)
		}
	}()


}

func ContextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		cancel()
	}()
	return newCtx
}
