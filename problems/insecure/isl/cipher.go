package isl

import (
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"math/bits"
)

const (
	cipherKindEnd         = byte(0)
	cipherKindReverseBits = byte(1)
	cipherKindXor         = byte(2)
	cipherKindXorPosition = byte(3)
	cipherKindAdd         = byte(4)
	cipherKindAddPos      = byte(5)
)

var (
	ErrUnknownCipherKind = errors.New("unknown cipher kind")
	ErrNoOpCipher        = errors.New("no-op cipher")
)

type cipher interface {
	Mutate(input []byte, pos uint64, invert bool)
}

type cipherChain struct {
	ciphers []cipher
}

func (cc *cipherChain) Encrypt(data []byte, pos uint64) {
	for _, curr := range cc.ciphers {
		curr.Mutate(data, pos, false)
	}
}

func (cc *cipherChain) Decrypt(data []byte, pos uint64) {
	for i := len(cc.ciphers) - 1; i >= 0; i-- {
		cc.ciphers[i].Mutate(data, pos, true)
	}
}

func buildCipherChain(r io.Reader) (*cipherChain, error) {
	var chain []cipher
	for {
		c, err := buildSingleCipher(r)
		if err != nil {
			return nil, err
		}
		if c == nil {
			// end of chain
			break
		}
		chain = append(chain, c)
	}

	cc := &cipherChain{
		ciphers: chain,
	}

	if cc.isNopCipherChain() {
		return nil, ErrNoOpCipher
	}

	return cc, nil
}

func (cc *cipherChain) isNopCipherChain() bool {
	test := make([]byte, 256)
	for i := range test {
		test[i] = byte(i)
	}

	cc.Encrypt(test, 0)
	for i := range test {
		if test[i] != byte(i) {
			return false
		}
	}

	return true
}

func buildSingleCipher(r io.Reader) (cipher, error) {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return nil, err
	}

	switch b[0] {
	case cipherKindEnd:
		log.Debug().Msg("end of cipher chain")
		return nil, nil
	case cipherKindReverseBits:
		return &reverseBits{}, nil
	case cipherKindXor:
		_, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		return &xorByte{k: b[0]}, nil
	case cipherKindXorPosition:
		return &xorPos{}, nil
	case cipherKindAdd:
		_, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		return &addByte{k: b[0]}, nil
	case cipherKindAddPos:
		return &addPos{}, nil
	default:
		return nil, ErrUnknownCipherKind
	}

}

type reverseBits struct{}

func (rb *reverseBits) Mutate(input []byte, pos uint64, reverse bool) {
	for i, c := range input {
		input[i] = bits.Reverse8(c)
	}
}

type xorByte struct {
	k byte
}

func (xb *xorByte) Mutate(input []byte, pos uint64, reverse bool) {
	for i, c := range input {
		input[i] = c ^ xb.k
	}
}

type xorPos struct{}

func (xp *xorPos) Mutate(input []byte, pos uint64, reverse bool) {
	for i, c := range input {
		input[i] = c ^ (byte(pos) + byte(i))
	}
}

type addByte struct {
	k byte
}

func (ab *addByte) Mutate(input []byte, pos uint64, reverse bool) {
	k := ab.k
	if reverse {
		k = ^k + 1 // 2s complement
	}
	for i, c := range input {
		input[i] = c + k
	}
}

type addPos struct{}

func (ap *addPos) Mutate(input []byte, pos uint64, reverse bool) {
	for i, c := range input {
		k := byte(pos) + byte(i)
		if reverse {
			k = ^k + 1
		}
		input[i] = c + k
	}
}
