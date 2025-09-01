Initially a simple retry logic was implemented for

```go
func SendEmailWithRetry()(){
    for attempt := 1; attempt <= maxRetries; attempt++ {
        err:= mailService.Send(mail)
        if err == nil {
            return nil
        }
        waitTime := time.Duration(attempt*2) * time.Second
        time.Sleep(waitTime)
    }
    err := fmt.Errorf("SendMail failed after %d retries", maxRetries)
}

```

But this is not Ideal since it can cause `Thundering Herd` problem.

The "thundering herd" problem happens when a large number of processes or threads are all waiting for the same event (usually I/O related), and when that event happens, they all wake up at once.

Imagine this function is running inside a worker pool or email dispatch service. Now picture this scenario:

SendGrid is down temporarily (network issue, API outage, or rate limit).

Hundreds (or thousands) of workers/processes call SendEmailWithRetry around the same time.

Each worker:

Fails immediately on Send.

Sleeps for 2s, then all wake up at once (because they retried around the same time).

They all retry together SendGrid gets hammered with requests simultaneously.

Most fail again, repeat cycle you get a retry storm (classic thundering herd).

To fix this we use exponential backoff with with jitter.

Jitter here means adding timedelays.

```go
    baseTime := 1 * time.Second
    backoffDelay := baseTime * time.Duration(1<<(attempt-1))
    jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
    waitTime := backoffDelay * jitter
```

Why Exponential Backoff?

Reduce load on a failing service
Better resource usage on your side
Plays well with rate limits
