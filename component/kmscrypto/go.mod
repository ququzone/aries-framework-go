// Copyright Gen Digital Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/hyperledger/aries-framework-go/component/kmscrypto

go 1.19

require (
	github.com/bluele/gcache v0.0.0-20190518031135-bc40bd653833
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/go-jose/go-jose/v3 v3.0.1-0.20221117193127-916db76e8214
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.5.2
	github.com/google/tink/go v1.7.0
	github.com/hyperledger/aries-framework-go/component/log v0.0.0-20230417184158-344a7f82c4c2
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20230417184158-344a7f82c4c2
	github.com/hyperledger/ursa-wrapper-go v0.3.1
	github.com/kilic/bls12-381 v0.1.1-0.20210503002446-7b7597926c69
	github.com/stretchr/testify v1.8.1
	github.com/teserakt-io/golang-ed25519 v0.0.0-20210104091850-3888c087a4c8
	golang.org/x/crypto v0.8.0
	golang.org/x/sys v0.7.0
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/ethereum/go-ethereum v1.11.6 // indirect
	github.com/holiman/uint256 v1.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/hyperledger/aries-framework-go/component/log => ../log
	github.com/hyperledger/aries-framework-go/spi => ../../spi
)
