package models

type Token struct {
	Id           string `json:"id" firestore:"id"`
	CreationDate string `json:"creation_date" firestore:"creation_date"`
	ExpiryDate   string `json:"expiry_date" firestore:"expiry_date"`
	Revoked      bool   `json:"revoked" firestore:"revoked"`
	Archived     bool   `json:"archive" firestore:"archive"`
}
