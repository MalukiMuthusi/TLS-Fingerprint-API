package session

import (
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

	var token models.Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		panic("failed to check the provided token")
	}

	return &token, nil
}
