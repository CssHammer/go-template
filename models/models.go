package models

type HelloMessage struct {
	Message string `json:"message"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
