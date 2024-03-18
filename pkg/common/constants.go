package common

const INCREMENT_MESSAGE_COUNT_ACTION = "increment"
const DECREMENT_MESSAGE_COUNT_ACTION = "decrement"

type MessageCountRequest struct {
	MessageID   string `pg:"id"`
	AuthorID    string `pg:"author_id"`
	RecepientID string `pg:"recepient_id"`
	Action      string `pg:"action"`
}
