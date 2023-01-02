package searchreader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Match is search result, see SearcherReader.search().
type Match struct {
	Index         uint // Index of search []*bytes.Reader given in New (see: Option)
	StartPosition uint // StartPosition of first matched byte (relative position of Read())
	Length        uint // Length in bytes
}

type findThis struct {
	r             Stream // Reader containing what bytes we are searching for
	firstRune     rune   // Quick lookup for first matching rune
	size          uint64 // Size in bytes
	caseSensitive bool   // is search case-sensitive?
}

type SearcherReader struct {
	src           *bufio.Reader     // Source io.Reader which is used for search source
	searchers     map[uint]findThis // What we are searching for
	requiredBytes uint64            // Minimum buffer size required
	buffer        []byte            // Internal source buffer for the SearcherReader.searchers
}

func New(source io.Reader, opts ...Option) (sr *SearcherReader) {

	sr = &SearcherReader{
		requiredBytes: 0,
		src:           bufio.NewReader(source),
		searchers:     make(map[uint]findThis),
	}

	if opts != nil {
		for sIdx, opt := range opts {
			if opt == nil {
				// should this error out?
				continue
			}

			// Default
			def := &findThis{
				r:             nil,
				firstRune:     0,
				size:          0,
				caseSensitive: false,
			}

			// Apply option
			opt(def)

			if def.r == nil {
				panic(fmt.Errorf(`nil searcher #%d`, sIdx))
			}

			def.size = uint64(def.r.Size())

			if def.size > sr.requiredBytes {
				// Update minimum required bytes required for searching the source SearcherReader.src reader
				sr.requiredBytes = def.size
			}

			// Get first byte as a possible search match hint so that lookup loops are a bit faster
			firstByte, _, err := def.r.ReadRune()
			if err != nil {
				panic(err)
			}
			def.firstRune = firstByte

			// Rewind to start
			_, err = def.r.Seek(0, io.SeekStart)
			if err != nil {
				panic(err)
			}

			sr.searchers[uint(sIdx)] = *def

		}
	}

	return sr
}

// readActual reads from the source io.Reader to a internal buffer
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
		if errors.Is(err, io.EOF) && len(sr.buffer) > 0 {
			// We have data still in the sr.buffer
			err = nil
		}
	}

	// Add read data from source stream to internal buffer
	if readBytes > 0 {
		sr.buffer = append(sr.buffer, buffer[0:readBytes]...)
	}

	return err
}

// search searches each sr.searchers bytes from internal buffer
func (sr *SearcherReader) search() (matches []Match) {

	if sr.searchers == nil {
		// No searchers, skip
		return
	}

	if len(sr.searchers) == 0 {
		// No searchers, skip
		return
	}

	// key = searcher Index and value(s) are potential match start position(s)
	potentialMatches := make(map[uint][]uint64)

	// Change raw bytes internal buffer to string so that case-sensitive search can be made
	sourceBuffer := strings.NewReader(string(sr.buffer))

	for searcherIndex, searcher := range sr.searchers {

		// Rewind to start
		_, _ = sourceBuffer.Seek(0, io.SeekStart)

		for {
			offset, err := sourceBuffer.Seek(0, io.SeekCurrent)
			if err != nil {
				panic(err)
			}

			if uint64(offset) > sr.requiredBytes {
				// Do not search beyond sr.requiredBytes limit
				break
			}

			srcRune, srcSize, err := sourceBuffer.ReadRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				panic(err)
			}

			if srcSize == 0 {
				panic(`size 0??`)
			}

			if unicode.IsLetter(srcRune) && !searcher.caseSensitive {
				// searcher has all letters in lower case if it's case-insensitive.
				// See: Option.WithCaseInsensitive
				srcRune = unicode.ToLower(srcRune)
			}

			if srcRune == searcher.firstRune {
				// Add potential start for match start
				potentialMatches[searcherIndex] = append(potentialMatches[searcherIndex], uint64(offset))
			}
		}

	}

	if potentialMatches == nil {
		// No matches
		return
	}

	if len(potentialMatches) == 0 {
		// No matches
		return
	}

	// Search the potential matches
	for searcherIndex, startingPositions := range potentialMatches {
		for _, startingPosition := range startingPositions {

			// Rewind source buffer to starting position of potential match
			_, err := sourceBuffer.Seek(int64(startingPosition), io.SeekStart)
			if err != nil {
				panic(err)
			}

			searcher, ok := sr.searchers[searcherIndex]
			if !ok {
				panic(`invalid searcher index`)
			}
			foundSize := uint64(0)

			// Rewind searcher to start
			_, err = searcher.r.Seek(0, io.SeekStart)
			if err != nil {
				panic(err)
			}

			// Search rune-by-rune
			for {
				srcRune, srcSize, err := sourceBuffer.ReadRune()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					panic(err)
				}

				if srcSize == 0 {
					panic(`size 0??`)
				}

				if unicode.IsLetter(srcRune) && !searcher.caseSensitive {
					// searcher has all letters in lower case if it's case-insensitive.
					// See: Option.WithCaseInsensitive
					srcRune = unicode.ToLower(srcRune)
				}

				findRune, chsize, err := searcher.r.ReadRune()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					panic(err)
				}

				if chsize == 0 {
					panic(`size 0??`)
				}

				if srcSize == chsize && findRune == srcRune {
					foundSize++
				} else {
					// No match
					foundSize = 0
					break
				}

			}

			if foundSize == searcher.size {
				// match found
				matches = append(matches, Match{
					Index:         searcherIndex,
					StartPosition: uint(startingPosition),
					Length:        uint(searcher.size),
				})

			}

		}
	}

	return matches
}

// Read reads internal buffer and returns bytes from the internal buffer's start and possible matches.
// End user has to keep track of matching Match data lengths search window.
func (sr *SearcherReader) Read(b []byte) (readBytes int, matches []Match, err error) {
	if uint64(len(b)) > sr.requiredBytes {
		// Update sr.requiredBytes if it's larger than search string(s)
		sr.requiredBytes = uint64(len(b))
	}

	// Read to internal buffer
	err = sr.readActual()
	if err != nil {
		return 0, matches, err
	}

	if len(sr.searchers) > 0 {
		// We have searchers, do search
		matches = sr.search()
	}

	tmp := bytes.NewReader(sr.buffer)
	readBytes, err = tmp.Read(b)

	// remove already read data from buffer
	sr.buffer = sr.buffer[readBytes:]

	return readBytes, matches, err
}
