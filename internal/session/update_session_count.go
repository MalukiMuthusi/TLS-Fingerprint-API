package session

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"tlsapi/internal/models"
)

var bys = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func UpdateSessionCount(token string, sessionCount int) {
	updateSessionCountPost := models.UpdateSessionCountPost{
		SessionCount: sessionCount,
		Token:        token,
	}

	encryptedToken, err := Encrypt(&updateSessionCountPost)
	if err != nil {
		panic(err)
	}

	encryptedSessionCount := models.EncryptedSessionCount{
		Token: encryptedToken,
	}

	b, err := json.Marshal(encryptedSessionCount)
	if err != nil {
		fmt.Println(err)
	}

	resp, err := http.Post("https://tfa-z4mvziz65a-uc.a.run.app/session_count", "application/json", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
}

func Encrypt(token *models.UpdateSessionCountPost) (string, error) {

	block, err := aes.NewCipher([]byte("abc&1*~#^2^#s0^=)^^7%b34"))
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, bys)
	cipherText := make([]byte, len(b))
	cfb.XORKeyStream(cipherText, b)

	return Encode(cipherText), nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}
