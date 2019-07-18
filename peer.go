package p2p

import (
	"crypto/rsa"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// import (
//     "crypto"
//     "crypto/rand"
//     "crypto/rsa"
//     "crypto/sha256"
//     "fmt"
//     "os"
// )

//PeerInfo .
type PeerInfo struct {
	self                           *rsa.PrivateKey
	seldID                         string
	peers                          []Peer
	gateway                        *Gateway
	UUIDMap                        *sync.Map
	UserInput, GlobalOutput, Debug chan string
}

//Peer .
type Peer struct {
	name   string
	nodeID rsa.PublicKey
	status bool
}

func generatePeerInfo() *PeerInfo {
	return &PeerInfo{}
}

//InitPeerInfo .
func InitPeerInfo() *PeerInfo {
	peerInfo := generatePeerInfo()
	peerInfo.peers = []Peer{}
	prikey, pubkey := GenerateKeyPair()
	peerInfo.self = prikey
	selfid, err := ExportRsaPublicKey(&pubkey)
	if err != nil {
		log.Fatal(err)
	}
	peerInfo.seldID = selfid
	peerInfo.gateway = InitGatewayServer()
	peerInfo.UUIDMap = generateUUIDMap()
	peerInfo.UserInput = make(chan string, 10)
	peerInfo.GlobalOutput = make(chan string, 20)
	peerInfo.Debug = make(chan string, 10)
	peerInfo.Debug <- peerInfo.seldID
	peerInfo.Debug <- "peer initialized"
	go peerInfo.parseUserInput()
	go peerInfo.parseGatewayInput()
	return peerInfo
}

//parseUserInput .
func (pInfo *PeerInfo) parseUserInput() {
	for {
		uInput := <-pInfo.UserInput
		// pInfo.Debug <- "got inp"
		inp := strings.SplitN(uInput, ":", 2)
		if inp[0] == "add" {
			peer := strings.SplitN(uInput, ":", 2)
			pInfo.addPeer(peer[0], peer[1])
		} else if inp[0] == "send" {
			// pInfo.Debug <- "found send"
			pInfo.gateway.send <- pInfo.sendMessage(inp[1])
			// pInfo.Debug <- "sent to gateway"
		} else if inp[0] == "del" {
			pInfo.deletePeer(inp[1])
		} else if inp[0] == "chname" {
			peer := strings.SplitN(uInput, ":", 2)
			pInfo.changeName(peer[0], peer[1])
		}
	}
}

func (pInfo *PeerInfo) parseGatewayInput() {
	for {
		message := <-pInfo.gateway.recieve
		// pInfo.Debug <- "recieved something"
		output, err := pInfo.recieveMessage(message)
		if err == "" {
			pInfo.GlobalOutput <- output
			pInfo.gateway.send <- message
		} else if err == "notself" {
			pInfo.gateway.send <- message
		} else if err == "repeated" {
			continue
		}
	}
}

// func (pInfo *PeerInfo) sendGreeting() []string {
// 	data := []string{"", pInfo.seldID, "Greeting"}
// 	return data
// }

func (pInfo *PeerInfo) sendMessage(message string) []string {
	// pInfo.Debug <- "in sendmess"
	split := strings.SplitN(message, ":", 2)
	// pInfo.Debug <- "split second " + strconv.Itoa(len(split))
	for index, elem := range split {
		split[index] = strings.Trim(elem, " ")
	}
	peer := Peer{}
	if split[0] == "o" {
		return []string{"", pInfo.seldID, split[1]}
	}
	for _, tempPeer := range pInfo.peers {
		if tempPeer.name == split[0] {
			peer = tempPeer
			break
		}
	}

	nodeID, err := ExportRsaPublicKey(&peer.nodeID)
	if err != nil {
		log.Fatal(err)
	}
	// pInfo.Debug <- "waiting for ret"
	return []string{nodeID, pInfo.seldID, string(Encrypt(&(peer.nodeID), []byte(split[1])))}

}

func (pInfo *PeerInfo) recieveMessage(message []string) (string, string) {
	uuid := message[0]
	message = message[1:]
	_, repeated := pInfo.UUIDMap.LoadOrStore(uuid, message)
	if repeated {
		return "", "repeated"
	}
	// pInfo.Debug <- "1"
	peerID, err := ParseRsaPublicKey(message[1])
	if err != nil {
		time.After(100 * time.Millisecond)
		return "", "repeated"
	}
	if message[0] == "" {
		peer := Peer{}
		rand.Seed(time.Now().UnixNano())
		peer.name = "newguy" + strconv.Itoa(rand.Intn(100000))
		peer.nodeID = *peerID
		pInfo.peers = append(pInfo.peers, peer)
		return strings.Join([]string{" ", "open ->" + peer.name, message[2]}, ":"), ""
	}
	// pInfo.Debug <- "2"
	recieveID, err := ParseRsaPublicKey(message[0])
	if err != nil {
		return "", "repeated"
	}

	if *recieveID == pInfo.self.PublicKey {
		sentmessage := string(Decrypt(pInfo.self, []byte(message[2])))
		peer := Peer{}
		found := false
		for _, tempPeer := range pInfo.peers {
			if tempPeer.nodeID == *peerID {
				peer = tempPeer
				found = true
			}
		}
		if !found {
			rand.Seed(time.Now().UnixNano())
			peer.name = "newguy" + strconv.Itoa(rand.Intn(100000))
			peer.nodeID = *peerID
			pInfo.peers = append(pInfo.peers, peer)
		}
		return strings.Join([]string{peer.name, sentmessage}, " : "), ""
	}
	return "", "notself"

}

func (pInfo *PeerInfo) changeName(old, new string) {
	for index, peer := range pInfo.peers {
		if peer.name == old {
			peer.name = new
			pInfo.peers[index] = peer
		}
	}
}

func (pInfo *PeerInfo) deletePeer(name string) {
	for index, peer := range pInfo.peers {
		if peer.name == name {
			pInfo.peers[index] = pInfo.peers[len(pInfo.peers)-1]
		}
	}
	pInfo.peers = pInfo.peers[:len(pInfo.peers)-1]
}

func (pInfo *PeerInfo) addPeer(name string, node string) {

	peerID, err := ParseRsaPublicKey(node)
	if err != nil {
		log.Fatal(err)
	}
	pInfo.peers = append(pInfo.peers, Peer{name, *peerID, false})
}
