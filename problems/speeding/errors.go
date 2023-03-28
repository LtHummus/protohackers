package speeding

import "errors"

var (
	InvalidMsgPayload = []byte{0x10, 0x0b, 0x69, 0x6c, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x20, 0x6d, 0x73, 0x67}

	HeartbeatPayload = []byte{MessageKindHeartbeat}
)

var (
	ErrNotACamera      = errors.New("not a camera")
	ErrAlreadyAssigned = errors.New("already claimed a role")
	ErrCouldNotRead    = errors.New("invalid payloa")
)
