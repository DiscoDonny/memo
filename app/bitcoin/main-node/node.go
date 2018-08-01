package main_node

import (
	"fmt"
	"github.com/jchavannes/btcd/peer"
	"github.com/jchavannes/btcd/wire"
	"github.com/memocash/memo/app/bitcoin/transaction"
	"github.com/memocash/memo/app/bitcoin/wallet"
	"github.com/memocash/memo/app/config"
	"github.com/memocash/memo/app/db"
	"log"
	"net"
)

var ActionNode Node
var UserNode Node

type Node struct {
	Peer               *peer.Peer
	NodeStatus         *db.NodeStatus
	UserNode           bool
	BlocksQueued       int
	HeaderSyncComplete bool
	BlocksSyncComplete bool
}

func StartActionNode() {
	ActionNode.Start(false)
}

func WaitForActionNodeDisconnect() {
	ActionNode.Peer.WaitForDisconnect()
}

func StartUserNode() {
	UserNode.Start(true)
}

func WaitForUserNodeDisconnect() {
	UserNode.Peer.WaitForDisconnect()
}

func (n *Node) Start(userNode bool) {
	nodeStatus, err := db.GetNodeStatus()
	if err != nil {
		log.Fatal(err)
	}
	n.UserNode = userNode
	transaction.EnableBatchPostProcessing()
	bitcoinNodeConfig := config.GetBitcoinNode()
	n.NodeStatus = nodeStatus
	p, err := peer.NewOutboundPeer(&peer.Config{
		UserAgentName:    "bch-lite-node",
		UserAgentVersion: "0.1.0",
		ChainParams:      &wallet.MainNetParams,
		Listeners: peer.MessageListeners{
			OnVerAck:  n.OnVerAck,
			OnHeaders: n.OnHeaders,
			OnInv:     n.OnInv,
			OnBlock:   n.OnBlock,
			OnTx:      n.OnTx,
			OnReject:  n.OnReject,
			OnPing:    n.OnPing,
		},
	}, bitcoinNodeConfig.GetConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	n.Peer = p
	fmt.Printf("Starting bitcoin node: %s\n", bitcoinNodeConfig.GetConnectionString())
	conn, err := net.Dial("tcp", bitcoinNodeConfig.GetConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	p.AssociateConnection(conn)
}

func (n *Node) OnVerAck(p *peer.Peer, msg *wire.MsgVerAck) {
	onVerAck(n, msg)
}

func (n *Node) OnHeaders(p *peer.Peer, msg *wire.MsgHeaders) {
	onHeaders(n, msg)
}

func (n *Node) OnInv(p *peer.Peer, msg *wire.MsgInv) {
	onInv(n, msg)
}

func (n *Node) OnTx(p *peer.Peer, msg *wire.MsgTx) {
	onTx(n, msg)
}

func (n *Node) OnBlock(p *peer.Peer, msg *wire.MsgBlock, buf []byte) {
	onBlock(n, msg)
}

func (n *Node) OnReject(p *peer.Peer, msg *wire.MsgReject) {
	onReject(n, msg)
}

func (n *Node) OnPing(p *peer.Peer, msg *wire.MsgPing) {
	n.Peer.QueueMessage(wire.NewMsgPong(msg.Nonce), nil)
}
