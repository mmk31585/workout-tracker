package scheduledworkout

import "testing"

func TestCanTransitionTo_LegalTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ScheduleStatus
		to   ScheduleStatus
	}{
		{"scheduled to in_progress", ScheduleScheduled, ScheduleInProgress},
		{"scheduled to cancelled", ScheduleScheduled, ScheduleCancelled},
		{"in_progress to completed", ScheduleInProgress, ScheduleCompleted},
		{"in_progress to cancelled", ScheduleInProgress, ScheduleCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.from.CanTransitionTo(tt.to) {
				t.Errorf("expected %q -> %q to be legal", tt.from, tt.to)
			}
		})
	}
}

func TestCanTransitionTo_IllegalTransitions(t *testing.T) {
	tests := []struct {
		name string
		from ScheduleStatus
		to   ScheduleStatus
	}{
		{"scheduled to scheduled (self)", ScheduleScheduled, ScheduleScheduled},
		{"scheduled to completed (direct)", ScheduleScheduled, ScheduleCompleted},
		{"in_progress to scheduled (backwards)", ScheduleInProgress, ScheduleScheduled},
		{"in_progress to in_progress (self)", ScheduleInProgress, ScheduleInProgress},
		{"completed to scheduled", ScheduleCompleted, ScheduleScheduled},
		{"completed to in_progress", ScheduleCompleted, ScheduleInProgress},
		{"completed to completed (repeat complete)", ScheduleCompleted, ScheduleCompleted},
		{"completed to cancelled", ScheduleCompleted, ScheduleCancelled},
		{"cancelled to scheduled", ScheduleCancelled, ScheduleScheduled},
		{"cancelled to in_progress", ScheduleCancelled, ScheduleInProgress},
		{"cancelled to completed", ScheduleCancelled, ScheduleCompleted},
		{"cancelled to cancelled (self)", ScheduleCancelled, ScheduleCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.from.CanTransitionTo(tt.to) {
				t.Errorf("expected %q -> %q to be illegal", tt.from, tt.to)
			}
		})
	}
}

func TestCanTransitionTo_UnknownStatusIsAlwaysIllegal(t *testing.T) {
	unknown := ScheduleStatus("banana")
	targets := []ScheduleStatus{
		ScheduleScheduled,
		ScheduleInProgress,
		ScheduleCompleted,
		ScheduleCancelled,
		unknown,
	}

	for _, target := range targets {
		if unknown.CanTransitionTo(target) {
			t.Errorf("expected unknown status -> %q to be illegal", target)
		}
	}
}

func TestLegalTransitions_CoversAllStatuses(t *testing.T) {
	all := []ScheduleStatus{
		ScheduleScheduled,
		ScheduleInProgress,
		ScheduleCompleted,
		ScheduleCancelled,
	}

	for _, status := range all {
		if _, ok := legalTransitions[status]; !ok {
			t.Errorf("status %q missing from legalTransitions map", status)
		}
	}
}
