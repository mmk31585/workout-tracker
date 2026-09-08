package workoutplan

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/mmk31585/workout-tracker/internal/db"
)

type WorkoutPlanService struct {
	planRepo WorkoutPlanRepository
	itemRepo WorkoutPlanItemRepository
	db       *sqlx.DB
}

func NewWorkoutPlanService(planRepo WorkoutPlanRepository, itemRepo WorkoutPlanItemRepository, db *sqlx.DB) *WorkoutPlanService {
	return &WorkoutPlanService{
		planRepo: planRepo,
		itemRepo: itemRepo,
		db:       db,
	}
}

func (s *WorkoutPlanService) CreatePlan(ctx context.Context, userID string, req CreateWorkoutPlanRequest) (WorkoutPlanResponse, error) {
	var result WorkoutPlanResponse
	err := db.RunInTx(ctx, s.db, func(tx *sqlx.Tx) error {
		// Create plan
		plan, err := s.planRepo.Create(ctx, WorkoutPlan{
			UserID:      userID,
			Title:       req.Title,
			Description: req.Description,
			Status:      "active",
		})
		if err != nil {
			return err
		}

		// Create items
		items := make([]WorkoutPlanItem, len(req.Items))
		for i, itemReq := range req.Items {
			items[i] = WorkoutPlanItem{
				WorkoutPlanID: plan.ID,
				ExerciseID:    itemReq.ExerciseID,
				OrderIndex:    itemReq.OrderIndex,
				Sets:          itemReq.Sets,
				Reps:          itemReq.Reps,
				Weight:        itemReq.Weight,
				Unit:          itemReq.Unit,
			}
		}

		// Create items in batch
		newItems, err := s.itemRepo.CreateBatch(ctx, items)
		if err != nil {
			return err
		}

		result = WorkoutPlanResponse{
			ID:          plan.ID,
			Title:       plan.Title,
			Description: plan.Description,
			Status:      plan.Status,
			Items:       toResponseItems(newItems),
			CreatedAt:   plan.CreatedAt,
			UpdatedAt:   plan.UpdatedAt,
		}
		return nil
	})
	return result, err
}

func (s *WorkoutPlanService) GetPlan(ctx context.Context, userID, planID string) (WorkoutPlanResponse, error) {
	plan, err := s.planRepo.GetByID(ctx, planID, userID)
	if err != nil {
		return WorkoutPlanResponse{}, err
	}

	items, err := s.itemRepo.GetByPlanID(ctx, planID, userID)
	if err != nil {
		return WorkoutPlanResponse{}, err
	}

	return WorkoutPlanResponse{
		ID:          plan.ID,
		Title:       plan.Title,
		Description: plan.Description,
		Status:      plan.Status,
		Items:       toResponseItems(items),
		CreatedAt:   plan.CreatedAt,
		UpdatedAt:   plan.UpdatedAt,
	}, nil
}

func (s *WorkoutPlanService) ListPlans(ctx context.Context, userID string) ([]WorkoutPlanResponse, error) {
	plans, err := s.planRepo.ListByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]WorkoutPlanResponse, len(plans))
	for i, plan := range plans {
		items, err := s.itemRepo.GetByPlanID(ctx, plan.ID, userID)
		if err != nil {
			return nil, err
		}
		responses[i] = WorkoutPlanResponse{
			ID:          plan.ID,
			Title:       plan.Title,
			Description: plan.Description,
			Status:      plan.Status,
			Items:       toResponseItems(items),
			CreatedAt:   plan.CreatedAt,
			UpdatedAt:   plan.UpdatedAt,
		}
	}
	return responses, nil
}

func (s *WorkoutPlanService) UpdatePlan(ctx context.Context, userID, planID string, req UpdateWorkoutPlanRequest) (WorkoutPlanResponse, error) {
	var result WorkoutPlanResponse
	err := db.RunInTx(ctx, s.db, func(tx *sqlx.Tx) error {

		existingPlan, err := s.planRepo.GetByID(ctx, planID, userID)
		if err != nil {
			return err
		}

		existingItems, err := s.itemRepo.GetByPlanID(ctx, planID, userID)
		if err != nil {
			return err
		}

		updatedPlan := existingPlan
		if req.Title != nil {
			updatedPlan.Title = *req.Title
		}
		if req.Description != nil {
			updatedPlan.Description = *req.Description
		}
		if req.Status != nil {
			updatedPlan.Status = *req.Status
		}

		updatedPlan, err = s.planRepo.UpdateByID(ctx, updatedPlan, userID)
		if err != nil {
			return err
		}

		if req.Items != nil {
			// Delete existing items
			itemIDs := make([]string, len(existingItems))
			for i, item := range existingItems {
				itemIDs[i] = item.ID
			}
			if err := s.itemRepo.DeleteBatch(ctx, itemIDs, userID); err != nil {
				return err
			}

			// Create new items
			items := make([]WorkoutPlanItem, len(req.Items))
			for i, itemReq := range req.Items {
				items[i] = WorkoutPlanItem{
					WorkoutPlanID: planID,
					ExerciseID:    itemReq.ExerciseID,
					OrderIndex:    itemReq.OrderIndex,
					Sets:          itemReq.Sets,
					Reps:          itemReq.Reps,
					Weight:        itemReq.Weight,
					Unit:          itemReq.Unit,
				}
			}

			newItems, err := s.itemRepo.CreateBatch(ctx, items)
			if err != nil {
				return err
			}

			result = WorkoutPlanResponse{
				ID:          updatedPlan.ID,
				Title:       updatedPlan.Title,
				Description: updatedPlan.Description,
				Status:      updatedPlan.Status,
				Items:       toResponseItems(newItems),
				CreatedAt:   updatedPlan.CreatedAt,
				UpdatedAt:   updatedPlan.UpdatedAt,
			}
		} else {
			// No items update, return existing items
			result = WorkoutPlanResponse{
				ID:          updatedPlan.ID,
				Title:       updatedPlan.Title,
				Description: updatedPlan.Description,
				Status:      updatedPlan.Status,
				Items:       toResponseItems(existingItems),
				CreatedAt:   updatedPlan.CreatedAt,
				UpdatedAt:   updatedPlan.UpdatedAt,
			}
		}
		return nil
	})
	return result, err
}

func (s *WorkoutPlanService) DeletePlan(ctx context.Context, userID, planID string) error {
	return s.planRepo.DeleteByID(ctx, planID, userID)
}

func toResponseItems(items []WorkoutPlanItem) []WorkoutPlanItemResponse {
	responses := make([]WorkoutPlanItemResponse, len(items))
	for i, item := range items {
		responses[i] = WorkoutPlanItemResponse{
			ID:         item.ID,
			ExerciseID: item.ExerciseID,
			OrderIndex: item.OrderIndex,
			Sets:       item.Sets,
			Reps:       item.Reps,
			Weight:     item.Weight,
			Unit:       item.Unit,
			CreatedAt:  item.CreatedAt,
		}
	}
	return responses
}
