// Package tinyfuzz is a small randomized testing package for tinygo
package tinyfuzz

import (
	"encoding/binary"
	"fmt"
	"math/rand"

	"github.com/dgryski/go-ddmin"
)

// FuzzError is an error found by the fuzzer
type FuzzError struct {
	Input []byte
}

func (e *FuzzError) Error() string {
	return fmt.Sprintf("tinyfuzz: failing input: %v", e.Input)
}

type Config struct {
	Len   int // length of []byte to create
	Count int // iterations to fuzz
}

var defaultConfig = &Config{
	Len:   0,
	Count: 1000,
}

func Fuzz(f func([]byte) bool, config *Config) error {
	if config == nil {
		config = defaultConfig
	}

	var bufferLen = 2048
	if config.Len != 0 {
		bufferLen = config.Len
	}

	data := make([]byte, bufferLen)
	buf := make([]byte, bufferLen)

	for i := 0; i < config.Count; i++ {
		l := bufferLen
		if config.Len == 0 {
			l = rand.Intn(bufferLen)
		}

		fillBuffer(data, l)
		// data might be modified by f(), keep our own copy
		copy(buf[:l], data[:l])
		if ok := f(buf[:l]); !ok {
			// calling f() returned an error!
			if config.Len == 0 {
				data = ddmin.Minimize(data, func(d []byte) ddmin.Result {
					if !f(d) {
						return ddmin.Fail
					}

					return ddmin.Pass
				})
			}
			return &FuzzError{
				Input: data,
			}
		}
	}

	return nil
}

// fill data with l random bytes
func fillBuffer(data []byte, l int) {
	var i int
	for i = 0; i+8 < l; i += 8 {
		b := rand.Uint64()
		binary.LittleEndian.PutUint64(data[i:], b)
	}

	for b := rand.Uint64(); i < l; i++ {
		data[i] = byte(b)
		b >>= 8
	}
}
