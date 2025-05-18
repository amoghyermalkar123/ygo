package decoder

import (
	"encoding/json"
	"fmt"

	"github.com/amoghyermalkar123/ygo/internal/block"
)

func DecodeUpdate(update []byte) (*block.Updates, error) {
	// decode the binary `update` into a `DecodedUpdate` struct
	remoteUpdates := &block.Updates{}

	if err := json.Unmarshal(update, remoteUpdates); err != nil {
		return nil, fmt.Errorf("decode updates: %w", err)
	}

	return remoteUpdates, nil
}
