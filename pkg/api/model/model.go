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
