package pkguid

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/bwmarrin/snowflake"
)

type Snowflake struct {
	node *snowflake.Node
}

func generateRandomNodeID() (int64, error) {
	var nodeID int64
	err := binary.Read(rand.Reader, binary.BigEndian, &nodeID)
	if err != nil {
		return 0, err
	}

	return nodeID & (1<<10 - 1), nil // Limiting to 10 bits for node ID
}

func NewSnowflake() (*Snowflake, error) {
	nodeID, err := generateRandomNodeID()
	if err != nil {
		return nil, err
	}

	snowflake.Epoch = 1764522000000 // Mon Dec 01 2025 00:00:00.000 WIB

	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}

	return &Snowflake{node: node}, nil
}

func (s *Snowflake) Generate() uint64 {
	return uint64(s.node.Generate().Int64())
}
