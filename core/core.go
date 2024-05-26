package core

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"time"

	"nis3607/mylogger"
	"nis3607/myrpc"
)

// Represents a consensus node.
// Run() will start the node.
type Consensus struct {
	id   uint8  // node id
	n    uint8  // total number of nodes
	port uint64 // rpc port
	seq  uint64 // commit log-array seq number
	//BlockChain
	blockChain *BlockChain // block chain that runs over consensus
	//logger
	logger *mylogger.MyLogger // logger with debug flag
	//rpc network
	peers []*myrpc.ClientEnd // rpc client, including port and rpc.Client which is connected and ready to Call

	//message
	commitedEntries chan *myrpc.ConsensusMsg

	raftState *RaftState
}

// init Consensus node with node `config`, `BlockChain`
// Also records peers(rpc clients), and a message channel used for rpc communication
// rpc server started at init
func InitConsensus(config *Configuration) *Consensus {
	rand.Seed(time.Now().UnixNano())
	c := &Consensus{
		id:         config.Id,
		n:          config.N,
		port:       config.Port,
		seq:        0,
		blockChain: InitBlockChain(config.Id, config.BlockSize),
		logger:     mylogger.InitLogger("node", config.Id),
		peers:      make([]*myrpc.ClientEnd, 0),

		commitedEntries: make(chan *myrpc.ConsensusMsg, 1024),
		raftState:       InitRaftNode(config.Id, config.N),
	}
	for _, peer := range config.Committee {
		clientEnd := &myrpc.ClientEnd{Port: uint64(peer)}
		c.peers = append(c.peers, clientEnd)
	}
	go c.serve()

	return c
}

// start a goroutine to serve rpc.
// This function uses default http rpc settings.
func (c *Consensus) serve() {
	rpc.Register(c)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(int(c.port)))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (c *Consensus) OnReceiveMessage(args *myrpc.ConsensusMsg, reply *myrpc.ConsensusMsgReply) error {

	c.logger.DPrintf("Invoke RpcExample: receive message from %v at %v", args.From, time.Now().Nanosecond())
	c.commitedEntries <- args
	return nil
}

func (c *Consensus) requestLeader(msg *myrpc.ConsensusMsg) {
	leaderId := c.raftState.CurrentLeaderID
	c.peers[leaderId].Call("Consensus.")
}

// This is a rpc server function that handles received message.
func (c *Consensus) handleMsgExample(msg *myrpc.ConsensusMsg) {
	block := &Block{
		Seq:  msg.Seq,
		Data: msg.Data,
	}
	c.blockChain.commitBlock(block)
}

// 模拟节点中的一个挖矿进程，使用共识服务
func (c *Consensus) proposeLoop() {
	for {
		// generate a new block. The seq is fake
		block := c.blockChain.getBlock(c.seq)
		msg := &myrpc.ConsensusMsg{
			From: c.id,
			Seq:  block.Seq,
			Data: block.Data,
		}
		c.requestLeader(msg)
	}

}

func (c *Consensus) Run() {
	// wait for other node to start
	time.Sleep(time.Duration(1) * time.Second)
	//init rpc client
	for id := range c.peers {
		c.peers[id].Connect()
	}

	go c.proposeLoop()
	//handle received message
	for {
		msg := <-c.commitedEntries
		c.blockChain.commitBlock(&Block{
			Seq:  msg.Seq,
			Data: msg.Data})
	}
}
