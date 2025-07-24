package base58

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"
)

type testValues struct {
	dec []byte
	enc string
}

var (
	n         = 5000000
	testPairs = make([]testValues, 0, n)
)

func init() {
	// If we do not seed the prng - it will default to a seed of (1)
	// https://golang.org/pkg/math/rand/#Seed
	rand.Seed(time.Now().UTC().UnixNano())
}

func initTestPairs() {
	if len(testPairs) > 0 {
		return
	}
	// pre-make the test pairs, so it doesn't take up benchmark time...
	for i := 0; i < n; i++ {
		data := make([]byte, 32)
		rand.Read(data)
		testPairs = append(testPairs, testValues{dec: data, enc: FastBase58Encoding(data)})
	}
}

func randAlphabet() *Alphabet {
	// Permutes [0, 127] and returns the first 58 elements.
	var randomness [128]byte
	rand.Read(randomness[:])

	var bts [128]byte
	for i, r := range randomness {
		j := int(r) % (i + 1)
		bts[i] = bts[j]
		bts[j] = byte(i)
	}
	return NewAlphabet(string(bts[:58]))
}

var btcDigits = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func TestInvalidAlphabetTooShort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet being too short did not occur")
		}
	}()

	_ = NewAlphabet(btcDigits[1:])
}

func TestInvalidAlphabetTooLong(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet being too long did not occur")
		}
	}()

	_ = NewAlphabet("0" + btcDigits)
}

func TestInvalidAlphabetNon127(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet containing non-ascii chars did not occur")
		}
	}()

	_ = NewAlphabet("\xFF" + btcDigits[1:])
}

func TestInvalidAlphabetDup(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on alphabet containing duplicate chars did not occur")
		}
	}()

	_ = NewAlphabet("z" + btcDigits[1:])
}

func TestFastEqTrivialEncodingAndDecoding(t *testing.T) {
	for k := 0; k < 10; k++ {
		testEncDecLoop(t, randAlphabet())
	}
	testEncDecLoop(t, BTCAlphabet)
	testEncDecLoop(t, FlickrAlphabet)
}

func testEncDecLoop(t *testing.T, alph *Alphabet) {
	for j := 1; j < 256; j++ {
		b := make([]byte, j)
		for i := 0; i < 100; i++ {
			rand.Read(b)
			fe := FastBase58EncodingAlphabet(b, alph)
			te := TrivialBase58EncodingAlphabet(b, alph)

			if fe != te {
				t.Errorf("encoding err: %#v", hex.EncodeToString(b))
			}

			fd, ferr := FastBase58DecodingAlphabet(fe, alph)
			if ferr != nil {
				t.Errorf("fast error: %v", ferr)
			}
			td, terr := TrivialBase58DecodingAlphabet(te, alph)
			if terr != nil {
				t.Errorf("trivial error: %v", terr)
			}

			if hex.EncodeToString(b) != hex.EncodeToString(td) {
				t.Errorf("decoding err: %s != %s", hex.EncodeToString(b), hex.EncodeToString(td))
			}
			if hex.EncodeToString(b) != hex.EncodeToString(fd) {
				t.Errorf("decoding err: %s != %s", hex.EncodeToString(b), hex.EncodeToString(fd))
			}
		}
	}
}

func BenchmarkTrivialBase58Encoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		TrivialBase58Encoding([]byte(testPairs[i].dec))
	}
}

func BenchmarkFastBase58Encoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FastBase58Encoding(testPairs[i].dec)
	}
}

func BenchmarkTrivialBase58Decoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		TrivialBase58Decoding(testPairs[i].enc)
	}
}

func BenchmarkFastBase58Decoding(b *testing.B) {
	initTestPairs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FastBase58Decoding(testPairs[i].enc)
	}
}

func TestAppend(t *testing.T) {
	initTestPairs()
	for i := 0; i < len(testPairs); i++ {
		// Append the encoding to an empty slice.
		enc := FastBase58Encoding(testPairs[i].dec)
		dst := make([]byte, 0)
		dst = Append(dst, testPairs[i].dec)
		if string(dst) != enc {
			t.Errorf("Append failed: expected %s, got %s", enc, string(dst))
		}
	}
}

func TestSanityCheck(t *testing.T) {
	testCases := []string{
		"ComputeBudget111111111111111111111111111111",
		"11111111111111111111111111111111",
		"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
	}
	// parse then encode again
	for _, tc := range testCases {
		dec, err := FastBase58Decoding(tc)
		if err != nil {
			t.Errorf("Failed to decode %s: %v", tc, err)
			continue
		}
		enc := FastBase58Encoding(dec)
		if enc != tc {
			t.Errorf("Sanity check failed: expected %s, got %s", tc, enc)
		}
	}
}
