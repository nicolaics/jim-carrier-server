package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"strconv"
)

// EncryptData encrypts data with public key
func EncryptData(msg []byte, pubE string, pubM string) (string, error) {
	m := new(big.Int)
	m, ok := m.SetString(pubM, 10)
	if !ok {
		return "", fmt.Errorf("error set string")
	}

	e, err := strconv.Atoi(pubE)
	if err != nil {
		return "", err
	}

	publicKey := &rsa.PublicKey{
		N: m,
		E: e,
	}
	
	hash := sha256.New()

	encrypted, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, msg, nil)
	if err != nil {
		log.Println("error encrypted: ", err.Error())
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}