package extension

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

type EncryptionKeys struct {
	Expiration time.Time
	Encrypt    cipher.BlockMode
	Decrypt    cipher.BlockMode
	Sign       func([]byte) []byte
}

func (a *Extension) GenerateEncryptionKeys(ctx context.Context) (*EncryptionKeys, error) {
	type Payload struct {
		PublicKey         string `json:"publicKey"`
		SessionIdentifier string `json:"sessionIdentifier"`
	}

	type Response struct {
		Expiration int64  `json:"expiration"`
		IV         string `json:"iv"`
		PublicKey  string `json:"publicKey"`
	}

	key, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	pkix, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return nil, err
	}

	payload := Payload{
		PublicKey:         base64.StdEncoding.EncodeToString(pkix),
		SessionIdentifier: a.clientReferenceId,
	}

	req, err := a.newWibRequest(ctx, http.MethodPost, "token/crypto/session/exchange", payload)
	if err != nil {
		return nil, err
	}

	var response Response
	if err := a.do(req, &response); err != nil {
		return nil, err
	}

	remoteBytes, err := base64.StdEncoding.DecodeString(response.PublicKey)
	if err != nil {
		return nil, err
	}

	remote, err := x509.ParsePKIXPublicKey(remoteBytes)
	if err != nil {
		return nil, err
	}

	remoteEcdh, err := remote.(*ecdsa.PublicKey).ECDH()
	if err != nil {
		return nil, err
	}

	shared, err := key.ECDH(remoteEcdh)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(shared)
	if err != nil {
		return nil, err
	}

	iv, err := base64.StdEncoding.DecodeString(response.IV)
	if err != nil {
		return nil, err
	}

	encrypt := cipher.NewCBCEncrypter(block, iv)
	decrypt := cipher.NewCBCDecrypter(block, iv)
	sign := func(b []byte) []byte {
		hash := hmac.New(sha256.New, shared)
		hash.Write(b)
		return hash.Sum(nil)
	}

	return &EncryptionKeys{
		Expiration: time.UnixMilli(response.Expiration),
		Encrypt:    encrypt,
		Decrypt:    decrypt,
		Sign:       sign,
	}, nil
}

func (a *Extension) GenerateKeys(ctx context.Context) (*EncryptionKeys, error) {
	a.muKeys.Lock()
	defer a.muKeys.Unlock()

	keys, err := a.GenerateEncryptionKeys(ctx)
	if err != nil {
		return nil, err
	}

	return keys, nil
}

// func (a *API) Encrypt(ctx context.Context, data string) (string, error) {
// 	keys, err := a.GetEncryptionKeys(ctx)
// 	if err != nil {
// 		return "", err
// 	}

// 	dataBytes := []byte(data)
// 	blockSize := keys.Encrypt.BlockSize()
// 	paddingSize := blockSize - len(dataBytes)%blockSize
// 	padding := bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)
// 	dataBytes = append(dataBytes, padding...)

// 	encrypted := make([]byte, len(dataBytes))
// 	keys.Encrypt.CryptBlocks(encrypted, dataBytes)

// 	return base64.StdEncoding.EncodeToString(encrypted), nil
// }

func getPublicKey(b64 string) *rsa.PublicKey {
	pubKey, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}

	key, err := x509.ParsePKIXPublicKey([]byte(pubKey))
	if err != nil {
		panic(err)
	}

	return key.(*rsa.PublicKey)
}

var (
	pubKey = getPublicKey(
		"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAiwcxZqZy5tgQiWuI6e0r" +
			"zoBelUXEuxtNLeCqFChS4peivGH0IP+QDnp2tOKQ2dA3sT8qu5YdOXhOc5RXC1Kq" +
			"P/Bbhvk0/R4m5UgPtq8YQcpbrC39GLyIWSEpGgDq7adjBa3cDoQdtkG3AfYVP6rK" +
			"FpKpMZZi0/MzUf+FMLnydqDez2pHAXzZTefq4OaUukBiKum764z7hEtNmxHQd4LN" +
			"hxcz6CHDlPyuWRYkDQt6S1iPCFnT0VZTjaEFTyUIMFDFFT1FhBVW+D9CIaLu/WAC" +
			"Az4QKEilsYZGOn7+vMhozKz2yzVFzRQI/r+WvnBzVFaY4BbbJd37dFzggpsUf+ze" +
			"9wIDAQAB",
	)
	ewaPubKey = getPublicKey(
		"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoXZDEi4FrW49rJIL+GQ1" +
			"NIT5OBxkWpHV82WiCZfjKdZ7IgAHAPJN+U3GvL8/Wv958B9ng/MqpXqAcHS9Oado" +
			"9IbbnvIhr+MwzJK9WL4OqoFm+iytnVTLNY72poxPTr+yF3Yb5lAgL8Nr2xuWPLEB" +
			"ofp01A0awqD6u+33mDWlG3vSeYVifFissVW+IE82qnO8imAX/MBvxZVJOcQe8zoC" +
			"Dld6YeyQ6KGNOBO9zPEu4e/May2YXYWG0A45l6BAFofX+I6bx/DLP9RLwLsYaD2i" +
			"JPsCW82dTn/9lbjfT/JL3AaSGhaei+iyRZPrnYNIs4l86GG4GwVr10ts912yAWIP" +
			"XQIDAQAB",
	)
)

func (a *Extension) EncryptEWA(ctx context.Context, data string) (string, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, ewaPubKey, []byte(data))
	if err != nil {
		return "", fmt.Errorf("ewa encrypt data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (a *Extension) Encrypt(ctx context.Context, data string) (string, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(data))
	if err != nil {
		return "", fmt.Errorf("encrypt data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (a *Extension) Decrypt(ctx context.Context, keys *EncryptionKeys, data string) (string, error) {
	dataBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	decrypted := make([]byte, len(dataBytes))
	keys.Decrypt.CryptBlocks(decrypted, dataBytes)

	paddingSize := int(decrypted[len(decrypted)-1])
	decrypted = decrypted[:len(decrypted)-paddingSize]

	return string(decrypted), nil
}
