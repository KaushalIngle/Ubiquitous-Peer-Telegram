package p2p

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"log"
	"os"
)

//GenerateKeyPair .
func GenerateKeyPair() (*rsa.PrivateKey, rsa.PublicKey) {
	priKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	pubKey := priKey.PublicKey
	// fmt.Println("Private Key : ", mariaPrivateKey)
	// fmt.Println("Public key ", mariaPublicKey)
	return priKey, pubKey
}

//GenerateOpenKeyPair .
func GenerateOpenKeyPair() (*rsa.PrivateKey, rsa.PublicKey) {
	priKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	pubKey := priKey.PublicKey
	// fmt.Println("Private Key : ", mariaPrivateKey)
	// fmt.Println("Public key ", mariaPublicKey)
	return priKey, pubKey
}

//GeneratePublicKey  .
func GeneratePublicKey(priKey *rsa.PrivateKey) rsa.PublicKey {
	pubKey := priKey.PublicKey
	return pubKey
}

//Encrypt .
func Encrypt(bob *rsa.PublicKey, message []byte) []byte {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(
		hash,
		rand.Reader,
		bob,
		message,
		[]byte(""))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return ciphertext
}

//Decrypt .
func Decrypt(priKey *rsa.PrivateKey, ciphertext []byte) []byte {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		priKey,
		ciphertext,
		[]byte(""))
	if err != nil {
		return []byte{}
	}
	return plaintext
}

//Sign .
func Sign(priKey *rsa.PrivateKey, message []byte) []byte {

	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto

	hashFunc := crypto.SHA256
	pssh := hashFunc.New()
	pssh.Write(message)
	hashed := pssh.Sum(nil)

	signature, err := rsa.SignPSS(
		rand.Reader,
		priKey,
		hashFunc,
		hashed,
		&opts)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return signature
}

//VerifySign .
func VerifySign(bob *rsa.PublicKey, signature []byte, message []byte) bool {
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	hashFunc := crypto.SHA256
	pssh := hashFunc.New()
	pssh.Write(message)
	hashed := pssh.Sum(nil)

	err := rsa.VerifyPSS(
		bob,
		hashFunc,
		hashed,
		signature,
		&opts)
	if err != nil {
		return false
	}
	return true

}

//ExportRsaPublicKey .
func ExportRsaPublicKey(pubkey *rsa.PublicKey) (string, error) {
	pubkeybytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return "", err
	}
	pubkeypem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkeybytes,
		},
	)

	return string(pubkeypem), nil
}

//ExportsRsaPublicKey .
func ExportsRsaPublicKey(pubkey *rsa.PublicKey) string {
	pubkeybytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return ""
	}
	pubkeypem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkeybytes,
		},
	)

	return string(pubkeypem)
}

//ParseRsaPublicKey .
func ParseRsaPublicKey(pub string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pub))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pubKey := pubKey.(type) {
	case *rsa.PublicKey:
		return pubKey, nil
	default:
		break // fall through
	}
	return nil, errors.New("Key type is not RSA")
}
