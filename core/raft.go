package core

const (
	RFOLLOWER = iota
	RCANDIDATE
	RLEADER
)

type RaftEntry struct {
	Term         uint64
	CallFunction func(any, any) error
	Args         any
	Reply        *any
}

type RaftState struct {
	// Persistent states on all servers
	Role            uint8
	CurrentTerm     uint64
	CurrentLeaderID uint8
	VotedFor        uint8
	LogEntries      []RaftEntry

	// Volatile states on all servers
	CommitIndex uint64
	LastApplied uint64

	// Volatile states on leaders
	PeerNextIndex  []uint64
	PeerMatchIndex []uint64
}

// // Leader replicate entries to followers
// func (rn *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
// 	return nil
// }

// Init: I am follower but leaderId is myself
func InitRaftNode(id uint8, N uint8) *RaftState {
	return &RaftState{
		Role:            RFOLLOWER,
		CurrentTerm:     0,
		CurrentLeaderID: id,
		VotedFor:        id,
		LogEntries:      make([]RaftEntry, 0),

		CommitIndex: 0,
		LastApplied: 0,

		PeerNextIndex:  make([]uint64, N),
		PeerMatchIndex: make([]uint64, N),
	}

}
