package entities

// User represents a user in the system
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // ADMIN, MANAGER, MEMBER
}
