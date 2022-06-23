package session

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"tlsapi/internal/models"
)

func GetToken(t string) (*models.Token, error) {

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://tfa-z4mvziz65a-uc.a.run.app/token/%s", t), nil)
	if err != nil {
		panic(err)
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		panic("failed to check the provided token")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("failed to check the provided token")
	}

	var encryptedToken models.GetTokenResponse

	err = json.Unmarshal(body, &encryptedToken)
	if err != nil {
		panic("failed to check the provided token")
	}

	token, err := Decrypt(encryptedToken)
	if err != nil {
		panic("error getting token")
	}

	return token, nil
}

func Decrypt(encryptedToken models.GetTokenResponse) (*models.Token, error) {

	block, err := aes.NewCipher([]byte("abc&1*~#^2^#s0^=)^^7%b34"))
	if err != nil {
		return nil, err
	}

	cipherText := Decode(encryptedToken.Token)

	cfb := cipher.NewCFBDecrypter(block, bys)

	plainText := make([]byte, len(cipherText))

	cfb.XORKeyStream(plainText, cipherText)

	var token models.Token

	err = json.Unmarshal(plainText, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}
