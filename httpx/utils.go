package httpx

import (
	"crypto/rand"
	"fmt"
)

// @see http://stackoverflow.com/a/25736155/750852
func uuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}
