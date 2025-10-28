package utf16le

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

var le = binary.LittleEndian

// Character mapping strategy that can be used when a reserved character is
// encountered in a file name.
type MapChars int

const (
	// Don't map reserved characters
	MapCharsNone MapChars = 0
	// Map reserved characters using the Services for Mac scheme. This is
	// equivalent to using the 'mapposix' when mounting a volume in Linux.
	MapCharsSFM MapChars = 1
	// Map reserved characters using the Services for Unix scheme. This is
	// equivalent to using 'mapchars' when mounting a volume in Linux.
	MapCharsSFU MapChars = 2
)

const (
	SFMDoubleQuote uint16 = 0xF020
	SFMAsterisk    uint16 = 0xF021
	SFMColon       uint16 = 0xF022
	SFMLessThan    uint16 = 0xF023
	SFMGreaterThan uint16 = 0xF024
	SFMQuestion    uint16 = 0xF025
	SFMSlash       uint16 = 0xF026
	SFMPipe        uint16 = 0xF027
	SFMSpace       uint16 = 0xF028
	SFMPeriod      uint16 = 0xF029
)

const (
	SFUAsterisk    uint16 = '*' + 0xF000
	SFUQuestion    uint16 = '?' + 0xF000
	SFUColon       uint16 = ':' + 0xF000
	SFUGreaterThan uint16 = '>' + 0xF000
	SFULessThan    uint16 = '<' + 0xF000
	SFUPipe        uint16 = '|' + 0xF000
	SFUSlash       uint16 = '\\' + 0xF000
)

const unicodeBackSlash = '\\'

func Decode(src []byte, mc MapChars) string {
	if len(src) == 0 {
		return ""
	}

	u16s := make([]uint16, len(src)/2)
	for i := range u16s {
		u16 := le.Uint16(src[2*i : 2*i+2])

		switch mc {
		case MapCharsSFM:
			if u16 >= 0xF001 && u16 <= 0xF01F {
				u16 -= 0xF000
			} else {
				switch u16 {
				case SFMDoubleQuote:
					u16 = '"'
				case SFMAsterisk:
					u16 = '*'
				case SFMColon:
					u16 = ':'
				case SFMLessThan:
					u16 = '<'
				case SFMGreaterThan:
					u16 = '>'
				case SFMQuestion:
					u16 = '?'
				case SFMPipe:
					u16 = '|'
				case SFMSpace:
					u16 = ' '
				case SFMPeriod:
					u16 = '.'
				}
			}
		case MapCharsSFU:
			switch u16 {
			case SFUColon:
				u16 = ':'
			case SFUAsterisk:
				u16 = '*'
			case SFUQuestion:
				u16 = '?'
			case SFUPipe:
				u16 = '|'
			case SFUGreaterThan:
				u16 = '>'
			case SFULessThan:
				u16 = '<'
			}
		}

		u16s[i] = u16
	}

	// Remove the null terminator
	if len(u16s) > 0 && u16s[len(u16s)-1] == 0 {
		u16s = u16s[:len(u16s)-1]
	}

	return string(utf16.Decode(u16s))
}

func Encode(src string, mc MapChars) []byte {
	if len(src) == 0 {
		return nil
	}

	u16s := pathToU16s(src, mc)
	bs := make([]byte, len(u16s)*2)
	for i, u16 := range u16s {
		le.PutUint16(bs[2*i:2*i+2], u16)
	}
	return bs
}

func EncodedStringLen(s string) int {
	l := 0
	for _, r := range s {
		if 0x10000 <= r && r <= '\U0010FFFF' {
			l += 4
		} else {
			l += 2
		}
	}
	return l
}

func EncodeSlice(dst []byte, src string, mc MapChars) int {
	if len(src) == 0 {
		return 0
	}

	u16s := pathToU16s(src, mc)

	for i, u16 := range u16s {
		le.PutUint16(dst[2*i:2*i+2], u16)
	}

	return len(u16s) * 2
}

// Takes a single path component (e.g. the 'mydir' or 'foo.txt' in 'mydir\foo.txt'), pathComp, converts it into utf-16, and performs character mapping based on mc. This function should not be used directly, but rather as a part of pathToU16s().
func mappedU16s(pathComp string, mc MapChars) []uint16 {
	u16s := utf16.Encode([]rune(pathComp))
	switch mc {
	case MapCharsSFM:
		for i, u16 := range u16s {
			if sfm := toSFM(u16, i == len(u16s)-1); sfm != 0 {
				// a mapping needs to be performed
				u16s[i] = sfm
			}
		}
	case MapCharsSFU:
		for i, u16 := range u16s {
			if sfu := toSFU(u16); sfu != 0 {
				u16s[i] = sfu
			}
		}
	}

	return u16s
}

// Accepts a path (e.g. 'mydir\foo.txt'), path, breaks it up into components, converts it into utf-16, and performs character mappping according to mc.
func pathToU16s(path string, mc MapChars) []uint16 {
	parts := strings.Split(path, `\`)
	var u16s []uint16
	for i, p := range parts {
		if len(p) == 0 {
			u16s = append(u16s, unicodeBackSlash)
		} else {
			u16s = append(u16s, mappedU16s(p, mc)...)
			if i != len(parts)-1 { // as long as we're not the last path component
				// even if we are not the last path component, we only want to add a separating slash
				// if the next part is a path component, and not just a terminating backslash
				if len(parts[i+1]) != 0 {
					u16s = append(u16s, unicodeBackSlash)
				}
			}
		}
	}

	return u16s
}

func toSFM(u16 uint16, endOfString bool) uint16 {
	if u16 >= 0x01 && u16 <= 0x1F {
		return uint16(u16) + 0xF000
	} else {
		switch u16 {
		case '"':
			return SFMDoubleQuote
		case '*':
			return SFMAsterisk
		case ':':
			return SFMColon
		case '<':
			return SFMLessThan
		case '>':
			return SFMGreaterThan
		case '?':
			return SFMQuestion
		case '|':
			return SFMPipe
		case '.':
			if endOfString {
				return SFMPeriod
			}
		case ' ':
			if endOfString {
				return SFMSpace
			}
		}
	}
	return 0
}

func toSFU(r uint16) uint16 {
	switch r {
	case ':', '*', '?', '<', '>', '|':
		return r + 0xF000
	}
	return 0
}
