/*
 * Copyright (c) 2019 wellwell.work, LLC by Zoe
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package x

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Code are comes from google/uuid
// Listen with it.

// A UUID is a 128 bit (16 byte) Universal Unique IDentifier as defined in RFC
// 4122.
type UUID [16]byte

// Well known namespace IDs and UUIDs
var (
	rander = rand.Reader // random function
	Nil           UUID // empty UUID, all zeros
)

// String returns the string form of uuid, xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// , or "" if uuid is invalid.
func (uuid UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return Bytes2Str(buf[:])
}

// URN returns the RFC 2141 URN form of uuid,
// urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx,  or "" if uuid is invalid.
func (uuid UUID) URN() string {
	var buf [36 + 9]byte
	copy(buf[:], "urn:uuid:")
	encodeHex(buf[9:], uuid)
	return Bytes2Str(buf[:])
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}

// MustUUID returns uuid if err is nil and panics otherwise.
func MustUUID(uuid UUID, err error) UUID {
	if err != nil {
		panic(err)
	}
	return uuid
}

// NewUUID creates a new random UUID or panics.  New is equivalent to
// the expression
//
//    uuid.Must(uuid.NewRandom())
func NewUUID() UUID {
	return MustUUID(NewRandomUUID())
}

// NewRandomUUID returns a Random (Version 4) UUID.
//
// The strength of the UUIDs is based on the strength of the crypto/rand
// package.
//
// A note about uniqueness derived from the UUID Wikipedia entry:
//
//  Randomly generated UUIDs have 122 random bits.  One's annual risk of being
//  hit by a meteorite is estimated to be one chance in 17 billion, that
//  means the probability is about 0.00000000006 (6 × 10−11),
//  equivalent to the odds of creating a few tens of trillions of UUIDs in a
//  year and having one duplicate.
func NewRandomUUID() (UUID, error) {
	return NewRandomUUIDFromReader(rander)
}

// NewRandomUUIDFromReader returns a Random (Version 4) UUID.
func NewRandomUUIDFromReader(r io.Reader) (UUID, error) {
	var uuid UUID
	_, err := io.ReadFull(r, uuid[:])
	if err != nil {
		return Nil, err
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid, nil
}

// ParseUUID decodes s into a UUID or returns an error.  Both the standard UUID
// forms of xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx and
// urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx are decoded as well as the
// Microsoft encoding {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx} and the raw hex
// encoding: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.
func ParseUUID(s string) (UUID, error) {
	var uuid UUID
	switch len(s) {
	// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case 36:

	// urn:uuid:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	case 36 + 9:
		if strings.ToLower(s[:9]) != "urn:uuid:" {
			return uuid, fmt.Errorf("invalid urn prefix: %q", s[:9])
		}
		s = s[9:]

	// {xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}
	case 36 + 2:
		s = s[1:]

	// xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
	case 32:
		var ok bool
		for i := range uuid {
			uuid[i], ok = xtob(s[i*2], s[i*2+1])
			if !ok {
				return uuid, errors.New("invalid UUID format")
			}
		}
		return uuid, nil
	default:
		return uuid, fmt.Errorf("invalid UUID length: %d", len(s))
	}
	// s is now at least 36 bytes long
	// it must be of the form  xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return uuid, errors.New("invalid UUID format")
	}
	for i, x := range [16]int{
		0, 2, 4, 6,
		9, 11,
		14, 16,
		19, 21,
		24, 26, 28, 30, 32, 34} {
		v, ok := xtob(s[x], s[x+1])
		if !ok {
			return uuid, errors.New("invalid UUID format")
		}
		uuid[i] = v
	}
	return uuid, nil
}

// randomBits completely fills slice b with random data.
func randomBits(b []byte) {
	if _, err := io.ReadFull(rander, b); err != nil {
		panic(err.Error()) // rand should never fail
	}
}

// xvalues returns the value of a byte as a hexadecimal digit or 255.
var xvalues = [256]byte{
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 255, 255, 255, 255, 255, 255,
	255, 10, 11, 12, 13, 14, 15, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 10, 11, 12, 13, 14, 15, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
}

// xtob converts hex characters x1 and x2 into a byte.
func xtob(x1, x2 byte) (byte, bool) {
	b1 := xvalues[x1]
	b2 := xvalues[x2]
	return (b1 << 4) | b2, b1 != 255 && b2 != 255
}