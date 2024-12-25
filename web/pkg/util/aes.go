package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

//高级加密标准（Adevanced Encryption Standard ,AES）

//16,24,32位字符串的话，分别对应AES-128，AES-192，AES-256 加密方法
//key不能泄露
//var PwdKey = []byte("71f703cd0bdd4507")

// 重组aes_key
func RecombinationAesKey(key string) string {
	b := []byte(key)
	return string(b[5:10]) + string(b[0:5]) + string(b[10:16])
}

/**
加密base64  CBC
data: 需要加密的数据
pwdKey: aes密钥
*/
func AesEnCode(data []byte, pwdKey []byte) (string, error) {
	result, err := AesEcrypt(data, pwdKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), err
}

/**
解密base64  CBC
data: 需要解密的数据
pwdKey: aes密钥
*/
func AesDeCode(data string, pwdKey []byte) ([]byte, error) {
	//解密base64字符串
	pwdByte, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	//执行AES解密
	return AesDeCrypt(pwdByte, pwdKey)
}

//PKCS7 填充模式
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	//Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//填充的反向操作，删除填充字符串
//func PKCS7UnPadding(origData []byte) ([]byte, error) {
//	//获取数据长度
//	length := len(origData)
//	if length == 0 {
//		return nil, errors.New("string error！")
//	} else {
//		//获取填充字符串长度
//		unpadding := int(origData[length-1])
//		fmt.Println("length:", length)
//		fmt.Println("unpadding:", unpadding)
//		//截取切片，删除填充字节，并且返回明文
//		return origData[:(length - unpadding)], nil
//	}
//}

//填充的反向操作，删除填充字符串
func PKCS7UnPadding(origData []byte) ([]byte, error) {
	//获取数据长度
	length := len(origData)
	if length == 0 {
		return nil, errors.New("string error！")
	}
	//获取填充字符串长度
	unpadding := int(origData[length-1])
	if unpadding > length {
		return nil, errors.New("invalid padding size")
	}
	//截取切片，删除填充字节，并且返回明文
	return origData[:(length - unpadding)], nil
}

//实现加密 使用 CBC
func AesEcrypt(origData []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块的大小
	blockSize := block.BlockSize()
	//对数据进行填充，让数据长度满足需求
	origData = PKCS7Padding(origData, blockSize)
	//采用AES加密方法中CBC加密模式
	blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	//执行加密
	blocMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//实现解密  使用 CBC
func AesDeCrypt(cypted []byte, key []byte) ([]byte, error) {
	//创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//获取块大小
	blockSize := block.BlockSize()
	//创建加密客户端实例
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	//这个函数也可以用来解密
	blockMode.CryptBlocks(origData, cypted)
	//去除填充字符串
	origData, err = PKCS7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, err
}
