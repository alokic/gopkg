// Package hashid is implementation of http://www.hashids.org under MIT license
// Generates hashes from an array of integers, eg. for YouTube like hashes
// Uses: go get github.com/speps/go-hashids
package hashid

import (
	"github.com/alokic/gopkg/typeutils"
	"github.com/speps/go-hashids"
)

const (
	defaultSalt = "RnfK86AiEaSImrDBIucMEyCk0yBcv9uQ"
)

var (
	hashID *hashids.HashID
)

// Hash hashes a number or numeric string to a string containing at least MinLength characters taken from the Alphabet.
func Hash(id interface{}, salt ...string) string {
	h := initHashID(salt...)

	str, err := h.EncodeInt64([]int64{typeutils.ToInt64(id)})
	if err != nil {
		return ""
	}
	return str
}

// ID unhashes the string passed to an uint64 id.
func ID(hash interface{}, salt ...string) uint64 {
	h := initHashID(salt...)

	ids, err := h.DecodeInt64WithError(typeutils.ToStr(hash))
	if err != nil || (len(ids) == 0) {
		return uint64(0)
	}
	return uint64(ids[0])
}

func initHashID(salt ...string) *hashids.HashID {
	s := defaultSalt

	if len(salt) > 0 {
		s = salt[0]
	}
	if hashID == nil {
		hd := hashids.NewData()
		hd.Salt = s
		hd.MinLength = 8

		hashID, _ = hashids.NewWithData(hd)
	}

	return hashID
}
