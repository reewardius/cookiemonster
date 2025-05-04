package monster

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

type flaskParsedData struct {
	data             string
	timestamp        string
	decodedTimestamp []byte
	signature        string
	decodedSignature []byte
	algorithm        string
	toBeSigned       []byte

	compressed bool
	parsed     bool
}

func (d *flaskParsedData) String() string {
	if !d.parsed {
		return "Unparsed data"
	}

	return fmt.Sprintf("Compressed: %t\nData: %s\nTimestamp: %s\nSignature: %s\nAlgorithm: %s\n", d.compressed, d.data, d.timestamp, d.signature, d.algorithm)
}

const (
	flaskDecoder   = "flask"
	flaskMinLength = 10

	flaskSeparator = `.`
)

var (
	flaskAlgorithmLength = map[int]string{
		20: "sha1",
		32: "sha256",
		48: "sha384",
		64: "sha512",
	}

	flaskSalt = []byte(`cookie-session`)
)

func flaskDecode(c *Cookie) bool {
	if len(c.raw) < flaskMinLength {
		return false
	}

	rawData := c.raw
	var parsedData flaskParsedData

	// If the first character is a dot, it's compressed.
	if rawData[0] == '.' {
		parsedData.compressed = true
		rawData = rawData[1:]
	}

	// Break the cookie out into the session data, timestamp, and signature,
	// in that order. Note that we assume the use of `TimestampSigner`.
	components := strings.Split(rawData, flaskSeparator)
	if len(components) != 3 {
		return false
	}

	parsedData.data = components[0]
	parsedData.timestamp = components[1]
	parsedData.signature = components[2]

	// The current timestamp is embedded in a `>Q` Python struct. This can
	// be up to eight bytes (usually three), but never more.
	decodedTimestamp, err := base64.RawURLEncoding.DecodeString(parsedData.timestamp)
	if err != nil || len(decodedTimestamp) > 8 {
		return false
	}

	parsedData.decodedTimestamp = decodedTimestamp

	// Flask encodes the signature with URL-safe base64
	// without padding, so we must use `RawURLEncoding`.
	decodedSignature, err := base64.RawURLEncoding.DecodeString(parsedData.signature)
	if err != nil {
		return false
	}

	// Determine the algorithm from the digest length, or give up if we can't
	// figure it out.
	if alg, ok := flaskAlgorithmLength[len(decodedSignature)]; ok {
		parsedData.algorithm = alg
	} else {
		return false
	}

	parsedData.decodedSignature = decodedSignature
	parsedData.toBeSigned = []byte(parsedData.data + flaskSeparator + parsedData.timestamp)

	// If this is a compressed cookie, it needs to have the dot in front which
	// we previously stripped from `data`.
	if parsedData.compressed {
		parsedData.toBeSigned = append([]byte("."), parsedData.toBeSigned...)
	}

	parsedData.parsed = true
	c.wasDecodedBy(flaskDecoder, &parsedData)
	return true
}

func flaskUnsign(c *Cookie, secret []byte) bool {
	// We need to extract `toBeSigned` to prepare what we'll be signing.
	parsedData := c.parsedDataFor(flaskDecoder).(*flaskParsedData)

	// Derive the correct signature, if this was the correct secret key.
	computedSignature := flaskCompute(parsedData.algorithm, secret, parsedData.toBeSigned)

	// Compare this signature to the one in the `Cookie`.
	return bytes.Equal(parsedData.decodedSignature, computedSignature)
}

func flaskResign(c *Cookie, data string, secret []byte) string {
	// We need to extract `toBeSigned` to prepare what we'll be signing.
	parsedData := c.parsedDataFor(flaskDecoder).(*flaskParsedData)

	// We need to assemble the TBS string with new data.
	toBeSigned := base64.RawURLEncoding.EncodeToString([]byte(data)) + flaskSeparator + parsedData.timestamp

	return toBeSigned + flaskSeparator + base64.RawURLEncoding.EncodeToString(flaskCompute(parsedData.algorithm, secret, []byte(toBeSigned)))
}

func flaskCompute(algorithm string, secret []byte, data []byte) []byte {
	switch algorithm {
	case "sha1":
		// Flask forces us to derive a key for HMAC-ing.
		derivedKey := sha1HMAC(secret, flaskSalt)

		// Derive the correct signature, if this was the correct secret key.
		return sha1HMAC(derivedKey, data)
	case "sha256":
		// Flask forces us to derive a key for HMAC-ing.
		derivedKey := sha256HMAC(secret, flaskSalt)

		// Derive the correct signature, if this was the correct secret key.
		return sha256HMAC(derivedKey, data)
	case "sha384":
		// Flask forces us to derive a key for HMAC-ing.
		derivedKey := sha384HMAC(secret, flaskSalt)

		// Derive the correct signature, if this was the correct secret key.
		return sha384HMAC(derivedKey, data)
	case "sha512":
		// Flask forces us to derive a key for HMAC-ing.
		derivedKey := sha512HMAC(secret, flaskSalt)

		// Derive the correct signature, if this was the correct secret key.
		return sha512HMAC(derivedKey, data)
	default:
		panic("unknown algorithm")
	}
}
