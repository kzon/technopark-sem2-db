package user

type userToInput struct {
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
}

type userOutput struct {
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	About    string `json:"about"`
	Nickname string `json:"nickname"`
}
