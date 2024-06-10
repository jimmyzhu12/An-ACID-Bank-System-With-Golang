package branch

import "time"

type TentativeWrite struct {
	Timestamp time.Time
	Value     int
	Id        string
}

type ReadTimestamp struct {
	Timestamp time.Time
	Id        string
}

type Account struct {
	AccountName   string
	CommitedValue int
	CommitedTS    time.Time
	RTS           []ReadTimestamp
	TW            []TentativeWrite
}
