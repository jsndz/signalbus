package gosms

type Sender interface {
	Send (SMS) error
}


type SMS struct{
	To string `json:"to"`
	Text string    	`json:"text,omitempty"`
	IdempotencyKey string
}

type SMSOption func(*SMS) 

func NewSMS(from string,to string,text string, opts... SMSOption) SMS{
	s:= SMS{
		To:to,
		Text: text,
	}
	for _,opt := range opts{
		opt(&s)
	}
	return s
}

func WithIdempotencyKey(key string) (SMSOption) {
	return func(s *SMS) {
		s.IdempotencyKey = key
	}
}