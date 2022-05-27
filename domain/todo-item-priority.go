package domain

// Function used in TodoItem initialisation to set PriorityLevel
type PriorityOption func(itm *TodoItem)

type PriorityLevel int

const (
	None PriorityLevel = iota
	Low
	Medium
	High
	DateBased
)
