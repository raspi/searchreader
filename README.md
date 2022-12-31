# searchreader
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

## Is it any good?

Yes.
