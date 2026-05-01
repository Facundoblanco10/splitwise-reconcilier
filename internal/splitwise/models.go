package splitwise

type CurrentUserResponse struct {
	User User `json:"user"`
}

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type GroupsResponse struct {
	Groups []Group `json:"groups"`
}

type Group struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExpensesResponse struct {
	Expenses []Expense `json:"expenses"`
}

type Expense struct {
	ID          int            `json:"id"`
	Description string         `json:"description"`
	Cost        string         `json:"cost"`
	Date        string         `json:"date"`
	Details     string         `json:"details"`
	DeletedAt   *string        `json:"deleted_at"`
	Category    Category       `json:"category"`
	Users       []ExpenseUser  `json:"users"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExpenseUser struct {
	UserID     int    `json:"user_id"`
	PaidShare  string `json:"paid_share"`
	OwedShare  string `json:"owed_share"`
}
