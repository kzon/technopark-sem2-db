package model

type (
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

	Vote struct {
		Nickname string `json:"nickname"`
		Voice    int    `json:"voice"`
	}
)
