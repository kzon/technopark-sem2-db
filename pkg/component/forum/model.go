package forum

type (
	forumCreate struct {
		Slug  string `json:"slug"`
		Title string `json:"title"`
		User  string `json:"user"`
	}

	threadCreate struct {
		Author  string `json:"author"`
		Created string `json:"created"`
		Message string `json:"message"`
		Slug    string `json:"slug"`
		Title   string `json:"title"`
	}

	threadUpdate struct {
		Message string `json:"message"`
		Title   string `json:"title"`
	}

	postCreate struct {
		Author  string `json:"author"`
		Message string `json:"message"`
		Parent  int    `json:"parent"`
	}

	vote struct {
		Nickname string `json:"nickname"`
		Voice    int    `json:"voice"`
	}
)
