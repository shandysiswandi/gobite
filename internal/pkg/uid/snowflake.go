package uid

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/bwmarrin/snowflake"
)

// Snowflake generates numeric IDs using the Snowflake algorithm.
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

// NewSnowflake constructs a Snowflake generator with a random node ID.
func NewSnowflake() (*Snowflake, error) {
	nodeID, err := generateRandomNodeID()
	if err != nil {
		return nil, err
	}

	snowflake.Epoch = 1766854800000 // Mon Dec 28 2025 00:00:00.000 WIB

	node, err := snowflake.NewNode(nodeID)
	if err != nil {
		return nil, err
	}

	return &Snowflake{node: node}, nil
}

// Generate returns a new unique numeric ID.
func (s *Snowflake) Generate() int64 {
	return s.node.Generate().Int64()
}
