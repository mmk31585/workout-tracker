package scheduledworkout

import "errors"

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
	for _, target := range legalTransitions[s] {
		if target == next {
			return true
		}
	}
	return false
}
