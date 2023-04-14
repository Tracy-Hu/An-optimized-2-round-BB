package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
)

func main() {
	n := 4
	id := make([]int, n)
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		id[i] = i + 1
		ids[i] = strconv.Itoa(i + 1)
		//fmt.Printf("id=%v,ids=%v\n",id[i],ids[i])
	}

	/////////// ECDSA ///////////
	prikeys := make([]string, n)
	pubkeys := make([]string, n)
	for i := 0; i < n; i++ {
		k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		privateKeyPem := encodeSK(k)
		publicKeyPem := encodePK(&k.PublicKey)
		prikeys[i] = privateKeyPem
		pubkeys[i] = publicKeyPem
	}

	path6 := "sks.txt"
	writeECDSA(path6, prikeys)

	path7 := "pks.txt"
	writeECDSA(path7, pubkeys)
}

// for ecdsa key encodes to string
func encodeSK(k *ecdsa.PrivateKey) string {
	sk, _ := x509.MarshalECPrivateKey(k)
	privateBlock := pem.Block{
		Type:    "PRIVATE KEY",
		Headers: nil,
		Bytes:   sk,
	}
	privateKeyPem := string(pem.EncodeToMemory(&privateBlock))

	return privateKeyPem
}

func encodePK(k *ecdsa.PublicKey) string {
	pk, _ := x509.MarshalPKIXPublicKey(k)
	publicKeyBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   pk,
	}
	publicKeyPem := string(pem.EncodeToMemory(&publicKeyBlock))
	return publicKeyPem
}

func writeECDSA(file string, k []string) {
	newFile, err := os.Create(file)
	if err != nil {
		fmt.Println("err is :", err)
	}
	defer newFile.Close()

	info, err := os.Stat(file)
	if err != nil {

		if os.IsNotExist(err) {
			fmt.Println("the file doesn't exist")
		} else {
			fmt.Println("err is :", err)
		}
	}

	bufferWrite := bufio.NewWriter(newFile)
	for i := 0; i < len(k); i++ {
		s := k[i] + "\n"
		bufferWrite.WriteString(s)
	}
	bufferWrite.Flush()
	fmt.Printf("file '%v' is done.\n", info.Name())
}
