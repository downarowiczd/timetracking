package domain

import "time"

/**
 * Project status:
 * 0 - active
 * 1 - inactive
**/

type Project struct {
	Tag    string
	Name   string
	Type   string
	Status int
}

/**
 * Recording status:
 * 0 - active
 * 1 - inactive
**/
type Recording struct {
	ID         int64
	ProjectTag string
	StartTime  time.Time
	EndTime    time.Time
	Name       string
	Billable   bool
	Note       string
	Status     int
}
