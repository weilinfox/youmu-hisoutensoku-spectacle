package main

import (
	"net"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/weilinfox/youmu-hisoutensoku-protocol/lib"
)

func main() {

	file, _ := os.OpenFile("log.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	defer file.Close()
	logrus.SetOutput(file)

	udpLAddr, _ := net.ResolveUDPAddr("udp", "0.0.0.0:4646")
	udpRAddr, _ := net.ResolveUDPAddr("udp", "localhost:10800")
	udpLConn, _ := net.ListenUDP("udp", udpLAddr)
	udpRConn, _ := net.DialUDP("udp", nil, udpRAddr)

	lib.Sync(udpRConn, udpLConn)
}
