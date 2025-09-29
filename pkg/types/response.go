package types

import "time"

type SendResponse struct{
	Provider       string 
    ProviderID     string   
    Status         string 
    RawResponse    []byte    
    Timestamp      time.Time
}