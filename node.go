package p2p

import (
	"crypto/rsa"
)

//Node .
type Node struct {
	priKey *rsa.PrivateKey
	nodeID NodeID
}

//NodeID .
type NodeID rsa.PublicKey

//GenerateNode .
func GenerateNode() *Node {
	return &Node{}
}

//InitNode .
func InitNode() *Node {
	node := GenerateNode()
	priKey, pubKey := GenerateKeyPair()
	node.priKey = priKey
	node.nodeID = NodeID(pubKey)
	return node
}

//zInitNode  .
func zInitNode(priKey *rsa.PrivateKey) *Node {
	node := GenerateNode()
	pubKey := GeneratePublicKey(priKey)
	node.priKey = priKey
	node.nodeID = NodeID(pubKey)
	return node
}
