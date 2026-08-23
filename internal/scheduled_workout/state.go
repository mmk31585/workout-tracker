package scheduledworkout

import (
	"errors"
	"slices"
)

type ScheduleStatus string

const (
	ScheduleScheduled  ScheduleStatus = "scheduled"
	ScheduleInProgress ScheduleStatus = "in_progress"
	ScheduleCompleted  ScheduleStatus = "completed"
	ScheduleCancelled  ScheduleStatus = "cancelled"
)

var ErrIllegalScheduleTransition = errors.New("illegal schedule transition")

var legalTransitions = map[ScheduleStatus][]ScheduleStatus{
	ScheduleScheduled:  {ScheduleInProgress, ScheduleCancelled},
	ScheduleInProgress: {ScheduleCompleted, ScheduleCancelled},
	ScheduleCompleted:  {},
	ScheduleCancelled:  {},
}

func (s ScheduleStatus) CanTransitionTo(next ScheduleStatus) bool {
	return slices.Contains(legalTransitions[s], next)
}
