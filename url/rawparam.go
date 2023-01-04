package urlutil

import (
	"bytes"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

/* Reserved Chars from RFC includes
! * ' ( ) ; : @ & = + $ , / ? % # [ ]
*/

// MustEscapeCharSet are special chars that are always escaped and are based on reserved chars from RFC
// Some of Reserved Chars From RFC were excluded and some were added for various reasons
// and goal here is to encode parameters key and value only
var MustEscapeCharSet []rune = []rune{'?', '#', '@', ';', '&', ',', '[', ']', '^'}

type Params map[string][]string

func NewParams() Params {
	p := make(Params)
	return p
}

// Add Parameters to store
func (p Params) Add(key string, value string) {
	if p.Has(key) {
		p[key] = append(p[key], value)
	} else {
		p[key] = []string{value}
	}
}

// Set sets the key to value and replaces if already exists
func (p Params) Set(key string, value string) {
	if p == nil {
		p = make(Params)
	}
	p[key] = []string{value}
}

// Get returns first value of given key
func (p Params) Get(key string) string {
	if p.Has(key) {
		return p[key][0]
	} else {
		return ""
	}
}

// Has returns if given key exists
func (p Params) Has(key string) bool {
	if p == nil {
		p = make(Params)
	}
	_, ok := p[key]
	return ok
}

// Del deletes values associated with key
func (p Params) Del(key string) {
	if p == nil {
		return
	} else {
		delete(p, key)
	}
}

// Encode URL encodes and returns values ("bar=baz&foo=quux") sorted by key.
func (p Params) Encode() string {
	if p == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := p[k]
		keyEscaped := ParamEncode(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(ParamEncode(v))
		}
	}
	return buf.String()
}

// ParamEncode  encodes Key characters only. key characters include
// whitespaces + non printable chars + non-ascii + `almost` all reserved
// with some execptions i.e (). and other
// also this does not double encode encoded characters
func ParamEncode(data string) string {
	return URLEncodeWithEscapes(data, MustEscapeCharSet...)
}

// URLEncodeWithEscapes URL encodes data with given special characters escaped (similar to burpsuite intruder)
// Note `MustEscapeCharSet` is not included
func URLEncodeWithEscapes(data string, charset ...rune) string {
	mustescape := getrunemap(charset)
	var buff bytes.Buffer
	totallen := len(data)
	// In any case
	buff.Grow(totallen)

	for _, r := range data {
		switch true {
		case r < rune(20):
			// control character
			buff.WriteRune('%')
			buff.WriteString(strconv.FormatInt(int64(r), 16)) // 2 digit hex
		case r == ' ':
			// prefer using + when space
			buff.WriteRune('+')
			// case
		case r < rune(127):
			if _, ok := mustescape[r]; ok {
				// reserved char must escape
				buff.WriteRune('%')
				buff.WriteString(strconv.FormatInt(int64(r), 16))
			} else {
				// do not percent encode
				buff.WriteRune(r)
			}
		case r == rune(127):
			// [DEL] char should be encoded
			buff.WriteRune('%')
			buff.WriteString(strconv.FormatInt(int64(r), 16))
		case r > rune(128):
			// non-ascii characters i.e chinese chars or any other utf-8
			buff.WriteRune('%')
			buff.WriteString(getutf8hex(r))
		}
	}
	return buff.String()
}

// PercentEncoding encodes all characters to percent encoded format just like burpsuite decoder
func PercentEncoding(data string) string {
	var buff bytes.Buffer
	totallen := len(data)
	// In any case
	buff.Grow(totallen)
	for _, r := range data {
		buff.WriteRune('%')
		if r <= rune(127) {
			// these are all ascii characters
			buff.WriteString(strconv.FormatInt(int64(r), 16))
		} else {
			// unicode characters
			buff.WriteString(getutf8hex(r))
		}
	}
	return buff.String()
}

// GetParams return Params type using url.Values
func GetParams(query url.Values) Params {
	if query == nil {
		return nil
	}
	p := NewParams()
	for k, v := range query {
		p[k] = v
	}
	return p
}

func getrunemap(runes []rune) map[rune]struct{} {
	x := map[rune]struct{}{}
	for _, v := range runes {
		x[v] = struct{}{}
	}
	return x
}

func getutf8hex(r rune) string {
	// Percent Encoding is only done in hexadecimal values and in ASCII Range only
	// other UTF-8 chars (chinese etc) can be used by utf-8 encoding and byte conversion
	// let golang do utf-8 encoding of rune
	var buff bytes.Buffer
	utfchar := string(r)
	hexencstr := hex.EncodeToString([]byte(utfchar))
	for k, v := range hexencstr {
		if k != 0 && k%2 == 0 {
			buff.WriteRune('%')
		}
		buff.WriteRune(v)
	}
	return buff.String()
}
