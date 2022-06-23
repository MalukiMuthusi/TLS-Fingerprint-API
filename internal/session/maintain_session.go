package session

import (
	"fmt"
	"time"
	"tlsapi/internal/models"

	"github.com/beevik/ntp"
)

func ManageSession(t string) {

	for {
		// get session count
		token, err := GetToken(t)
		if err != nil {
			panic("token error")
		}

		count := token.SessionCount

		// increment session count
		UpdateSessionCount(t, count+1)

		time.Sleep(time.Second * 15)

		// get the new session count
		newToken, err := GetToken(t)
		if err != nil {
			panic("token error")
		}

		newCount := newToken.SessionCount

		// compare new session count with the previous count
		if newCount != count+1 {
			panic(fmt.Sprintf("multiple sessions using same token: %s", t))
		}

		CheckExpiry(newToken)
		CheckRevoked(newToken)

	}
}

func CheckExpiry(token *models.Token) {
	expiry, err := time.Parse(time.RFC3339, token.ExpiryDate)
	if err != nil {
		panic("provide a valid token")
	}

	now, err := ntp.Time("0.beevik-ntp.pool.ntp.org")
	if err != nil {
		now = time.Now()
	}

	if now.After(expiry) {
		panic("token is expired")
	}
}

func CheckRevoked(token *models.Token) {
	// check that the token is not revoked
	if token.Revoked {
		panic("Your access token was revoked")
	}

	// check the token is not archived
	if token.Archived {
		panic("token can no longer be used")
	}
}
