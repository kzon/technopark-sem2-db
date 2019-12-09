package user

type (
	userInput struct {
		Email    string `json:"email"`
		Fullname string `json:"fullname"`
		About    string `json:"about"`
	}
)
