package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var PwdKey = []byte("fwerc3e2ex21e234x3565v4v")

// aes加密+crt
func aesEncrypt(plainText, key []byte) ([]byte, error) {
	// 1.创建一个底层使用aes的密码接口的对象
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 3.创建一个使用ctr加密接口对象
	iv := []byte("12345678abcdefgh") //初始化向量
	stream := cipher.NewCTR(block, iv)
	// 4.加密
	cipherText := make([]byte, len(plainText))
	stream.XORKeyStream(cipherText, plainText)
	return cipherText, nil
}

// aes解密
func aesDecrypt(cipherText, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := []byte("12345678abcdefgh")
	stream := cipher.NewCTR(block, iv)
	plainText := make([]byte, len(cipherText))
	stream.XORKeyStream(plainText, cipherText)
	return plainText
}

// base64编码密文
func EnPwdCode(pwd []byte) (string, error) {
	result, err := aesEncrypt(pwd, PwdKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

// base64解码,并解密aes
func DePwdCode(pwd string) ([]byte, error) {
	pwdByte, err := base64.StdEncoding.DecodeString(pwd)
	if err != nil {
		return nil, err
	}
	return aesDecrypt(pwdByte, PwdKey), nil
}
