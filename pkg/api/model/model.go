package model

type (
	UserInput struct {
		Email    string `json:"email"`
		Fullname string `json:"fullname"`
		About    string `json:"about"`
	}

	ForumCreate struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
		User  string `json:"user"`
	}

	ThreadCreate struct {
		Author  string `json:"author"`
		Created string `json:"created"`
		Message string `json:"message"`
		Slug    string `json:"slug"`
		Title   string `json:"title"`
	}

	ThreadUpdate struct {
		Message string `json:"message"`
		Title   string `json:"title"`
	}

	PostCreate struct {
		Author  string `json:"author"`
		Message string `json:"message"`
		Parent  int    `json:"parent"`
	}

	PostUpdate struct {
		Message string `json:"message"`
	}

	PostOutput struct {
		ID       int    `db:"id" json:"id"`
		Parent   int    `db:"parent" json:"parent"`
		Author   string `db:"author" json:"author"`
		Forum    string `db:"forum" json:"forum"`
		Thread   int    `db:"thread" json:"thread"`
		Message  string `db:"message" json:"message"`
		IsEdited bool   `db:"isEdited" json:"isEdited"`
		Created  string `db:"created" json:"created"`
	}

	Vote struct {
		Nickname string `json:"nickname"`
		Voice    int    `json:"voice"`
	}

	Status struct {
		Forum  int `json:"forum"`
		Post   int `json:"post"`
		Thread int `json:"thread"`
		User   int `json:"user"`
	}
)
