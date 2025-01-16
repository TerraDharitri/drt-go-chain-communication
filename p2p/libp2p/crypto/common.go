package crypto

import (
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	crypto "github.com/TerraDharitri/drt-go-chain-crypto"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
)

// ConvertPrivateKeyToLibp2pPrivateKey will convert common private key to libp2p private key
func ConvertPrivateKeyToLibp2pPrivateKey(privateKey crypto.PrivateKey) (libp2pCrypto.PrivKey, error) {
	if check.IfNil(privateKey) {
		return nil, ErrNilPrivateKey
	}

	p2pPrivateKeyBytes, err := privateKey.ToByteArray()
	if err != nil {
		return nil, err
	}

	return libp2pCrypto.UnmarshalSecp256k1PrivateKey(p2pPrivateKeyBytes)
}
