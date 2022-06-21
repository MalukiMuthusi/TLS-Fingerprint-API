package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"tlsapi/internal/models"
)

func UpdateSession(token string, sessionActive bool) {

	updateSessionPost := models.UpdateSessionPost{
		SessionActive: sessionActive,
		Token:         token,
	}

	b, err := json.Marshal(updateSessionPost)
	if err != nil {
		fmt.Print(err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://tfa-fur355ca3q-uc.a.run.app/session", bytes.NewBuffer(b))
	if err != nil {
		fmt.Print(err)
	}

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err)
	}

}
