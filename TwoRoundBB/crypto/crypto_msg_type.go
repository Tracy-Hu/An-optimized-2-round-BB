package crypto

import "math/big"

type ECDSAsign struct {
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`
}
