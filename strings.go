package gostree

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"path"
	"runtime"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567789=!@#$%^&*()_+~`;:' ")

// RandomString returns a string of random characters with the specified
// length, Remember to call rand.Seed() first:
//
//     rand.Seed(time.Now().UnixNano())
//
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// PrettyJson returns the formatted, indented form of the specified json
func PrettyJson(j string) (string, error) {

	um := make(map[string]interface{})
	err := json.Unmarshal([]byte(j), &um)
	if err != nil {
		return "", errors.New(fmt.Sprint("PrettyJson error in Unmarshal: ", err))
	}

	p, err := json.MarshalIndent(um, "", "  ")
	if err != nil {
		return "", errors.New(fmt.Sprint("PrettyJson error in MarshalIndent: ", err))
	}

	return string(p), nil
}

// PackageDirectory returns the current absolute filesystem location
// of the package of its caller.
func PackageDirectory() (string, error) {
	var err error
	_, dir, _, ok := runtime.Caller(1)
	if !ok {
		err = fmt.Errorf("PackageDirectory Caller failed")
	}
	return path.Dir(dir), err
}
