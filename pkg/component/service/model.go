package service

type status struct {
	// Кол-во пользователей в базе данных.
	User int `json:"user"`

	// Кол-во разделов в базе данных.
	Forum int `json:"forum"`

	// Кол-во веток обсуждения в базе данных.
	Thread int `json:"thread"`

	// Кол-во сообщений в базе данных.
	Post int `json:"post"`
}
