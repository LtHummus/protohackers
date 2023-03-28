package speeding

import "io"

func ReadString(r io.Reader) (string, error) {
	lenByte := make([]byte, 1)
	_, err := r.Read(lenByte)
	if err != nil {
		return "", err
	}

	strBytes := make([]byte, lenByte[0])
	_, err = r.Read(strBytes)
	if err != nil {
		return "", err
	}

	return string(strBytes), nil
}

func WriteString(w io.Writer, s string) (int, error) {
	payload := make([]byte, 1)
	payload[0] = uint8(len(s))

	payload = append(payload, []byte(s)...)

	return w.Write(payload)
}
