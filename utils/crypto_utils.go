package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/deatil/go-cryptobin/cryptobin/crypto"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"os"
)

// GenerateSaltFromUUID 生成盐（IV），取UUID的前n字节和后16-n字节，凑够16字节
func GenerateSaltFromUUID(accountID string, n int) ([]byte, error) {
	u, err := uuid.Parse(accountID)
	if err != nil {
		return nil, fmt.Errorf("error parsing account ID: %v", err)
	}

	uuidBytes := u[:]

	if n < 0 || n > 16 {
		return nil, fmt.Errorf("n must be between 0 and 16")
	}

	frontPart := uuidBytes[:n]

	backPart := uuidBytes[len(uuidBytes)-(16-n):]

	salt := append(frontPart, backPart...)

	if len(salt) != 16 {
		return nil, fmt.Errorf("invalid salt length: %d", len(salt))
	}

	return salt, nil
}

// EncryptStr 加密函数，使用 salt 作为盐，key 作为密钥
func EncryptStr(plaintext string, salt []byte, key string) (string, error) {
	encrypted := crypto.
		FromString(plaintext).
		SetKey(key).  // 使用 key 作为密钥  必须是16、24、32字节(字符)
		WithIv(salt). // 使用盐作为 IV  16字节
		Aes().
		CBC().
		PKCS7Padding().
		Encrypt().
		ToBase64String()

	//fmt.Println("加密结果：", encrypted)
	return encrypted, nil
}

// DecryptStr 解密函数，使用 salt 作为盐，key 作为密钥
func DecryptStr(ciphertext string, salt []byte, key string) (string, error) {
	decrypted := crypto.
		FromBase64String(ciphertext).
		SetKey(key).  // 使用 key 作为密钥
		WithIv(salt). // 使用盐作为 IV
		Aes().
		CBC().
		PKCS7Padding().
		Decrypt().
		ToString()

	return decrypted, nil
}

// 使用环境变量来读取服务器上的key

// LoadEncryptionKey 从环境变量中加载加密密钥
func LoadEncryptionKey() *[32]byte {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		log.Error("ENCRYPTION_KEY environment variable is not set")
	}
	if len(keyStr) != 32 {
		log.Fatalf("Encryption key must be 32 characters long, but got %d characters", len(keyStr))
	}

	var key [32]byte
	copy(key[:], keyStr)
	return &key
}

// Encrypt the data and return a Base64-encoded string
func Encrypt(originStr string, key *[32]byte) (string, error) {
	data := []byte(originStr)
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", err
	}

	encrypted := secretbox.Seal(nonce[:], data, &nonce, key)

	// encode encrypted data into base64 strings
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return encoded, nil
}

// Decrypt decrypts Base64-encoded strings and returns the original data as a string
func Decrypt(encoded string, key *[32]byte) (string, error) {
	// decode Base64 encoded string to get the encrypted data
	encrypted, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	var nonce [24]byte
	copy(nonce[:], encrypted[:24])

	decrypted, ok := secretbox.Open(nil, encrypted[24:], &nonce, key)
	if !ok {
		return "", fmt.Errorf("decryption failed")
	}
	// Convert the decrypted byte slice to a string and return
	return string(decrypted), nil
}
