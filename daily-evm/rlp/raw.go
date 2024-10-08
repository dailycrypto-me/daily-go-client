// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package rlp

import (
	"io"
	"reflect"

	"github.com/dailycrypto-me/daily-evm/daily/util"
)

// RawValue represents an encoded RLP value and can be used to delay
// RLP decoding or to precompute an encoding. Note that the decoder does
// not verify whether the content of RawValues is valid RLP.
type RawValue []byte

var rawValueType = reflect.TypeOf(RawValue{})

// Split returns the content of first RLP value and any
// bytes after the value as subslices of b.
func Split(b []byte) (k Kind, content, rest []byte, err error) {
	k, tagsize, total_size, err := ReadKind(b)
	if err != nil {
		return 0, nil, b, err
	}
	return k, b[tagsize:total_size], b[total_size:], nil
}

func MustSplit(b []byte) (k Kind, content, rest []byte) {
	var err error
	k, content, rest, err = Split(b)
	util.PanicIfNotNil(err)
	return
}

func MustSplitString(b []byte) (content, rest []byte) {
	var err error
	content, rest, err = SplitString(b)
	util.PanicIfNotNil(err)
	return
}

func MustSplitList(b []byte) (content, rest []byte) {
	var err error
	content, rest, err = SplitList(b)
	util.PanicIfNotNil(err)
	return
}

// SplitString splits b into the content of an RLP string
// and any remaining bytes after the string.
func SplitString(b []byte) (content, rest []byte, err error) {
	k, content, rest, err := Split(b)
	if err != nil {
		return nil, b, err
	}
	if k == List {
		return nil, b, ErrExpectedString
	}
	return content, rest, nil
}

// SplitList splits b into the content of a list and any remaining
// bytes after the list.
func SplitList(b []byte) (content, rest []byte, err error) {
	k, content, rest, err := Split(b)
	if err != nil {
		return nil, b, err
	}
	if k != List {
		return nil, b, ErrExpectedList
	}
	return content, rest, nil
}

// CountValues counts the number of encoded values in b.
func CountValues(b []byte) (int, error) {
	i := 0
	for ; len(b) > 0; i++ {
		_, _, total_size, err := ReadKind(b)
		if err != nil {
			return 0, err
		}
		b = b[total_size:]
	}
	return i, nil
}

func ReadKind(buf []byte) (k Kind, tagsize byte, total_size uint64, err error) {
	buflen := uint64(len(buf))
	if buflen == 0 {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}
	b := buf[0]
	var contentsize uint64
	switch {
	case b < 0x80:
		k = Byte
		tagsize = 0
		contentsize = 1
	case b < 0xB8:
		k = String
		tagsize = 1
		contentsize = uint64(b - 0x80)
		// Reject strings that should've been single bytes.
		if contentsize == 1 && buflen > 1 && buf[1] < 128 {
			return 0, 0, 0, ErrCanonSize
		}
	case b < 0xC0:
		k = String
		tagsize = b - 0xB7 + 1
		contentsize, err = readSize(buf[1:], b-0xB7)
	case b < 0xF8:
		k = List
		tagsize = 1
		contentsize = uint64(b - 0xC0)
	default:
		k = List
		tagsize = b - 0xF7 + 1
		contentsize, err = readSize(buf[1:], b-0xF7)
	}
	if err != nil {
		return 0, 0, 0, err
	}
	total_size = uint64(tagsize) + contentsize
	// Reject values larger than the input slice.
	if contentsize > buflen-uint64(tagsize) {
		return 0, 0, 0, ErrValueTooLarge
	}
	return
}

func readSize(b []byte, slen byte) (uint64, error) {
	if int(slen) > len(b) {
		return 0, io.ErrUnexpectedEOF
	}
	var s uint64
	switch slen {
	case 1:
		s = uint64(b[0])
	case 2:
		s = uint64(b[0])<<8 | uint64(b[1])
	case 3:
		s = uint64(b[0])<<16 | uint64(b[1])<<8 | uint64(b[2])
	case 4:
		s = uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3])
	case 5:
		s = uint64(b[0])<<32 | uint64(b[1])<<24 | uint64(b[2])<<16 | uint64(b[3])<<8 | uint64(b[4])
	case 6:
		s = uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])
	case 7:
		s = uint64(b[0])<<48 | uint64(b[1])<<40 | uint64(b[2])<<32 | uint64(b[3])<<24 | uint64(b[4])<<16 | uint64(b[5])<<8 | uint64(b[6])
	case 8:
		s = uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 | uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])
	default:
	}
	// Reject sizes < 56 (shouldn't have separate size) and sizes with
	// leading zero bytes.
	if s < 56 || b[0] == 0 {
		return 0, ErrCanonSize
	}
	return s, nil
}
