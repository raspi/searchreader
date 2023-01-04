# searchreader

![GitHub All Releases](https://img.shields.io/github/downloads/raspi/searchreader/total?style=for-the-badge)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/raspi/searchreader?style=for-the-badge)
![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/raspi/searchreader?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/raspi/searchreader)](https://goreportcard.com/report/github.com/raspi/searchreader)

Search single `bytes.Reader` with different `strings.Reader`(s) containing the search byte(s) or string(s).

Part of [heksa issue](https://github.com/raspi/heksa/issues/8).

```go
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
```

Outputs:

```
found match at position 2 that matches search2 2
found match at position 13 that matches search1 1
found match at position 14 that matches search1 1
```

## Some goals

* Convert searched string(s) into different encodings such as ISO-8859-1 and ISO-8859-15 so that you can search more efficiently at once

## Is it any good?

Yes.
