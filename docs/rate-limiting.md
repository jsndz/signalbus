# approach used is sliding window log 

you have timeframe of 60 seconds and you have to remove all the requests that have timestamp 
Limit is given 
And count of request is calculated for a client
and if it is more than the count then dont allow it 


and the algo goes like this:
    i=now;         or  now -60sec
    j=now+60sec        now

    while(application is running){
        while(there are no request less than timestamp i for client ):
            remove all client request from the store that are out window
        count the request
        if more than limit:
            dont allow the request 

        add the request to the  store
        allow the request to pass through
        
        set up expire date ideal users
    }

there is a sliding window here the  i and j and window slides at fix rate 


```go
func NotificationMiddleware(cfg *MiddlewareConfig) gin.HandlerFunc {
	// limit is 10 requests per minute per IP
	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", clientIP)
		now := time.Now().Unix() // this is j
		window := int64(60)
		limit := int64(10)
		// this acts a loop to remove old entries of client out of time frame and count new ones atomically
		cfg.RedisClient.ZRemRangeByScore(
			ctx,
			key,
			"0",
			fmt.Sprintf("%d", now-window), // this is the i
		)
		count, err := cfg.RedisClient.ZCard(ctx, key).Result()
		if err != nil {
			ctx.Next()
			return
		}
		if count >= limit {
			metrics.HttpRateLimitRejectionsTotal.Add(1)
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		id := uuid.New().String()
		member := fmt.Sprintf("%d:%s", now, id)
		cfg.RedisClient.ZAdd(ctx, key, redis.Z{
			Score:  float64(now),
			Member: member,
		})

		// 5. Set expiration so idle users are cleaned up
		cfg.RedisClient.Expire(ctx, key, time.Minute)

		ctx.Next()
	}
}


```