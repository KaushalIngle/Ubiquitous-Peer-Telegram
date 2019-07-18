package p2p

import (
	"log"
	"net"
	"sync"
	"time"
)

//Gateway .
type Gateway struct {
	localAddr        *net.UDPAddr
	server           *net.UDPConn
	recieve, send    chan []string
	ssend, srecieve  chan *Packet
	connList         chan *net.UDPAddr
	uuidMap, connMap *sync.Map
	debug            chan string
}

//Debug .
func (gateway *Gateway) Debug() string {
	return <-gateway.debug
}

//InitGatewayServer .
func InitGatewayServer() *Gateway {

	time.Sleep(2000 * time.Millisecond)
	gateway := &Gateway{}
	gateway.server, gateway.localAddr = InitUDPServer()
	gateway.recieve = make(chan []string, 30)
	gateway.send = make(chan []string, 10)
	gateway.srecieve = make(chan *Packet, 300)
	gateway.ssend = make(chan *Packet, 100)
	gateway.connList = make(chan *net.UDPAddr, 1000)
	gateway.uuidMap = generateUUIDMap()
	gateway.connMap = generateUUIDMap()
	gateway.debug = make(chan string, 10)

	gateway.debug <- "gateway initialized"

	go gateway.InitiateConnList()
	go gateway.MaintainConnectionCount()
	//go gateway.MaintainConnections()
	go gateway.SendRoutine()    //(gateway.server, gateway.send, gateway.connList, gateway.debug)
	go gateway.SendSRoutine()   //(gateway.server, gateway.ssend, gateway.debug)
	go gateway.RecieveRoutine() //(gateway.server, gateway.recieve, gateway.srecieve, gateway.connList, gateway.debug)
	go gateway.ServicePackets() //(gateway.srecieve, gateway.ssend, gateway.connList, gateway.uuidMap, gateway.debug)
	return gateway
}

//InitiateConnList .
func (gateway *Gateway) InitiateConnList() {
	ip := "127.0.0.1"
	port := "40000"
	bootstrapAddr, err := net.ResolveUDPAddr("udp4", ip+":"+port)
	if err != nil {
		log.Fatal(err)
	}
	if gateway.localAddr.String() != bootstrapAddr.String() {
		gateway.connList <- bootstrapAddr
		gateway.connList <- bootstrapAddr

		//gateway.GetConnections(bootstrapAddr)
	}
	// GetConnections(gateway.localAddr, bootstrapAddr, gateway.ssend, gateway.connList, gateway.uuidMap)
	// GetConnections(gateway.localAddr, bootstrapAddr, gateway.ssend, gateway.connList, gateway.uuidMap)
	// GetConnections(gateway.localAddr, bootstrapAddr, gateway.ssend, gateway.connList, gateway.uuidMap)
	// GetConnections(gateway.localAddr, bootstrapAddr, gateway.ssend, gateway.connList, gateway.uuidMap)
}

// //MaintainConnections .
// func (gateway *Gateway) MaintainConnections() {
// 	for {
// 		addrMap := make(map[string]*net.UDPAddr)
// 		for addr := range gateway.connList {
// 			if _, value := addrMap[(*addr).String()]; !value {
// 				addrMap[(*addr).String()] = addr
// 			}
// 		}
// 		for _, addr := range addrMap {
// 			gateway.connList <- addr
// 		}
// 		time.Sleep(60 * time.Second)
// 	}
// }

//MaintainConnectionCount .
func (gateway *Gateway) MaintainConnectionCount() {
	count := 0
	for {
		if count < 3 {
			if len(gateway.connList) < (cap(gateway.connList) / 3) {
				select {
				case udpAddr := <-gateway.connList:
					count++
					if gateway.CheckConnection(udpAddr) {
						gateway.connList <- udpAddr
						if gateway.GetConnections(udpAddr) {
							count = 0
							// time.Sleep(30000 * time.Millisecond)
						}
					}
				case <-time.After(1 * time.Millisecond):
					count++
				}
			}
		} else {
			gateway.InitiateConnList()
		}
		time.Sleep(30000 * time.Millisecond)
	}
}

//GetConnections .
func (gateway *Gateway) GetConnections(reciever *net.UDPAddr) bool {
	spacket := InitServicePacket()
	spacket.SAddr = *gateway.localAddr
	spacket.RAddr = *reciever
	spacket.Data = []string{}
	spacket.Data = append(spacket.Data, "GetConn")
	replychan := make(chan *Packet)
	gateway.uuidMap.Store(spacket.UUID, replychan)
	gateway.ssend <- spacket
	// gateway.debug <- "getconns sent"
	select {
	case rpacket := <-replychan:
		if rpacket.Data[0] == "Conns" {
			// gateway.debug <- "conns recieved"
			//if rpacket.SAddr.String() == spacket.RAddr.String() && rpacket.RAddr.String() == spacket.SAddr.String() {
			for _, connString := range rpacket.Data[1:] {
				udpAddr, err := net.ResolveUDPAddr("udp4", connString)
				if err != nil {
					// gateway.debug <- "udpaddr not read"
					continue

					//need to add err
				}
				if gateway.localAddr.String() != udpAddr.String() {
					if len(gateway.connList) != cap(gateway.connList) {
						gateway.connList <- udpAddr
					}
				}
			}
			//}
			gateway.uuidMap.Delete(spacket.UUID)
			return true
		}
		gateway.uuidMap.Delete(spacket.UUID)
		return false
	case <-time.After(2000 * time.Millisecond):
		gateway.uuidMap.Delete(spacket.UUID)
		// debug <- "conns timeout"
		return false
	}
}

