package jose

import (
	"crypto/elliptic"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

var s256Curve = &S256Curve{secp256k1.S256()}

type S256Curve struct {
	*secp256k1.BitCurve
}

func S256() elliptic.Curve {
	return s256Curve
}

func (BitCurve *S256Curve) Params() *elliptic.CurveParams {
	return &elliptic.CurveParams{
		P:       BitCurve.P,
		N:       BitCurve.N,
		B:       BitCurve.B,
		Gx:      BitCurve.Gx,
		Gy:      BitCurve.Gy,
		BitSize: BitCurve.BitSize,
		Name:    "secp256k1",
	}
}
