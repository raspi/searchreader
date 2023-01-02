package searchreader

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	// Source
	src := bytes.NewReader([]byte(`hello, world!`))

	// What to search
	search1 := strings.NewReader(`l`)
	search2 := strings.NewReader(`ll`)

	sr := New(src,
		WithCaseSensitive(search1),
		WithCaseSensitive(search2),
	)

	buffer := make([]byte, 1024)

	rb, results, err := sr.Read(buffer)

	if err != nil {
		if errors.Is(err, io.EOF) {
			t.Fail()
		}

		panic(err)
	}

	if rb != 13 {
		t.Errorf(`expected %d, got %d`, 13, rb)
		t.Fail()
	}

	if results == nil {
		t.Errorf(`expected results, got nil`)
		t.Fail()
	}

	if len(results) != 4 {
		// 3 x `l` 1 x `ll`
		t.Errorf(`expected 4 results, got %d`, len(results))
		t.Fail()
	}

	search1result1 := uint(2) // l
	search1result1found := false
	search1result2 := uint(3) // l
	search1result2found := false
	search1result3 := uint(10) // l
	search1result3found := false

	search2result1 := uint(2) // ll
	search2resultfound := false

	for _, res := range results {

		if res.Index == 0 && res.StartPosition == search1result1 {
			search1result1found = true
		}

		if res.Index == 0 && res.StartPosition == search1result2 {
			search1result2found = true
		}

		if res.Index == 0 && res.StartPosition == search1result3 {
			search1result3found = true
		}

		if res.Index == 1 && res.StartPosition == search2result1 {
			search2resultfound = true
		}

	}

	if !search1result1found {
		t.Errorf(`search1 result 1 (first l) not found`)
		t.Fail()
	}

	if !search1result2found {
		t.Errorf(`search1 result 2 (second l) not found`)
		t.Fail()
	}

	if !search1result3found {
		t.Errorf(`search1 result 3 (third l) not found`)
		t.Fail()
	}

	if !search2resultfound {
		t.Errorf(`search2 result (ll) not found`)
		t.Fail()
	}

}

// TestReaderChunk test chunk search
func TestReaderChunk(t *testing.T) {
	// Source
	src := bytes.NewReader([]byte(`xAAx`))

	// What to search
	search1 := strings.NewReader(`x`)
	search2 := strings.NewReader(`AA`)

	sr := New(src,
		WithCaseSensitive(search1),
		WithCaseSensitive(search2),
	)

	// Only read 2 bytes
	buffer := make([]byte, 2)

	rb, results1, err := sr.Read(buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			t.Fail()
		}

		panic(err)
	}

	if rb != 2 {
		t.Errorf(`expected %d, got %d`, 2, rb)
		t.Fail()
	}

	if len(results1) != 2 {
		t.Errorf(`expected %d results, got %d`, 2, len(results1))
		t.Fail()
	}

	search1result1 := uint(0) // x
	search1result1found := false
	search2result1 := uint(1) // AA
	search2result1found := false

	for _, res := range results1 {
		if res.Index == 0 && res.StartPosition == search1result1 {
			search1result1found = true
		}
		if res.Index == 1 && res.StartPosition == search2result1 {
			search2result1found = true
		}
	}

	if !search1result1found {
		t.Fail()
	}

	if !search2result1found {
		t.Fail()
	}

	rb, results2, err := sr.Read(buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			t.Fail()
		}

		panic(err)
	}

	if rb != 2 {
		t.Errorf(`expected %d, got %d`, 2, rb)
		t.Fail()
	}

	if len(results2) != 1 {
		t.Errorf(`expected %d results, got %d`, 1, len(results2))
		t.Fail()
	}

	search1result1 = uint(1) // x
	search1result1found = false

	for _, res := range results2 {
		if res.Index == 0 && res.StartPosition == search1result1 {
			search1result1found = true
		}
	}

	if !search1result1found {
		t.Fail()
	}

}
