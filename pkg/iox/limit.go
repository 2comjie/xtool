package iox

import "io"

func ReadFull(r io.Reader, bs []byte) error {
	_, err := io.ReadFull(r, bs)
	return err
}

func WriteFull(w io.Writer, bs []byte, limit int) error {
	for len(bs) > 0 {
		chunk := bs
		if limit > 0 && len(chunk) > limit {
			chunk = bs[:limit]
		}
		n, err := w.Write(chunk)
		if err != nil {
			return err
		}
		bs = bs[n:]
	}
	return nil
}
