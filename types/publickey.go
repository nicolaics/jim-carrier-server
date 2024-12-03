package types

import "time"

type PublicKeyStore interface {
	UpdatePublicKey(id int, e []byte, m []byte) error
	GetPublicKeyByUserID(userId int) (*PublicKey, error)
}

type PublicKey struct {
	ID             int       `json:"id"`
	UserID         int       `json:"userId"`
	E              string    `json:"e"`
	M              string    `json:"m"`
	LastModifiedAt time.Time `json:"lastModifiedAt"`
}
