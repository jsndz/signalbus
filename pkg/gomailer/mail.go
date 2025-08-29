package gomailer

//any method that uses Send (Email) error is a Mailer interface
//so All these mailers essentially are of type Mailer
//user can decide which to choose and call mailer effectively
// this can do abstractions like m= &SMTPMailer{}
// m.send() this gives abstraction
type Mailer interface {
	Send (Email) error
}

type Email struct{
	From string 
	To []string
	Subject string 
	Text string
	HTML string
	IdempotencyKey string
	Attachments []string
	Headers map[string]string
}


type EmailOption func(*Email)

func NewEmail(from string,to []string , opts ...EmailOption) Email{
	e:= Email{
		From: from,
		To: to,
	}
	// this calls all the optional functions
	for _,opt :=range opts{
		opt(&e)
	}
	return e
}

func WithSubject(sub string) EmailOption {
	return func(e *Email) {
		e.Subject = sub
	}
}

func WithText(text string) EmailOption {
	return func(e *Email) {
		e.Text = text
	}
}

func WithHTML(html string) EmailOption {
	return func(e *Email) {
		e.HTML = html
	}
}

func WithAttachments(files ...string) EmailOption {
	return func(e *Email) {
		e.Attachments = append(e.Attachments, files...)
	}
}

func Header(key,value string) EmailOption {
	return func(e *Email) {
		if e.Headers == nil {
			e.Headers = make(map[string]string)
		}
		e.Headers[key]=value
	}
}
