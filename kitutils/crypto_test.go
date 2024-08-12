package kitutils

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateSaltFromUUID(t *testing.T) {

	accountId := "25c05177-96a5-4197-82d0-6c3d85cf3b53"
	salt, err := GenerateSaltFromUUID(accountId, 13)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("盐值：", salt)
}

func TestEncryptStr(t *testing.T) {
	//accountId := "25c05177-96a5-4197-82d0-6c3d85cf3b53"
	key := "25c0517796a5419782d06c3d85cf3b53"
	salt := []byte{37, 192, 81, 119, 150, 165, 65, 151, 130, 208, 108, 61, 133, 108, 61, 133} // 16 字节的 IV
	apikey := "123456"
	encryptStr, err := EncryptStr(apikey, salt, key)
	if err != nil {
		// 如果出现错误，打印错误信息并结束测试
		t.Error("错误：", err)
	}
	fmt.Println("加密结果：", encryptStr)
	// m1I2N+6WM7b1+ZrJtYAyKA==
}

func TestDecryptStr(t *testing.T) {
	accountId := "25c05177-96a5-4197-82d0-6c3d85cf3b53"
	salt := []byte{37, 192, 81, 119, 150, 165, 65, 151, 130, 208, 108, 61, 133, 108, 61, 133} // 16 字节的 IV
	key := strings.ReplaceAll(accountId, "-", "")
	encryptStr := "m1I2N+6WM7b1+ZrJtYAyKA=="
	decryptStr, err := DecryptStr(encryptStr, salt, key)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("解密结果：", decryptStr)
}

func TestGetLoadEncryptionKey(t *testing.T) {
	// 在测试开始前设置环境变量
	os.Setenv("ENCRYPTION_KEY", "a9YQA9qo5OgMbBSy8K1ZfQjMLPTdAURd")

	key := LoadEncryptionKey()
	fmt.Printf("key的32字节byte：%x\n", key)
}

func TestLoadEncryptionKey(t *testing.T) {
	// 设置环境变量以确保 LoadEncryptionKey 可以正确读取
	os.Setenv("ENCRYPTION_KEY", "a9YQA9qo5OgMbBSy8K1ZfQjMLPTdAURd")

	// 期望的值
	expectedStr := "a9YQA9qo5OgMbBSy8K1ZfQjMLPTdAURd"
	var expectedKey [32]byte
	copy(expectedKey[:], expectedStr)

	tests := []struct {
		name string
		want *[32]byte
	}{
		// TODO: Add test cases.
		{
			name: "LoadEncryptionKey",
			want: &expectedKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LoadEncryptionKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadEncryptionKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncrypt(t *testing.T) { // 初始化密钥
	keyStr := "a9YQA9qo5OgMbBSy8K1ZfQjMLPTdAURd"
	var key [32]byte
	copy(key[:], keyStr)

	// 明文数据
	data := "Hello, World!"

	// 期望的加密结果（加密结果一般是动态的，尤其使用随机nonce，因此我们需要测试加密解密的整个过程）
	encrypted, err := Encrypt(data, &key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	type args struct {
		data string
		key  *[32]byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Encrypt Hello, World!",
			args: args{
				data: data,
				key:  &key,
			},
			want:    encrypted, // 因为加密结果每次可能不同，所以我们把刚加密的结果作为期望
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.args.data, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" || len(got) == 0 {
				t.Errorf("Encrypt() = %v, want a valid encrypted string", got)
			}
			if got != tt.want {
				fmt.Printf("Encrypt() = %v, want %v\n", got, tt.want)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	// 初始化密钥
	keyStr := "your-32-byte-key-goes-here!"
	var key [32]byte
	copy(key[:], keyStr)

	// 明文数据
	data := "Hello, World!"

	// 期望的加密结果
	encrypted, err := Encrypt(data, &key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	type args struct {
		encoded string
		key     *[32]byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Decrypt Hello, World!",
			args: args{
				encoded: encrypted,
				key:     &key,
			},
			want:    data, // 期望解密后的数据应该与原始数据相同
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decrypt(tt.args.encoded, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decrypt() = %v, want %v", got, tt.want)
			}
			fmt.Printf("got = %v, want = %v\n", got, tt.want)
		})
	}
}
