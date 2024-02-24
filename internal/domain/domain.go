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

func (p *Project) StatusString() string {
	if p.Status == 0 {
		return "active"
	} else {
		return "inactive"
	}
}

func (r *Recording) StatusString() string {
	if r.Status == 0 {
		return "active"
	} else {
		return "inactive"
	}
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
