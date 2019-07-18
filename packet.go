package p2p

import (
	"bytes"
	"encoding/gob"
	"log"
	"net"

	// "net"
	uuid "github.com/nu7hatch/gouuid"
)

//Packet .
type Packet struct {
	Stype        bool
	UUID         string
	SAddr, RAddr net.UDPAddr
	Data         []string
}

//InitPacket .
func InitPacket() *Packet {
	return &Packet{}
}

//InitServicePacket ..
func InitServicePacket() *Packet {
	sp := InitPacket()
	sp.Stype = true
	sp.UUID = getUUID()
	return sp
}

//ToBytes ...
func (p *Packet) ToBytes() []byte {

	var byteBuffer bytes.Buffer
	enc := gob.NewEncoder(&byteBuffer)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal("packet encode error")
	}
	var packetData []byte
	packetData = byteBuffer.Bytes()
	// if err != nil {
	// 	log.Fatal("packet encode error")
	// }
	return packetData

}

//FromBytes ...
func (p *Packet) FromBytes(b []byte) error {

	var bytesBuffer = bytes.NewBuffer(b)
	dec := gob.NewDecoder(bytesBuffer)
	err := dec.Decode(&p)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	return err
}

func getUUID() string {
	u4, err := uuid.NewV4()
	if err != nil {
		log.Fatal("UUID generation error:", err)
	}
	return u4.String()
}

// func zverify(p Packet) bool {
// 	if len(p.data) < 64512 {
// 		return true
// 	}
// 	return false

// }

// //InitPackets ...
// func InitPackets(data []byte) []Packet {
// 	// var byteBuffer bytes.Buffer
// 	// enc := gob.NewEncoder(&byteBuffer)
// 	// err := enc.Encode(f)
// 	// if err != nil {
// 	// 	log.Fatal("encode error:",err)
// 	// }
// 	// var temp []byte
// 	// _,err = byteBuffer.Read(temp)
// 	// if err != nil {
// 	// 	log.Fatal("buffer read error:",err)
// 	// }
// 	UUID := getUUID()
// 	var chunk []byte
// 	packets := make([]Packet, len(data)/64513)
// 	for i := 1; len(data) >= 64512; i++ {
// 		chunk, data = data[:64512], data[64512:]
// 		packets = append(packets, Packet{i, UUID, chunk})
// 	}
// 	return packets
// }
