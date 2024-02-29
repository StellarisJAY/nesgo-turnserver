package main

import (
	"flag"
	"fmt"
	"github.com/pion/logging"
	"github.com/pion/turn/v3"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	realm         string
	port          int
	publicIP      string
	redisAddr     string
	redisPassword string

	minPort uint
	maxPort uint
)

func init() {
	flag.StringVar(&publicIP, "public-ip", "", "relay public address ip")
	flag.StringVar(&realm, "realm", "nesgo", "turn server realm")
	flag.IntVar(&port, "port", 3478, "turn server port")
	flag.StringVar(&redisAddr, "redis", "0.0.0.0:6379", "redis server address")
	flag.StringVar(&redisPassword, "redis-pass", "", "redis password")
	flag.UintVar(&minPort, "min-port", 30000, "port range min")
	flag.UintVar(&maxPort, "max-port", 35000, "port range max")
	flag.Parse()
}

func main() {
	udpListener, err := net.ListenPacket("udp4", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		panic(fmt.Errorf("unable to create udp listener, error: %v", err))
	}
	auth := NewAuth(redisAddr, redisPassword)
	s, err := turn.NewServer(turn.ServerConfig{
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorPortRange{
					RelayAddress: net.ParseIP(publicIP),
					Address:      "0.0.0.0",
					MinPort:      uint16(minPort),
					MaxPort:      uint16(maxPort),
				},
			},
		},
		LoggerFactory: logging.NewDefaultLoggerFactory(),
		Realm:         realm,
		AuthHandler:   auth.authHandler,
	})
	if err != nil {
		panic(err)
	}
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	_ = s.Close()
}
