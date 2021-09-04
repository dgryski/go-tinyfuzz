// Package tinyfuzz is a small randomized testing package for tinygo
package tinyfuzz

import (
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

	var maxIterations = defaultConfig.Count
	if config.Count != 0 {
		maxIterations = config.Count
	}

	data := make([]byte, bufferLen)
	buf := make([]byte, bufferLen)

	for i := 0; i < maxIterations; i++ {
		l := bufferLen
		if config.Len == 0 {
			l = rand.Intn(bufferLen)
		}

		rand.Read(data[:l])
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
