package plan

type Workout struct {
	Exercises []Exercise
}

type Exercise struct {
	Name        string
	Description string
	Reps        int
	Weight      float32
	BodyPart    string
}
