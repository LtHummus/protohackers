package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type DialAuthority struct {
	Site uint32
}

func (d *DialAuthority) String() string {
	return fmt.Sprintf("DialAuthority{Site: %d}", d.Site)
}

func (d *DialAuthority) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}
	binary.Write(&buf, binary.BigEndian, d)

	return wrapPayload(MagicNumberDialAuthority, buf.Bytes()), nil
}

func deserializeDialAuthority(r io.Reader) (Message, error) {
	var site uint32
	err := binary.Read(r, binary.BigEndian, &site)
	if err != nil {
		return nil, err
	}

	return &DialAuthority{
		Site: site,
	}, nil

}