//CheckConnection .
func (gateway *Gateway) CheckConnection(reciever *net.UDPAddr) bool {
	spacket := InitServicePacket()
	spacket.SAddr = *gateway.localAddr
	spacket.RAddr = *reciever
	spacket.Data = []string{}
	spacket.Data = append(spacket.Data, "Ping")
	replychan := make(chan *Packet)
	gateway.uuidMap.Store(spacket.UUID, replychan)
	gateway.ssend <- spacket
	// gateway.debug <- "ping sent"
	select {
	case rpacket := <-replychan:
		if rpacket.Data[0] == "Pong" {
			// gateway.debug <- "pong recieved"
			//if rpacket.SAddr.String() == spacket.RAddr.String() && rpacket.RAddr.String() == spacket.SAddr.String() {
			gateway.uuidMap.Delete(spacket.UUID)
			return true
			//}
		}
		gateway.uuidMap.Delete(spacket.UUID)
		return false
	case <-time.After(1000 * time.Millisecond):
		gateway.uuidMap.Delete(spacket.UUID)
		// gateway.debug <- "pong timeout"
		return false
	}

}

//SendRoutine .
func (gateway *Gateway) SendRoutine() { //(conn *net.UDPConn, send chan *Packet, sendList chan *net.UDPAddr, debug chan string) {
	for {
		data := <-gateway.send
		packet := InitPacket()
		packet.UUID = getUUID()
		packet.Stype = false
		packet.Data = data
		packet.SAddr = *gateway.localAddr

		// size := unsafe.Sizeof(message)
		// gateway.debug <- string(size)
		addr := <-gateway.connList
		for {

			if addr.String() != gateway.localAddr.String() {
				break
			} else {
				addr = <-gateway.connList
			}
		}

		packet.RAddr = *addr
		message := packet.ToBytes()
		_, err := gateway.server.WriteToUDP(message, addr)
		if err != nil {
			log.Fatal(err)
		}
		// gateway.debug <- "sent packet" + addr.String()
		gateway.connList <- addr
	}
}

//SendSRoutine .
func (gateway *Gateway) SendSRoutine() { //(conn *net.UDPConn, ssend chan *Packet, debug chan string) {
	for {
		packet := <-gateway.ssend
		message := packet.ToBytes()
		//gateway.debug <- string(message)
		_, err := gateway.server.WriteToUDP(message, &packet.RAddr)
		if err != nil {
			log.Fatal(err)
		}
		gateway.debug <- "sent spacket" + packet.RAddr.String()
	}
}

//RecieveRoutine .
func (gateway *Gateway) RecieveRoutine() { //(conn *net.UDPConn, recieve chan *Packet, srecieve chan *Packet, connList chan *net.UDPAddr, debug chan string) {
	for {
		buffer := make([]byte, 100024)
		// gateway.debug <- "going to read"
		n, addr, err := gateway.server.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}
		// gateway.debug <- "read something"
		packetBytes := buffer[:n]
		packet := InitPacket()
		err = packet.FromBytes(packetBytes)
		if err != nil {
			log.Fatal(err)
		}
		if packet.Stype == true {
			gateway.srecieve <- packet
			// gateway.debug <- "recieved spacket"
		} else {
			if addr.String() != gateway.localAddr.String() {
				// if len(gateway.connList) != cap(gateway.connList) {
				// 	gateway.connList <- addr
				// }
				gateway.recieve <- append([]string{packet.UUID}, packet.Data...)
				// gateway.debug <- "recieved packet"
			}
		}
	}
}

//ServicePackets ..
func (gateway *Gateway) ServicePackets() { //(srecieve chan *Packet, ssend chan *Packet, connList chan *net.UDPAddr, uuidMap *sync.Map, debug chan string) {
	for {
		spacket := <-gateway.srecieve
		var rchan chan *Packet
		tchan, ans := gateway.uuidMap.Load(spacket.UUID)

		if ans {
			rchan = tchan.(chan *Packet)
			rchan <- spacket
			// gateway.debug <- "a ans recieved"
		} else {

			if spacket.Data[0] == "Ping" {
				spacket.RAddr, spacket.SAddr = spacket.SAddr, spacket.RAddr
				spacket.Data[0] = "Pong"
				gateway.ssend <- spacket
				if (&spacket.RAddr).String() != (gateway.localAddr).String() {
					if len(gateway.connList) != cap(gateway.connList) {
						gateway.connList <- &spacket.RAddr
					}
				}
				// gateway.debug <- "pong sent"

			} else if spacket.Data[0] == "GetConn" {
				spacket.RAddr, spacket.SAddr = spacket.SAddr, spacket.RAddr
				spacket.Data[0] = "Conns"
				for i := 1; i <= 5; i++ {
					select {
					case conn := <-gateway.connList:
						spacket.Data = append(spacket.Data, conn.String())
						gateway.connList <- conn
					case <-time.After(100 * time.Millisecond):
						continue
					}
				}
				gateway.ssend <- spacket
				if (&spacket.RAddr).String() != (gateway.localAddr).String() {
					if len(gateway.connList) != cap(gateway.connList) {
						gateway.connList <- &spacket.RAddr
					}
				}
				// gateway.debug <- "conn sent"
			}
		}
	}
}

func generateUUIDMap() *sync.Map {
	var sm sync.Map
	return &sm
}
