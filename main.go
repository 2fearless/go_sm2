package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/ZZMarquis/gm/sm2"
	"io"
	"os"
	"strings"
)

var (
	pripem = "./pri.pem"
	pubpem = "./pub.pem"
)

func cmd1() {
	hexPri, basePub := Generate()
	file_put_contents(pripem, hexPri)
	binPub, _ := base64.StdEncoding.DecodeString(basePub)
	strPub := hex.EncodeToString(binPub)
	file_put_contents(pubpem, strPub)
}
func cmd2(data string) string {
	//读取文件
	fileContent, err := os.ReadFile(pubpem)
	if err != nil {
		fmt.Println("读取错误：", err)
	}
	byteData := []byte(data)

	strPub, _ := hex.DecodeString(string(fileContent))
	basePub := base64.StdEncoding.EncodeToString(strPub)
	pub := Base64ToPub(basePub)
	cipherText, err := sm2.Encrypt(pub, byteData, sm2.C1C3C2)
	if err != nil {
		fmt.Println(err.Error())
	}
	return hex.EncodeToString(cipherText)
}
func cmd3(data string) string {
	fileContent, err := os.ReadFile(pripem)
	if err != nil {
		fmt.Println("读取错误：", err)
	}
	byteData, _ := hex.DecodeString(data)
	pri := HexToPri(string(fileContent))
	//println(data)
	//println(string(fileContent))
	word, err := sm2.Decrypt(pri, byteData, sm2.C1C3C2)
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(word)
}
func cmd4(data string, salt string) string {
	fileContent, err := os.ReadFile(pripem)
	if err != nil {
		fmt.Println("读取错误：", err)
	}
	pri := HexToPri(string(fileContent))
	return Sign(data, pri, salt)
}
func cmd5(data string, sign string, salt string) bool {
	fileContent, err := os.ReadFile(pubpem)
	if err != nil {
		fmt.Println("读取错误：", err)
	}
	strPub, _ := hex.DecodeString(string(fileContent))
	basePub := base64.StdEncoding.EncodeToString(strPub)
	pub := Base64ToPub(basePub)
	return Verify(data, sign, pub, salt)
}
func file_put_contents(fileName string, content string) {
	var (
		file *os.File
		err  error
	)
	//文件是否存在
	if Exists(fileName) {
		//使用追加模式打开文件
		file, err = os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println("打开文件错误：", err)
			return
		}
	} else {
		//不存在创建文件
		file, err = os.Create(fileName)
		if err != nil {
			fmt.Println("创建失败", err)
			return
		}
	}

	defer file.Close()
	//写入文件
	_, err = io.WriteString(file, content)
	if err != nil {
		fmt.Println("写入错误：", err)
		return
	}
	//fmt.Println("写入成功：n=", n)
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func main() {
	var opt int
	var data string
	var sign string
	flag.IntVar(&opt, "o", 1, "操作方式")
	flag.StringVar(&data, "d", "", "数据")
	flag.StringVar(&sign, "s", "", "签名")
	salt := "1234567812345678"
	flag.Parse()
	switch opt {
	case 0:
		//重置秘钥
		cmd1()
		fmt.Println("秘钥对已重置")
	case 1:
		//加密
		fmt.Println(cmd2(data))
	case 2:
		//解密
		fmt.Println(cmd3(data))
	case 3:
		fmt.Println(cmd4(data, salt))
	case 4:
		fmt.Println(cmd5(data, sign, salt))
	default:
		fmt.Println("参数错误")
	}
}

// 公钥转base64
func PubToBase64(pub *sm2.PublicKey) string {
	pub.GetRawBytes()
	bytes := pub.X.Bytes()
	bytes = append(bytes, pub.Y.Bytes()...)
	//salt := "3059301306072a8648ce3d020106082a811ccf5501822d03420004"
	salt := ""
	pubHex := salt + hex.EncodeToString(bytes)
	decode, _ := hex.DecodeString(pubHex)
	base64Pub := base64.StdEncoding.EncodeToString(decode)
	return base64Pub
}

// 私钥转hex
func PriToHex(pri *sm2.PrivateKey) string {
	hexa := hex.EncodeToString(pri.GetRawBytes())
	return hexa
}

// 私钥生成公钥
func PriToBase64Pub(pri *sm2.PrivateKey) string {
	pub := sm2.CalculatePubKey(pri)
	base64Pub := PubToBase64(pub)
	return base64Pub
}

// 生成密钥对
func Generate() (string, string) {
	pri, pub, _ := sm2.GenerateKey(rand.Reader)
	return PriToHex(pri), PubToBase64(pub)
}

// Hex私钥转私钥对象
func HexToPri(priStr string) *sm2.PrivateKey {
	// 解码hex私钥
	privateKeyByte, _ := hex.DecodeString(priStr)
	// 转成go版的私钥
	pri, err := sm2.RawBytesToPrivateKey(privateKeyByte)
	if err != nil {
		panic("私钥加载异常")
	}
	return pri
}

func Sign(data string, pri *sm2.PrivateKey, salt string) string {
	//salt := "1234567812345678"
	//salt := ""
	signature, err := sm2.Sign(pri, []byte(salt), []byte(data))
	if err != nil {
		panic("签名错误")
	}
	// 转 base64
	sign := base64.StdEncoding.EncodeToString(signature)
	return sign
}

func Verify(data, sign string, pub *sm2.PublicKey, salt string) bool {
	//salt := "1234567812345678"
	//salt := ""
	sign1, _ := base64.StdEncoding.DecodeString(sign)
	return sm2.Verify(pub, []byte(salt), []byte(data), sign1)
}

// base64公钥转公钥对象
func Base64ToPub(pubStr string) *sm2.PublicKey {
	decode, _ := base64.StdEncoding.DecodeString(pubStr)
	pubHex := hex.EncodeToString(decode)
	pubHex = strings.ReplaceAll(pubHex, "3059301306072a8648ce3d020106082a811ccf5501822d03420004", "")
	pubByte, _ := hex.DecodeString(pubHex)
	pub, _ := sm2.RawBytesToPublicKey(pubByte)
	return pub
}
