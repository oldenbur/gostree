# gostree

[![Build Status](https://travis-ci.org/oldenbur/gostree.svg?branch=master)](https://travis-ci.org/oldenbur/gostree)
[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/oldenbur/gostree)
[![top level coverage](https://gocover.io/_badge/github.com/oldenbur/gostree?0 "top level coverage")](http://gocover.io/github.com/oldenbur/gostree)

gostree is a utility written in [Go](https://golang.org/) for navigating string-indexed trees of unmarshalled yaml and json. A primary use case is to handle json or yaml that does not have a pre-defined structure. For example, to ingest an arbitrary b-tree structure:
```go
yaml := `
---
L1:
  L2.1:
    L3.1.1: 
    L3.1.2:
  L2.2:
    L3.2.1: 
    L3.2.2: 
`

type bnode struct {
	data string
	l, r *bnode
}

func newBnode(data string, t interface{}) *bnode {
	n := &bnode{data: data}
	if t == nil {
		return n
	}
	keys, _ := t.(STree).KeyStrings()
	if len(keys) > 0 {
		lChild, _ := t.(STree).STreeVal(AsPath(keys[0]))
		n.l = newBnode(keys[0], lChild)
	}
	if len(keys) > 1 {
		rChild, _ := t.(STree).STreeVal(AsPath(keys[1]))
		n.r = newBnode(keys[1], rChild)
	}
	return n
}

t, _ := gostree.NewSTreeYaml(strings.NewReader(yaml))
keys, _ := t.KeyStrings()
n := newBnode(keys[0], t.STreeValMust(AsPath(keys[0])))
```

### Marshaling and Unmarshaling

An STree can be created from either json or yaml using one of the following:
```
    func NewSTreeJson(r io.Reader) (stree STree, err error)
    func NewSTreeYaml(r io.Reader) (stree STree, err error)
```
An STree can be marshaled to either json or yaml regardless of how it was constructed using one of the following:
```
    func (s STree) WriteJson(pretty bool) ([]byte, error)
    func (s STree) WriteYaml() ([]byte, error)
```

### Value Access and Key Syntax

Once created, an element anywhere within an STree can be accessed using a path which is a simplified version of the syntax used by the [jq](https://stedolan.github.io/jq/) tool. For example:
```go
s, _ := NewSTreeJson(strings.NewReader(`
{
  "key1": "val1",
  "key2": 1234,
  "key3": {
    "key4": true,
    "key5": -12.34,
    "key6": {
      "key7": [1, "data", {"key8": "val8"}]
    }
  }
}`))
v1 := s.StrValMust(`.key1`)                     // v1 is string "val1"
v2 := s.FloatValMust(`.key3.key5`)              // v2 is float64 -12.34
v3 := s.IntValMust(`.key3.key6.key7[0]`)        // v3 is int 1
v4 := s.STreeValMust(`.key3.key6.key7[2]`)      // v4 is STree {"key8": "val8"}
v5 := s.StrValMust(`.key3.key6.key7[2].key8`)   // v5 is string "val8"
```

### Traverse an STree with a Visitor

Clients can definite a visitor to easily traverse an STree, handling primitive, nested object and slice differently.
```go
s, _ := NewSTreeJson(strings.NewReader(`
{
  "key1": "val1",
  "key2": 1234,
  "key3": {
    "key4": true,
    "key5": -12.34,
    "key6": {
      "key7": [1, 2, 3]
    }
  }
}`))

s.Visit(NewVisitorBuilder().
  WithPrimitiveVisitor(func(key string, val interface{}) error {
    fmt.Printf("primitive - %s: %v", key, val)
    return nil
  }).
  WithSTreeBeginVisitor(func(key string, val STree) error {
    fmt.Printf("object beginning - %s", key)
    return nil
  }).
  WithSTreeEndVisitor(func(key string, val STree) error {
    fmt.Printf("object ending - %s", key)
    return nil
  }).
  WithSliceBeginVisitor(func(key string, val []interface{}) error {
    fmt.Printf("slice beginning - %s len: %d", key, len(val))
    return nil
  }).
  WithSliceEndVisitor(func(key string, val []interface{}) error {
    fmt.Printf("slice ending - %s len: %d", key, len(val))
    return nil
  }).
  Visitor(),
)

/*
prints:

primitive - .key1: val1
primitive - .key2: 1234
object beginning - .key3
primitive - .key3.key4: true
primitive - .key3.key5: -12.340000
object beginning - .key3.key6
slice beginning - .key3.key6.key7 len: 3
primitive - .key3.key6.key7[0]: 1
primitive - .key3.key6.key7[1]: 2
primitive - .key3.key6.key7[2]: 3
slice ending - .key3.key6.key7 len: 3
object ending - .key3.key6
object ending - .key3
*/
```

### Comparing STrees

Two STrees can be compared to one another. The values, value types and the structure of each STree is taken into account:
```go
s1, _ := NewSTreeJson(strings.NewReader(`
{
  "key1": "val1",
  "key2": 99,
  "key3": [4.32, true, "val2"]
}
`))
s2, _ := NewSTreeYaml(strings.NewReader(`
---
key1: val1
key2: 88
key3:
  - four_point_three_two
  - true
key4: val4
`))
diff, _ := s1.CompareTo(s2)

/*
diff is:
map[string]FieldComparisonResult {
".key1": COMP_NO_DIFFERENCE,
".key2": COMP_VALUES_DIFFER,
".key3[0]": COMP_TYPES_DIFFER,
".key3[1]": COMP_NO_DIFFERENCE,
".key3[2]": COMP_OBJECT_LACKS,
".key4": COMP_SUBJECT_LACKS,
}
*/
```

### TODO

* Support nested lists, e.g. .listVal[1][2]

