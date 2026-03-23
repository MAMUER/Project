package training

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	pb "healthfit-platform/proto/gen/ml_generation"
)

type ProgramGenerator struct {
	mlClient pb.MLGenerationClient
}

func NewProgramGenerator(client pb.MLGenerationClient) *ProgramGenerator {
	return &ProgramGenerator{mlClient: client}
}

type GenerateProgramInput struct {
	UserID            string
	TrainingClass     string
	Contraindications []string
	Goals             []string
	FitnessLevel      string
	AgeGroup          string
	Gender            string
	HasInjury         bool
	DurationWeeks     int
}

type Program struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Weeks     []Week    `json:"weeks"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
}

type Week struct {
	WeekNumber int       `json:"week_number"`
	Schedule   []Workout `json:"schedule"`
	Notes      []string  `json:"notes"`
}

type Workout struct {
	Day             int    `json:"day"`
	WorkoutType     string `json:"workout_type"`
	Intensity       string `json:"intensity"`
	DurationMinutes int    `json:"duration_minutes"`
}

func (g *ProgramGenerator) Generate(ctx context.Context, input GenerateProgramInput) (*Program, error) {
	req := &pb.GenerateRequest{
		TrainingClass: input.TrainingClass,
		UserConditions: &pb.UserConditions{
			Contraindications: input.Contraindications,
			Goals:             input.Goals,
			FitnessLevel:      input.FitnessLevel,
			AgeGroup:          input.AgeGroup,
			Gender:            input.Gender,
			HasInjury:         input.HasInjury,
		},
		DurationWeeks: int32(input.DurationWeeks),
	}

	resp, err := g.mlClient.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	var weeks []Week
	for _, w := range resp.Program {
		var schedule []Workout
		for _, d := range w.Schedule {
			schedule = append(schedule, Workout{
				Day:             int(d.Day),
				WorkoutType:     d.WorkoutType,
				Intensity:       d.Intensity,
				DurationMinutes: int(d.DurationMinutes),
			})
		}
		weeks = append(weeks, Week{
			WeekNumber: int(w.Week),
			Schedule:   schedule,
			Notes:      w.Notes,
		})
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, input.DurationWeeks*7)

	return &Program{
		ID:        uuid.New().String(),
		UserID:    input.UserID,
		Weeks:     weeks,
		StartDate: startDate.Format(time.RFC3339),
		EndDate:   endDate.Format(time.RFC3339),
		CreatedAt: time.Now(),
	}, nil
}

func (g *ProgramGenerator) Serialize(program *Program) ([]byte, error) {
	return json.Marshal(program)
}
