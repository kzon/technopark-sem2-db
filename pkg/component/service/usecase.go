package service

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) Usecase {
	return Usecase{repo: repo}
}

func (u *Usecase) getStatus() (s status, err error) {
	forum, err := u.repo.countForums()
	if err != nil {
		return
	}
	post, err := u.repo.countPosts()
	if err != nil {
		return
	}
	thread, err := u.repo.countThreads()
	if err != nil {
		return
	}
	user, err := u.repo.countUsers()
	if err != nil {
		return
	}
	s = status{
		Forum:  forum,
		Post:   post,
		Thread: thread,
		User:   user,
	}
	return
}

func (u *Usecase) clear() error {
	return u.repo.clear()
}
