# searchreader
Search io.Reader 

```go
func main() {
	// Source
	src := bytes.NewReader([]byte(`hello, world!`))

	// What to search
	search1 := bytes.NewReader([]byte(`l`))
	search2 := bytes.NewReader([]byte(`ll`))

	var searchList []*bytes.Reader
	searchList = append(searchList, search1, search2)

	sr := searchreader.New(src, searchList)

	buffer := make([]byte, 1024)

	_, results, err := sr.Read(buffer)

	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf(`found match at position %d that matches search%d`+"\n", result.StartPosition, 1+result.Index)
	}

}
```

Outputs:

```
found match at position 2 that matches search1
found match at position 3 that matches search1
found match at position 10 that matches search1
found match at position 2 that matches search2
```
