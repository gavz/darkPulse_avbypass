package Encrypt

import (
	"MyPacker/Others"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 生成随机密钥
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// 随机字符串
func GenerateRandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// 异或加密
func XOREncryption(shellcode []byte, key string) []byte {
	encrypted := make([]byte, len(shellcode))
	keyLen := len(key)

	for i := 0; i < len(shellcode); i++ {
		encrypted[i] = shellcode[i] ^ key[i%keyLen]
	}

	return encrypted
}

// AES中填充操作
func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// AES加密
func AESEncryption(key string, iv string, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	// Apply PKCS7 padding to ensure plaintext length is a multiple of the block size
	paddedData := PKCS7Padding(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(paddedData))

	// Create a new CBC mode encrypter
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedData)

	return ciphertext, nil
}

// BytesToUUIDs 将字节slice分割成多个16字节的组，并转换成UUID格式的字符串切片
func BytesToUUIDs(b []byte) ([]string, error) {
	var uuids []string
	chunkSize := 16

	for len(b) > 0 {
		// 如果剩余的字节不足16字节，则用0补足
		if len(b) < chunkSize {
			padding := make([]byte, chunkSize-len(b))
			b = append(b, padding...)
		}

		// 截取16字节的组
		chunk := b[:chunkSize]
		b = b[chunkSize:]

		// 将字节转换为十六进制字符串
		hexString := hex.EncodeToString(chunk)

		// 格式化UUID字符串
		uuid := fmt.Sprintf("%s%s%s%s-%s%s-%s%s-%s-%s",
			hexString[6:8],
			hexString[4:6],
			hexString[2:4],
			hexString[0:2],
			hexString[10:12],
			hexString[8:10],
			hexString[14:16],
			hexString[12:14],
			hexString[16:20],
			hexString[20:32])

		uuids = append(uuids, uuid)
	}

	return uuids, nil
}

// 加密函数
func Encryption(shellcodeBytes []byte, encryption string, keyLength int) (string, string, string) {
	//生成xor随机密钥
	switch encryption {
	case "xor":
		key := GenerateRandomString(keyLength)
		fmt.Printf("[+] Generated XOR key: ")
		Others.PrintKeyDetails(key)
		XorShellcode := XOREncryption(shellcodeBytes, key)
		hexXorShellcode := hex.EncodeToString((XorShellcode))
		//fomrattedXorShellcode := Converters.FormattedHexShellcode(string(hexXorShellcode))
		return hexXorShellcode, key, ""
	case "aes":
		key := GenerateRandomString(16)
		iv := GenerateRandomString(16)
		fmt.Printf("[+] Generated AES key: ")
		Others.PrintKeyDetails(key)
		fmt.Printf("[+] Generated IV (16-byte): ")
		Others.PrintKeyDetails(iv)
		keyNotification := Others.DetectNotification(keyLength)
		fmt.Printf("[+] Using AES-%d-CBC encryption\n\n", keyNotification)
		AESShellcode, _ := AESEncryption(key, iv, shellcodeBytes)
		hexXorShellcode := hex.EncodeToString(AESShellcode)

		return string(hexXorShellcode), key, iv
	}
	return "", "", ""
}

func HexStringToBytes(hexStr string) ([]byte, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// 混淆操作
func Obfuscation(obfuscation string, shellcodeString string) (string, string, string) {
	switch strings.ToLower(obfuscation) {
	case "uuid":
		bytes, err := HexStringToBytes(shellcodeString)
		uuids, err := BytesToUUIDs([]byte(bytes))
		if err != nil {
			fmt.Println("Error:", err)

		}
		fmt.Printf("[+] Generated UUIDs:")
		// 输出UUIDs
		for _, uuid := range uuids {
			fmt.Print("\"", uuid, "\",\n")
		}
		fmt.Println("")
		var uuidsString string
		for _, uuid := range uuids {
			uuidsString += "\"" + uuid + "\","
		}
		return uuidsString, "", ""
	case "words":
		//调用python脚本，获取dataset和words
		decoded, err := hex.DecodeString(shellcodeString)
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile("words\\enc.bin", decoded, 0644)
		if err != nil {
			panic(err)
		}
		dir, err := os.Getwd()
		dir1 := filepath.Join(dir, "words", "Shellcode-to-English.py")
		dir2 := filepath.Join(dir, "words", "enc.bin")
		words_path := filepath.Join(dir, "words", "words.txt")
		dataset_path := filepath.Join(dir, "words", "dataset.txt")
		cmd := exec.Command("python", dir1, dir2)
		_, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error executing Python script:", err)
			return "", "", ""
		}
		words, err := ioutil.ReadFile(words_path)
		if err != nil {
			log.Fatal(err)
		}
		dataset, err := ioutil.ReadFile(dataset_path)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("[+] Generated dataset" + string(dataset) + "\n")
		fmt.Println("[+] Generated words:" + string(dataset) + "\n")
		return "", string(words), string(dataset)
	}
	return "", "", ""
}
