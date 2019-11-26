package orbitutil

import (
	"encoding/hex"

	"berty.tech/go-ipfs-log/keystore"

	"berty.tech/go/pkg/errcode"
	"github.com/libp2p/go-libp2p-core/crypto"
)

type BertySignedKeyStore struct {
	keys map[string]crypto.PrivKey
}

func (s *BertySignedKeyStore) SetKey(pk crypto.PrivKey) error {
	pubKeyBytes, err := pk.GetPublic().Raw()
	if err != nil {
		return errcode.TODO.Wrap(err)
	}

	keyID := hex.EncodeToString(pubKeyBytes)

	s.keys[keyID] = pk

	return nil
}

func (s *BertySignedKeyStore) HasKey(id string) (bool, error) {
	_, ok := s.keys[id]

	return ok, nil
}

func (s *BertySignedKeyStore) CreateKey(id string) (crypto.PrivKey, error) {
	return s.GetKey(id)
}

func (s *BertySignedKeyStore) GetKey(id string) (crypto.PrivKey, error) {
	if privKey, ok := s.keys[id]; ok {
		return privKey, nil
	}

	return nil, errcode.ErrGroupMemberUnknownGroupID
}

func (s *BertySignedKeyStore) Sign(privKey crypto.PrivKey, bytes []byte) ([]byte, error) {
	return privKey.Sign(bytes)
}

func (s *BertySignedKeyStore) Verify(signature []byte, publicKey crypto.PubKey, data []byte) error {
	ok, err := publicKey.Verify(data, signature)
	if err != nil {
		return err
	}

	if !ok {
		return errcode.ErrGroupMemberLogEventSignature
	}

	return nil
}

func NewBertySignedKeyStore() *BertySignedKeyStore {
	ks := &BertySignedKeyStore{
		keys: map[string]crypto.PrivKey{},
	}

	return ks
}

var _ keystore.Interface = (*BertySignedKeyStore)(nil)
