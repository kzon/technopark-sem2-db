package model

type Forum struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	User    string `db:"user"`
	Slug    string `db:"slug"`
	Posts   int    `db:"posts"`
	Threads int    `db:"threads"`
}

type User struct {
	ID       int    `db:"id"`
	Nickname string `db:"nickname"`
	Fullname string `db:"fullname"`
	About    string `db:"about"`
	Email    string `db:"email"`
}

type Thread struct {
	ID      int    `db:"id"`
	Title   string `db:"title"`
	User    string `db:"user"`
	Forum   string `db:"forum"`
	Message string `db:"message"`
	Votes   int    `db:"votes"`
	Slug    string `db:"slug"`
	Created string `db:"created"`
}
