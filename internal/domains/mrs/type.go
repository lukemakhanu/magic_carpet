package mrs

type Mrs struct {
	MrID          string
	RoundNumberID string
	TotalGoals    string
	GoalCount     string
	StartTime     string
	CompetitionID string
	Created       string
	Modified      string
}

// TotalGoalCount
type TotalGoalCount struct {
	RoundNumberID string `json:"round_number_id"`
	TotalGoals    string `json:"total_goals"`
	GoalCount     string `json:"goal_count"`
	StartTime     string `json:"start_time"`
	CompetitionID string `json:"competition_id"`
}
