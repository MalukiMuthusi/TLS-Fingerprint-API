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
		fmt.Println(err)
	}

	resp, err := http.Post("https://tfa-z4mvziz65a-uc.a.run.app/session", "application/json", bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("session active: %v\n", sessionActive)
}
