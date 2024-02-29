package main

import (
	"context"
	"errors"
	"github.com/pion/turn/v3"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"path"
)

type Auth struct {
	cli *redis.Client
}

func NewAuth(redisAddr, redisPassword string) *Auth {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
	})
	return &Auth{cli: client}
}

func (a *Auth) authHandler(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
	var srcIP string
	switch srcAddr.(type) {
	case *net.UDPAddr:
		srcIP = srcAddr.(*net.UDPAddr).IP.String()
	case *net.TCPAddr:
		srcIP = srcAddr.(*net.TCPAddr).IP.String()
	case *net.IPAddr:
		srcIP = srcAddr.(*net.IPAddr).IP.String()
	}
	if srcIP == "" {
		log.Println("unexpected empty ip: ", srcAddr.String())
		return nil, false
	}
	authKey := getAuthKey(username, realm, srcIP)
	result, err := a.cli.Get(context.Background(), authKey).Result()
	if errors.Is(err, redis.Nil) {
		log.Println("user:", username, "from:", srcAddr.String(), "authorization not found")
		return nil, false
	}
	if err != nil {
		log.Println("redis get error:", err)
		return nil, false
	}
	key := turn.GenerateAuthKey(username, realm, result)
	log.Println("user:", username, "from:", srcAddr.String(), "authorized")
	return key, true
}

func getAuthKey(username string, realm string, ipAddr string) string {
	return path.Join("turn", realm, username, ipAddr)
}
