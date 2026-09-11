package workoutplan

import (
	"testing"

	"github.com/mmk31585/workout-tracker/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestCreateWorkoutPlanRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateWorkoutPlanRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateWorkoutPlanRequest{
				Title:       "Valid Title",
				Description: "Valid Description",
				Items: []CreateWorkoutPlanItemRequest{
					{
						ExerciseID: "123e4567-e89b-12d3-a456-426614174000",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: CreateWorkoutPlanRequest{
				Title:       "",
				Description: "Valid Description",
				Items: []CreateWorkoutPlanItemRequest{
					{
						ExerciseID: "123e4567-e89b-12d3-a456-426614174000",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "title too long",
			req: CreateWorkoutPlanRequest{
				Title:       "This title is way too long and exceeds the maximum allowed length of one hundred fifty characters by quite a bit more text to make it longer and even more characters",
				Description: "Valid Description",
				Items: []CreateWorkoutPlanItemRequest{
					{
						ExerciseID: "123e4567-e89b-12d3-a456-426614174000",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid exercise id",
			req: CreateWorkoutPlanRequest{
				Title:       "Valid Title",
				Description: "Valid Description",
				Items: []CreateWorkoutPlanItemRequest{
					{
						ExerciseID: "invalid-uuid",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid unit",
			req: CreateWorkoutPlanRequest{
				Title:       "Valid Title",
				Description: "Valid Description",
				Items: []CreateWorkoutPlanItemRequest{
					{
						ExerciseID: "123e4567-e89b-12d3-a456-426614174000",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "lbs", // invalid unit
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.Validate.Struct(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateWorkoutPlanRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     UpdateWorkoutPlanRequest
		wantErr bool
	}{
		{
			name: "valid partial update",
			req: UpdateWorkoutPlanRequest{
				Title:       strPtr("New Title"),
				Items: []UpdateWorkoutPlanItemRequest{
					{
						ID:         "123e4567-e89b-12d3-a456-426614174000",
						ExerciseID: "123e4567-e89b-12d3-a456-426614174001",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid full update",
			req: UpdateWorkoutPlanRequest{
				Title:       strPtr("New Title"),
				Description: strPtr("New Description"),
				Status:      strPtr("archived"),
				Items: []UpdateWorkoutPlanItemRequest{
					{
						ID:         "123e4567-e89b-12d3-a456-426614174000",
						ExerciseID: "123e4567-e89b-12d3-a456-426614174001",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid status",
			req: UpdateWorkoutPlanRequest{
				Status: strPtr("invalid-status"),
			},
			wantErr: true,
		},
		{
			name: "invalid item id",
			req: UpdateWorkoutPlanRequest{
				Items: []UpdateWorkoutPlanItemRequest{
					{
						ID:         "invalid-uuid",
						ExerciseID: "123e4567-e89b-12d3-a456-426614174001",
						OrderIndex: 0,
						Sets:       3,
						Reps:       10,
						Weight:     50.0,
						Unit:       "kg",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.Validate.Struct(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func strPtr(s string) *string { return &s }