package repository

func (r *Repository) CountForums() (count int, err error) {
	return r.count("forum")
}

func (r *Repository) CountPosts() (count int, err error) {
	return r.count("post")
}

func (r *Repository) CountThreads() (count int, err error) {
	return r.count("thread")
}

func (r *Repository) CountUsers() (count int, err error) {
	return r.count("user")
}

func (r *Repository) count(table string) (count int, err error) {
	err = r.db.Get(&count, `select count(*) from "`+table+`"`)
	return
}

func (r *Repository) Clear() error {
	_, err := r.db.Exec(`truncate thread, post, forum, "user", vote, forum_user`)
	return err
}
