package session

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"tlsapi/internal/models"
)

func UpdateSessionCount(token string, sessionCount int) {
	updateSessionCountPost := models.UpdateSessionCountPost{
		SessionCount: sessionCount,
		Token:        token,
	}

	b, err := json.Marshal(updateSessionCountPost)
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
