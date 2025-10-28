// Package sid provides the ability to generate Sortable Identifiers. These are also universally unique.
//
//     id1 := sid.Id()
//     id2 := sid.Id()
//     fmt.Printf("id1 = %s\n", id1)
//     fmt.Printf("id2 = %s\n", id2)
//     // -> "id1 = 1IeSBAWW83j-2wgJ4PUtlAr"
//     // -> "id2 = 1IeSBAWW9kK-0cDG64GQgGJ"
//
// This package is simple and only provides one function. The aim here is not pure speed, it is for an easy use-case
// without having to worry about goroutines and locking.
package sid

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Remember the lastTime so that if (by chance) we get the same NanoSecond, we just incrememt the last random number.
var mu = &sync.Mutex{}
var lastTime int64
var lastRand int64
var chars = make([]string, 11, 11)

// 64 chars but ordered by ASCII
const base64 string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~"

func toStr(n int64) string {
	// now do the generation (backwards, so we just %64 then /64 along the way)
	for i := 10; i >= 0; i-- {
		index := n % 64
		chars[i] = string(base64[index])
		n = n / 64
	}

	return strings.Join(chars, "")
}

func toBase32(n int64) string {
	b32 := strconv.FormatInt(n, 32)

	for len(b32) < 13 {
		b32 = "0" + b32
	}

	// log.Printf("b32=%s\n", b32)

	return b32
}

func toHex(n int64) string {
	hex := fmt.Sprintf("%x", n)

	for len(hex) < 16 {
		hex = "0" + hex
	}

	return hex
}

// IdBase64 returns a 23 char string based on timestamp and a random number. The format is "XXXXXXXXXXX-YYYYYYYYYYY"
// where X is the timestamp and Y is the random number. If (by any chance) this is called in the same nanosecond, the
// random number is incremented instead of a new one being generated. This makes sure that two consecutive Ids
// generated in the same goroutine also ensure those Ids are also sortable.
//
// It is safe to call from different goroutines since it has it's own locking.
func IdBase64() string {
	// lock for lastTime, lastRand, and chars
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UTC().UnixNano()
	var r int64

	// if we have the same time, just inc lastRand, else create a new one
	if now == lastTime {
		lastRand++
	} else {
		lastRand = rand.Int63()
	}
	r = lastRand

	// remember this for next time
	lastTime = now

	return toStr(now) + "-" + toStr(r)
}

// IdBase32 returns a 27 char string based on timestamp and a random number. The format is
// "XXXXXXXXXXXXX-YYYYYYYYYYYYY" where X is the timestamp and Y is the random number. If (by any chance) this is called
// in the same nanosecond, the random number is incremented instead of a new one being generated. This makes sure that
// two consecutive Ids generated in the same goroutine also ensure those Ids are also sortable.
//
// It is safe to call from different goroutines since it has it's own locking.
func IdBase32() string {
	// lock for lastTime, lastRand, and chars
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UTC().UnixNano()
	var r int64

	// if we have the same time, just inc lastRand, else create a new one
	if now == lastTime {
		lastRand++
	} else {
		lastRand = rand.Int63()
	}
	r = lastRand

	// remember this for next time
	lastTime = now

	return toBase32(now) + "-" + toBase32(r)
}

// IdHex returns a char string based on timestamp and a random number. The format is
// "XXXXXXXXXXXXXXXX-YYYYYYYYYYYYYYYY" where X is the timestamp and Y is the random number. If (by any chance) this is
// called in the same nanosecond, the random number is incremented instead of a new one being generated. This makes
// sure that two consecutive Ids generated in the same goroutine also ensure those Ids are also sortable.
//
// It is safe to call from different goroutines since it has it's own locking.
func IdHex() string {
	// lock for lastTime, lastRand, and chars
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UTC().UnixNano()
	var r int64

	// if we have the same time, just inc lastRand, else create a new one
	if now == lastTime {
		lastRand++
	} else {
		lastRand = rand.Int63()
	}
	r = lastRand

	// remember this for next time
	lastTime = now

	return toHex(now) + "-" + toHex(r)
}

// Id returns a 39 char string based on timestamp and a random number. The format is
// "XXXXXXXXXXXXXXXXXXX-YYYYYYYYYYYYYYYYYYY" where X is the timestamp and Y is the random number. If (by any chance)
// this is called in the same nanosecond, the random number is incremented instead of a new one being generated. This
// makes sure that two consecutive Ids generated in the same goroutine also ensure those Ids are also sortable.
//
// It is safe to call from different goroutines since it has it's own locking.
func Id() string {
	// lock for lastTime, lastRand, and chars
	mu.Lock()
	defer mu.Unlock()

	now := time.Now().UTC().UnixNano()
	var r int64

	// if we have the same time, just inc lastRand, else create a new one
	if now == lastTime {
		lastRand++
	} else {
		lastRand = rand.Int63()
	}
	r = lastRand

	// remember this for next time
	lastTime = now

	nowStr := strconv.FormatInt(now, 10)
	rStr := strconv.FormatInt(r, 10)

	for len(rStr) < 19 {
		rStr = "0" + rStr
	}

	return nowStr + "-" + rStr
}
