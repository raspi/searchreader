package searchreader

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Stream interface {
	io.ReadSeeker
	io.RuneReader
	Size() int64
}

type Option func(this *findThis)

func WithCaseSensitive(source Stream) Option {
	if source == nil {
		panic(fmt.Errorf(`nil source`))
	}

	return func(r *findThis) {
		r.r = source
		r.caseSensitive = true
	}
}

func WithCaseInsensitive(source Stream) Option {
	if source == nil {
		panic(fmt.Errorf(`nil source`))
	}

	return func(r *findThis) {
		// Change all letters to lower case

		var sb = &strings.Builder{}
		sb.Reset()

		for {
			ch, size, err := source.ReadRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				panic(err)
			}

			if size == 0 {
				panic(`size 0??`)
			}

			if unicode.IsLetter(ch) {
				ch = unicode.ToLower(ch)
			}

			sb.WriteRune(ch)

		}

		r.r = strings.NewReader(sb.String())
		r.caseSensitive = false
	}
}
