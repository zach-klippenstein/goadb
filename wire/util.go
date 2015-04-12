package wire

import "io"

// writeFully writes all of data to w.
// Inverse of io.ReadFully().
func writeFully(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
