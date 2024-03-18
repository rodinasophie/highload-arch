package storage

import (
	"context"
	"highload-arch/pkg/common"
	"log"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
)

const DIALOG_PENDING_UNREAD_STATE = "PENDING_UNREAD"
const DIALOG_UNREAD_STATE = "UNREAD"
const DIALOG_PENDING_READ_STATE = "PENDING_READ"
const DIALOG_READ_STATE = "READ"

type SendRequest struct {
	ID          string    `pg:"id"`
	AuthorID    string    `pg:"author_id"`
	RecepientID string    `pg:"recepient_id"`
	DialogID    string    `pg:"dialog_id"`
	CreatedAt   time.Time `pg:"created_at"`
	Text        string    `pg:"text"`
	State       string    `pg:"state"`
}

func GetDialogId(author_id, recepient_id string) string {
	dialogID := ""
	if author_id > recepient_id {
		dialogID = author_id + "_" + recepient_id
	} else {
		dialogID = recepient_id + "_" + author_id
	}
	return dialogID
}

func (req *SendRequest) dbAddDialogMessage(ctx context.Context) (string, error) {
	dialogID := GetDialogId(req.AuthorID, req.RecepientID)
	var id string
	err := db.QueryRow(ctx,
		`INSERT INTO dialogs (author_id, recepient_id, dialog_id, text, created_at, state) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		req.AuthorID, req.RecepientID, dialogID, req.Text, req.CreatedAt, req.State).Scan(&id)

	return id, err
}

func dbGetDialogWithState(ctx context.Context, userID, to string, states []string) ([]SendRequest, error) {
	res := []SendRequest{}
	dialogID := GetDialogId(userID, to)

	rows, err := db.Query(ctx,
		`SELECT id, author_id, recepient_id, created_at, text FROM dialogs WHERE dialog_id = $1 AND state in ($2)`, dialogID, strings.Join(states, ", "))
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	if err := pgxscan.ScanAll(&res, rows); err != nil {
		return nil, err
	}

	return res, err
}

func dbUpdateMessageState(ctx context.Context, id, from_state, to_state string) error {
	rows, err := db.Query(ctx,
		`UPDATE dialogs SET state = $1 WHERE id = $2 and state = $3`, to_state, id, from_state)
	defer rows.Close()
	return err
}

func MessagedUpdated(ctx context.Context, req *common.MessageCountRequest) error {
	var err error
	if req.Action == common.INCREMENT_MESSAGE_COUNT_ACTION {
		err = dbUpdateMessageState(ctx, req.MessageID, DIALOG_PENDING_UNREAD_STATE, DIALOG_UNREAD_STATE)
	} else if req.Action == common.DECREMENT_MESSAGE_COUNT_ACTION {
		err = dbUpdateMessageState(ctx, req.MessageID, DIALOG_PENDING_READ_STATE, DIALOG_READ_STATE)
	} else {
		log.Printf("Unknown action: %s", req.Action)
	}
	return err
}

func SendMessageDB(ctx context.Context, userID, to, text string) (string, error) {
	req := &SendRequest{AuthorID: userID, Text: text, CreatedAt: time.Now(), RecepientID: to, State: DIALOG_PENDING_UNREAD_STATE}
	msg_id, err := req.dbAddDialogMessage(ctx)
	if err != nil {
		return "", err
	}
	return msg_id, err
}

func DialogListDB(ctx context.Context, userID, to string) ([]SendRequest, error) {
	dialog, err := dbGetDialogWithState(ctx, userID, to, []string{DIALOG_UNREAD_STATE, DIALOG_READ_STATE})
	if err != nil {
		return nil, err
	}
	return dialog, nil
}

func DialogListUnreadDB(ctx context.Context, userID, to string) ([]SendRequest, error) {
	dialog, err := dbGetDialogWithState(ctx, userID, to, []string{DIALOG_UNREAD_STATE})
	if err != nil {
		return nil, err
	}
	return dialog, nil
}

func DialogListReadDB(ctx context.Context, userID, to string) ([]SendRequest, error) {
	dialog, err := dbGetDialogWithState(ctx, userID, to, []string{DIALOG_READ_STATE})
	if err != nil {
		return nil, err
	}
	return dialog, nil
}
