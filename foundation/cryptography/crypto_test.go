package cryptography

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"testing"
)

func Test_Signature(t *testing.T) {
	type transaction struct {
		sender    string
		recipient string
		value     float32
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("failed to generate private key with err: %v", err)
	}

	senderPublicKeyString := GeneratePublicKeyString(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes())

	publicKeyString, err := PublicKeyFromString(senderPublicKeyString)
	if err != nil {
		t.Errorf("failed to generate PublicKeyFromString with err: %v", err)
	}

	b, err := json.Marshal(t)
	if err != nil {

	}
	h := sha256.Sum256(b)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, h[:])
	signature := &Signature{R: r, S: s}
	sign, err := SignatureFromString(signature.String())

	sender := "0x00"
	receiver := "0x01"
	tr := transaction{
		sender:    sender,
		recipient: receiver,
		value:     12,
	}

	bb, err := json.Marshal(tr)
	if err != nil {
		t.Errorf("failed to marshal transaction with err: %v", err)
	}
	hash := sha256.Sum256(bb)
	v := ecdsa.Verify(publicKeyString, hash[:], sign.R, sign.S)
	if !v {
		t.Errorf("signature prossess failed with err: %v", err)
	}
	t.Logf("\nSignature verified successfully\n -> %v <-", sign)
}
