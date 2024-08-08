package commons

import "time"

const (
	AllowedCurrencyLength       = 5
	MinimumCurrencyLength       = 3
	UserContextKey              = "user"
	AllowedRPS                  = 10
	ExternalClientMaxRetries    = 3
	ExternalClientBaseDelay     = time.Second
	ExternalClientMaxDelay      = 30 * time.Second
	RateUpdaterCacheExipiration = 1 * time.Hour
	ServerIdleTimeout           = time.Minute
	ServerReadTimeout           = 10 * time.Second
	ServerWriteTimeout          = 30 * time.Second
	CacheExpiration             = 1 * time.Hour
)
