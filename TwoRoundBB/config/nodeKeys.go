package config

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strconv"
)

type NodeKey struct {
	NodeID  string
	NodeSk  *ecdsa.PrivateKey
	NodePks map[string]*ecdsa.PublicKey // all n nodes pk
}

func LoadNodeKey(nodeID string) NodeKey {
	//--------ecdsa---------//
	Pks := map[string]*ecdsa.PublicKey{}
	for i := 0; i < N; i++ {
		pubS, err1 := read("pks.txt", i, 4)
		if err1 != nil {
			fmt.Println(err1)
			return NodeKey{}
		}
		publicKey := decodePK(pubS)
		Pks[strconv.Itoa(i+1)] = publicKey
	}

	id, _ := strconv.Atoi(nodeID)
	priS, err2 := read("sks.txt", id-1, 5)
	if err2 != nil {
		fmt.Println(err2)
		return NodeKey{}
	}
	Sk := decodeSK(priS)
	//-------end ecdsa--------//

	return NodeKey{nodeID, Sk, Pks}
}

func read(name string, id int, len int) (string, error) {
	file, err := os.Open(name)
	if err != nil {
		return "0", err
	}
	defer file.Close()

	//创建一个 *Reader ， 是带缓冲的
	reader := bufio.NewReader(file)
	var s string

	for i := 0; i < id; i++ {
		for j := 0; j < len+1; j++ {
			_, _ = reader.ReadString('\n')
		}
	}

	for i := 0; i < len; i++ {
		str, err := reader.ReadString('\n') //读到一个换行就结束
		if err == io.EOF {                  //io.EOF 表示文件的末尾
			break
		}
		s += str
	}
	return s, nil
}

func decodeSK(privateKeyPem string) *ecdsa.PrivateKey {
	blockPri, _ := pem.Decode([]byte(privateKeyPem))
	x509EncodedPri := blockPri.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509EncodedPri)
	return privateKey
}

func decodePK(publicKeyPem string) *ecdsa.PublicKey {
	blockPub, _ := pem.Decode([]byte(publicKeyPem))
	x509EncodedPub := blockPub.Bytes
	genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
	publicKey := genericPublicKey.(*ecdsa.PublicKey)
	return publicKey
}
