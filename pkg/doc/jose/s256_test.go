package jose

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestS256Curve(t *testing.T) {
	curve := S256()

	require.EqualValues(t, "secp256k1", curve.Params().Name)
	require.EqualValues(t, "7", curve.Params().B.String())
}
