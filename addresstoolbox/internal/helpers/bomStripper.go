package helpers

import (
	"fmt"
	"io"
)

// BOMType represents the type of BOM detected
type BOMType string

const (
	NoBOM      BOMType = "No BOM"
	UTF8BOM    BOMType = "UTF-8 BOM"
	UTF16BEBOM BOMType = "UTF-16 BE BOM"
	UTF16LEBOM BOMType = "UTF-16 LE BOM"
	UTF32BEBOM BOMType = "UTF-32 BE BOM"
	UTF32LEBOM BOMType = "UTF-32 LE BOM"
)

// BOMStrippingReader wraps an io.Reader to detect and optionally remove BOMs
type BOMStrippingReader struct {
	reader     io.Reader
	readBOM    bool
	removeBOM  bool
	bomType    BOMType
	bomBytes   int    // Number of bytes in the detected BOM
	extraBytes []byte // Buffered bytes read beyond BOM
}

// NewBOMStrippingReader creates a new BOMStrippingReader
func NewBOMStrippingReader(reader io.Reader, removeBOM bool) *BOMStrippingReader {
	return &BOMStrippingReader{
		reader:     reader,
		removeBOM:  removeBOM,
		bomType:    NoBOM,
		bomBytes:   0,
		extraBytes: nil,
	}
}

// Read implements io.Reader, detecting and optionally removing BOMs
func (r *BOMStrippingReader) Read(p []byte) (n int, err error) {
	if !r.readBOM {
		r.readBOM = true
		// Read up to 4 bytes to check for BOMs
		bom := make([]byte, 4)
		n, err := io.ReadAtLeast(r.reader, bom, 1) // Read at least 1 byte, up to 4
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return 0, fmt.Errorf("error reading BOM: %w", err)
		}

		// Check for BOMs
		if n >= 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
			r.bomType = UTF8BOM
			r.bomBytes = 3
			if r.removeBOM {
				// Buffer any extra bytes read beyond BOM
				if n > 3 {
					r.extraBytes = bom[3:n]
				}
				// If p is empty, return 0 to allow next read
				if len(p) == 0 {
					return 0, nil
				}
				// Return extra bytes or read from reader
				return r.readWithExtraBytes(p)
			}
			// Return BOM bytes
			copy(p, bom[:n])
			return n, err
		} else if n >= 2 && bom[0] == 0xFE && bom[1] == 0xFF {
			r.bomType = UTF16BEBOM
			r.bomBytes = 2
			if r.removeBOM {
				if n > 2 {
					r.extraBytes = bom[2:n]
				}
				if len(p) == 0 {
					return 0, nil
				}
				return r.readWithExtraBytes(p)
			}
			copy(p, bom[:n])
			return n, err
		} else if n >= 2 && bom[0] == 0xFF && bom[1] == 0xFE {
			r.bomType = UTF16LEBOM
			r.bomBytes = 2
			if n > 2 {
				r.extraBytes = bom[2:n]
			}
			if r.removeBOM {
				if len(p) == 0 {
					return 0, nil
				}
				return r.readWithExtraBytes(p)
			}
			copy(p, bom[:n])
			return n, err
		} else if n >= 4 && bom[0] == 0x00 && bom[1] == 0x00 && bom[2] == 0xFE && bom[3] == 0xFF {
			r.bomType = UTF32BEBOM
			r.bomBytes = 4
			if r.removeBOM {
				if len(p) == 0 {
					return 0, nil
				}
				return r.readWithExtraBytes(p)
			}
			copy(p, bom[:n])
			return n, err
		} else if n >= 4 && bom[0] == 0xFF && bom[1] == 0xFE && bom[2] == 0x00 && bom[3] == 0x00 {
			r.bomType = UTF32LEBOM
			r.bomBytes = 4
			if r.removeBOM {
				if len(p) == 0 {
					return 0, nil
				}
				return r.readWithExtraBytes(p)
			}
			copy(p, bom[:n])
			return n, err
		}
		// No BOM detected
		r.bomType = NoBOM
		r.bomBytes = 0
		r.extraBytes = bom[:n]
		if len(p) == 0 {
			return 0, nil
		}
		return r.readWithExtraBytes(p)
	}

	return r.readWithExtraBytes(p)
}

// readWithExtraBytes handles buffered extra bytes and subsequent reads
func (r *BOMStrippingReader) readWithExtraBytes(p []byte) (n int, err error) {
	if len(r.extraBytes) > 0 {
		// Return buffered extra bytes first
		n = copy(p, r.extraBytes)
		if n < len(r.extraBytes) {
			r.extraBytes = r.extraBytes[n:]
		} else {
			r.extraBytes = nil
		}
		return n, nil
	}
	// No extra bytes, read directly from underlying reader
	return r.reader.Read(p)
}

// GetBOMType returns the detected BOM type
func (r *BOMStrippingReader) GetBOMType() BOMType {
	return r.bomType
}

// GetBOMBytes returns the number of bytes in the detected BOM
func (r *BOMStrippingReader) GetBOMBytes() int {
	return r.bomBytes
}
