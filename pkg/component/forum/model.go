package forum

type forumCreate struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	User  string `json:"user"`
}

type threadCreate struct {
	Author  string `json:"author"`
	Created string `json:"created"`
	Message string `json:"message"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
}
