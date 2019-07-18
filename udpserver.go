package p2p

import (
	// "bytes"
	// "encoding/gob"

	"log"
	"net"
	// "time"
)

//InitUDPServer .
func InitUDPServer() (*net.UDPConn, *net.UDPAddr) {
	ip := "127.0.0.1" //getOutboundIP()
	//0 assigns a random port
	port := "40000"
	udpAddr, err := net.ResolveUDPAddr("udp4", ip+":"+port)
	if err != nil {
		log.Fatal(err)
	}
	udpServer, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		udpAddr, _ = net.ResolveUDPAddr("udp4", ip+":"+"0")
		udpServer, _ = net.ListenUDP("udp", udpAddr)
		udpAddr = udpServer.LocalAddr().(*net.UDPAddr)
	}
	return udpServer, udpAddr
}

//getOutboundIP .
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
