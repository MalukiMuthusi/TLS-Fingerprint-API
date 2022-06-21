package models

type Token struct {
	Id            string `json:"id" firestore:"id"`
	CreationDate  string `json:"creation_date" firestore:"creation_date"`
	ExpiryDate    string `json:"expiry_date" firestore:"expiry_date"`
	Revoked       bool   `json:"revoked" firestore:"revoked"`
	Archived      bool   `json:"archive" firestore:"archive"`
	SessionActive bool   `json:"session_active" firestore:"session_active"`
	SessionCount  int    `json:"session_count" firestore:"session_count"`
}

type UpdateSessionPost struct {
	SessionActive bool   `json:"session_active" form:"session_active"`
	Token         string `json:"token" form:"token"`
}
