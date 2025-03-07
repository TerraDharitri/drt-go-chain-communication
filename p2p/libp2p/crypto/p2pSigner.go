package crypto

import (
	"crypto/sha256"

	"github.com/TerraDharitri/drt-go-chain-communication/p2p"
	"github.com/TerraDharitri/drt-go-chain-core/core"
	"github.com/TerraDharitri/drt-go-chain-core/core/check"
	crypto "github.com/TerraDharitri/drt-go-chain-crypto"
)

// ArgsP2pSignerWrapper defines the arguments needed to create a p2p signer wrapper
type ArgsP2pSignerWrapper struct {
	PrivateKey      crypto.PrivateKey
	Signer          crypto.SingleSigner
	KeyGen          crypto.KeyGenerator
	P2PKeyConverter p2p.P2PKeyConverter
}

type p2pSignerWrapper struct {
	privateKey crypto.PrivateKey
	signer     crypto.SingleSigner
	keyGen     crypto.KeyGenerator
	p2pKeyConv p2p.P2PKeyConverter
}

// NewP2PSignerWrapper creates a new p2pSigner instance
func NewP2PSignerWrapper(args ArgsP2pSignerWrapper) (*p2pSignerWrapper, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &p2pSignerWrapper{
		privateKey: args.PrivateKey,
		signer:     args.Signer,
		keyGen:     args.KeyGen,
		p2pKeyConv: args.P2PKeyConverter,
	}, nil
}

func checkArgs(args ArgsP2pSignerWrapper) error {
	if check.IfNil(args.PrivateKey) {
		return ErrNilPrivateKey
	}
	if check.IfNil(args.Signer) {
		return ErrNilSingleSigner
	}
	if check.IfNil(args.KeyGen) {
		return ErrNilKeyGenerator
	}
	if check.IfNil(args.P2PKeyConverter) {
		return ErrNilP2PKeyConverter
	}

	return nil
}

// Sign will sign the hash of the payload with the internal private key
func (psw *p2pSignerWrapper) Sign(payload []byte) ([]byte, error) {
	// added hash over the payload to comply with libp2p internal implementation
	hash := sha256.Sum256(payload)
	return psw.signer.Sign(psw.privateKey, hash[:])
}

// Verify will check that the (hash of the payload, peer ID, signature) tuple is valid or not
func (psw *p2pSignerWrapper) Verify(payload []byte, pid core.PeerID, signature []byte) error {
	pubKey, err := psw.p2pKeyConv.ConvertPeerIDToPublicKey(psw.keyGen, pid)
	if err != nil {
		return err
	}

	// added hash over the payload to comply with libp2p internal implementation
	hash := sha256.Sum256(payload)
	err = psw.signer.Verify(pubKey, hash[:], signature)
	if err != nil {
		return err
	}

	return nil
}

// SignUsingPrivateKey will sign the hash of the payload with provided private key bytes
func (psw *p2pSignerWrapper) SignUsingPrivateKey(skBytes []byte, payload []byte) ([]byte, error) {
	sk, err := psw.keyGen.PrivateKeyFromByteArray(skBytes)
	if err != nil {
		return nil, err
	}

	// added hash over the payload to comply with libp2p internal implementation
	hash := sha256.Sum256(payload)
	return psw.signer.Sign(sk, hash[:])
}
