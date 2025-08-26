functional options pattern

```go

type EmailOption func(*Email)

func NewEmail(from ,to string , opts ...EmailOption) Email{
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

```
