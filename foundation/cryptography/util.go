package cryptography

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func (s *Signature) String() string {
	return fmt.Sprintf("%064x%064x", s.R, s.S)
}

func GenerateBlockchainAddress(pKey ecdsa.PrivateKey) string {
	// Perform SHA-256 hashing on the public key (32 bytes).
	h2 := sha256.New()
	h2.Write(pKey.X.Bytes())
	h2.Write(pKey.Y.Bytes())
	digest2 := h2.Sum(nil)
	// Perform RIPEMD-160 hashing on the result of SHA-256 (20 bytes).
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)
	// Add version byte in front of RIPEMD-160 hash (0x00 for Main Network).
	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])
	// Perform SHA-256 hash on the extended RIPEMD-160 result.
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)
	// Perform SHA-256 hash on the result of the previous SHA-256 hash.
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)
	// Take the first 4 bytes of the second SHA-256 hash for checksum.
	chsum := digest6[:4]
	// Add the 4 checksum bytes from 7 at the end of extended RIPEMD-160 hash from 4 (25 bytes).
	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], chsum[:])
	// Convert and return the result from a byte string into base58.
	return base58.Encode(dc8)
}

func SignatureFromString(s string) (*Signature, error) {
	x, y, err := StringToBigIntTupple(s)
	if err != nil {
		return nil, err
	}
	return &Signature{
		R: &x,
		S: &y,
	}, err
}

func StringToBigIntTupple(s string) (bix big.Int, biy big.Int, err error) {
	x, err := hex.DecodeString(s[:64])
	if err != nil {
		return big.Int{}, big.Int{}, err
	}
	y, err := hex.DecodeString(s[64:])
	if err != nil {
		return big.Int{}, big.Int{}, err
	}

	bix.SetBytes(x)
	biy.SetBytes(y)
	return
}

func PublicKeyFromString(s string) (*ecdsa.PublicKey, error) {
	x, y, err := StringToBigIntTupple(s)
	if err != nil {
		return nil, err
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     &x,
		Y:     &y,
	}, nil
}

func PrivateKeyFromString(s string, pubKey *ecdsa.PublicKey) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(s[:])
	if err != nil {
		return nil, err
	}
	var bi big.Int
	bi.SetBytes(b)
	return &ecdsa.PrivateKey{
		PublicKey: *pubKey,
		D:         nil,
	}, nil
}
