package plan

type Workout struct {
	Exercises []Exercise
}

type Exercise struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Reps        int     `json:"reps"`
	Weight      float32 `json:"weight"`
	BodyPart    string  `json:"body_part"`
}
