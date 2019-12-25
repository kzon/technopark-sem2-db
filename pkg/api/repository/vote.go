package repository

import (
	"database/sql"
	"github.com/kzon/technopark-sem2-db/pkg/model"
)

func (r *Repository) AddThreadVote(thread *model.Thread, nickname string, voice int) (newVotes int, err error) {
	oldVoice, err := r.getVoice(nickname, thread.ID)
	if err != nil {
		return
	}
	if oldVoice == voice {
		return thread.Votes, nil
	}
	tx, err := r.db.Beginx()
	if err != nil {
		return
	}
	err = tx.Get(
		&newVotes,
		`update thread set votes = votes + $1 where id = $2 returning votes`,
		voice-oldVoice, thread.ID,
	)
	if err != nil {
		tx.Rollback()
		return
	}
	if _, err = tx.Exec(`delete from vote where thread = $1 and nickname = $2`, thread.ID, nickname); err != nil {
		tx.Rollback()
		return
	}
	if _, err = tx.Exec(`insert into vote (thread, nickname, voice) values ($1, $2, $3)`, thread.ID, nickname, voice); err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

func (r *Repository) getVoice(nickname string, threadID int) (int, error) {
	var voice int
	err := r.db.Get(&voice, `select voice from vote where nickname = $1 and thread = $2`, nickname, threadID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return voice, err
}
