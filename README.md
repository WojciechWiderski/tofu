# Tofu

The Golang framework allows you to quickly create applications.

```bash
go get -u github.com/WojciechWiderski/tofu
```

## Example

```go
package main

```


w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
fmt.Fprintln(w, "a\tb\tc\td\t")
fmt.Fprintln(w, "aa\tbb\tcc\t")
fmt.Fprintln(w, "aaa\tbbb\tccc\t")
fmt.Fprintln(w, "aaaa\tbbbb\tcccc\tdddd\t")
w.Flush()