package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

const (
	password = "x35k9f"
	msg      = `0ba7cd8c624345451df4710b81d1a349ce401e61bc7eb704ca` +
		`a84a8cde9f9959699f75d0d1075d676f1fe2eb475cf81f62ef` +
		`f701fee6a433cfd289d231440cf549e40b6c13d8843197a95f` +
		`8639911b7ed39a3aec4dfa9d286095c705e1a825b10a9104c6` +
		`be55d1079e6c6167118ac91318fe`
)

func TestCipher(t *testing.T) {
	key := sha256.Sum256([]byte(password))
	fmt.Printf("key=%d\n", len(key))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		fmt.Printf("error NewCipher: %v\n", err)
		return
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		fmt.Printf("error NewGCM: %v\n", err)
		return
	}

	data, err := hex.DecodeString(msg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("NonceSize=%d\n", aesgcm.NonceSize())
	result, err := aesgcm.Open(nil, key[len(key)-aesgcm.NonceSize():], data, nil)
	if err != nil {
		fmt.Printf("error Open: %v\n", err)
		return
	}
	fmt.Printf("%s", result)

	// допишите код
	// 1) получите ключ из password, используя sha256.Sum256
	// 2) создайте aesblock и aesgcm
	// 3) получите вектор инициализации aesgcm.NonceSize() байт с конца ключа
	// 4) декодируйте сообщение msg в двоичный формат
	// 5) расшифруйте и выведите данные

	// ...
}
