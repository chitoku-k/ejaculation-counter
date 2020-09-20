package service

type Event interface {
	Name() string
}

type ReplyEvent struct {
	InReplyToID string `json:"in_reply_to_id"`
	Acct        string `json:"acct"`
	Body        string `json:"body"`
	Visibility  string `json:"visibility"`
}

func (*ReplyEvent) Name() string {
	return "events.reply"
}

type ReplyErrorEvent struct {
	InReplyToID string `json:"in_reply_to_id"`
	Acct        string `json:"acct"`
	Visibility  string `json:"visibility"`
	ActionName  string `json:"action_name"`
}

func (*ReplyErrorEvent) Name() string {
	return "events.reply_error"
}

type UpdateEvent struct {
	UserID string `json:"user_id"`
	Date   string `json:"date"`
}

func (*UpdateEvent) Name() string {
	return "events.update"
}

type IncrementEvent struct {
	UserID string `json:"user_id"`
	Date   string `json:"date"`
}

func (*IncrementEvent) Name() string {
	return "events.increment"
}

type AdministrationEvent struct {
	InReplyToID string `json:"in_reply_to_id"`
	Acct        string `json:"acct"`
	Type        string `json:"type"`
	Command     string `json:"command"`
}

func (*AdministrationEvent) Name() string {
	return "events.administration"
}
