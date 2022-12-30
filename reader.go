package searchreader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

// Result is search result, see SearcherReader.search()
type Result struct {
	Index         uint // Index of search []*bytes.Reader given in New
	StartPosition uint // StartPosition of first matched byte
	Length        uint // Length
}

type findThis struct {
	r         *bytes.Reader // Reader containing what bytes we are searching for
	firstByte byte          // Quick lookup for first matching byte
	size      uint64        // Size in bytes
}

type SearcherReader struct {
	src           *bufio.Reader     // Source io.Reader which is used for search source
	searchers     map[uint]findThis // What we are searching for
	requiredBytes uint64            // Minimum buffer size required
	buffer        []byte
	matches       []Result
}

func New(source io.Reader, searchers []*bytes.Reader) (sr *SearcherReader) {

	sr = &SearcherReader{
		requiredBytes: 0,
		src:           bufio.NewReader(source),
		searchers:     make(map[uint]findThis),
		matches:       []Result{},
	}

	if searchers != nil {
		for sIdx, s := range searchers {
			if s == nil {
				// should this error out?
				continue
			}

			if uint64(s.Size()) > sr.requiredBytes {
				// Update minimum required bytes required for searching the source SearcherReader.src reader
				sr.requiredBytes = uint64(s.Size())
			}

			// Get first byte as a possible search match hint so that lookup loops are a bit faster
			firstByte, err := s.ReadByte()
			if err != nil {
				panic(err)
			}

			// Rewind to start
			_, err = s.Seek(0, io.SeekStart)
			if err != nil {
				panic(err)
			}

			sr.searchers[uint(sIdx)] = findThis{
				r:         s,
				firstByte: firstByte,
				size:      uint64(s.Size()),
			}

		}
	}

	return sr
}

// readActual reads from the source io.Reader
func (sr *SearcherReader) readActual() (err error) {
	if uint64(len(sr.buffer)) > sr.requiredBytes*2 {
		// We have enough bytes in internal buffer for search, so we don't read more
		return nil
	}

	// We need double buffer space because the match could begin at the last Read() block byte
	// For example Read(buf) buffer length 5:
	//   find = `xyz`
	//   Internal buffer = [`a`, `b`, `c`, `d`, `x`, `y`, `z`, `0`, `1`, `2`]
	//   User gets [`a`, `b`, `c`, `d`, `x`] with Read() with result beginning at last index and
	//   internal buffer in the reader is left with  [`y`, `z`, `0`, `1`, `2`].
	//   User's next Read() gets the rest of the match

	buffer := make([]byte, sr.requiredBytes*2)
	readBytes, err := sr.src.Read(buffer)
	if err != nil {
		return err
	}

	sr.buffer = append(sr.buffer, buffer[0:readBytes]...)
	return nil
}

// search searches each sr.searchers bytes from internal buffer
func (sr *SearcherReader) search() {

	if sr.searchers == nil {
		return
	}

	if len(sr.searchers) == 0 {
		return
	}

	// key = searcher Index and value(s) are potential match start position(s)
	matches := make(map[uint][]uint64)

	for searcherIndex, searcher := range sr.searchers {
		// Search match starts from buffer and limit read size to sr.requiredBytes
		// this is so that what we are trying to find is contained in first half of the buffer
		for bIdx, bByte := range sr.buffer {
			if uint64(bIdx) >= sr.requiredBytes {
				break
			}

			if bByte == searcher.firstByte {
				// Potential match start
				matches[searcherIndex] = append(matches[searcherIndex], uint64(bIdx))
			}
		}
	}

	if matches == nil {
		// No matches
		return
	}

	if len(matches) == 0 {
		// No matches
		return
	}

	for matchIndex, startingPositions := range matches {
		for _, startingPosition := range startingPositions {

			searcher := sr.searchers[matchIndex]

			foundSize := uint64(0)

			// Rewind searcher to start
			_, err := searcher.r.Seek(0, io.SeekStart)
			if err != nil {
				panic(err)
			}

			// Search from where we found the first matching byte
			for _, bufferByte := range sr.buffer[uint(startingPosition):] {
				findByte, err := searcher.r.ReadByte()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					panic(err)
				}

				if findByte == bufferByte {
					foundSize++
				}

			}

			if foundSize == searcher.size {
				// match found
				sr.matches = append(sr.matches, Result{
					Index:         matchIndex,
					StartPosition: uint(startingPosition),
					Length:        uint(searcher.size),
				})

			}

		}
	}

}

// Read reads internal buffer and returns bytes from the internal buffer's start and possible matches
func (sr *SearcherReader) Read(b []byte) (readBytes int, results []Result, err error) {
	if uint64(len(b)) > sr.requiredBytes {
		sr.requiredBytes = uint64(len(b))
	}

	// Read to internal buffer
	err = sr.readActual()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, nil, err
		}

		panic(err)
	}

	if len(sr.searchers) > 0 {
		// We have searchers
		sr.search()
	}

	tmp := bytes.NewReader(sr.buffer)
	readBytes, err = tmp.Read(b)
	return readBytes, sr.matches, err
}
