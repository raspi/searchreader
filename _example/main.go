package main

import (
	"bytes"
	"fmt"
	"github.com/raspi/searchreader"
	"strings"
)

func main() {

	// Source
	src := bytes.NewReader(
		[]byte("heLLo, world!\000\000"),
	)

	// What to search
	search1 := strings.NewReader("\000")
	search2 := strings.NewReader(`ll`)

	sr := searchreader.New(src,
		searchreader.WithCaseSensitive(search1),
		searchreader.WithCaseInsensitive(search2),
	)

	buffer := make([]byte, 1024)

	_, results, err := sr.Read(buffer)

	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf(`found match at position %d that matches search%d %d`+"\n", result.StartPosition, 1+result.Index, result.Length)
	}

}
