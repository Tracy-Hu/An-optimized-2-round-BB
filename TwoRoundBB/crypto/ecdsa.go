package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
)

//-- secp256k1 --//
func SignECDSA(msg string, sk *ecdsa.PrivateKey) ECDSAsign {
	hash := sha256.Sum256([]byte(msg))
	r, s, _ := ecdsa.Sign(rand.Reader, sk, hash[:])
	sig := ECDSAsign{
		R: r,
		S: s,
	}
	return sig
}

func VrfECDSA(pk *ecdsa.PublicKey, sig ECDSAsign, msg string) bool {
	hash := sha256.Sum256([]byte(msg))
	return ecdsa.Verify(pk, hash[:], sig.R, sig.S)
}

//output signature type is string
func SignECDSAStr(msg string, sk *ecdsa.PrivateKey) string {
	hash := sha256.Sum256([]byte(msg))
	r, s, _ := ecdsa.Sign(rand.Reader, sk, hash[:])
	sig := ECDSAsign{
		R: r,
		S: s,
	}
	jsonMsg, _ := json.Marshal(sig)
	return string(jsonMsg)
}

func VrfECDSAStr(pk *ecdsa.PublicKey, sig string, msg string) bool {
	hash := sha256.Sum256([]byte(msg))
	var data ECDSAsign
	json.Unmarshal([]byte(sig), &data)
	return ecdsa.Verify(pk, hash[:], data.R, data.S)
}

func ConnectStr(s1 string, s2 string) string {
	return s1 + "\\" + s2
}
