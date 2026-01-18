package models

import "slices"

type TaskStatus int

const (
	StatusNew          TaskStatus = 0
	StatusReadyToStart TaskStatus = 1
	StatusProcessing   TaskStatus = 2
	StatusCompleted    TaskStatus = 3
	StatusFailed       TaskStatus = 4
)

func (s TaskStatus) OneOf(statuses ...TaskStatus) bool {
	return slices.Contains(statuses, s)
}

// LimitedList is the list of statuses that are counted towards client task limits.
func (s TaskStatus) LimitedList() []TaskStatus {
	return []TaskStatus{
		StatusNew,
		StatusReadyToStart,
		StatusProcessing,
	}
}
