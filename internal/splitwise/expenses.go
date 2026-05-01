package splitwise

import (
	"fmt"
	"net/url"
	"time"
)

func (c *Client) GetExpenses(groupID int, from, to time.Time) ([]Expense, error) {
	params := url.Values{}
	params.Set("group_id", fmt.Sprintf("%d", groupID))
	params.Set("dated_after", from.Format(time.RFC3339))
	params.Set("dated_before", to.Format(time.RFC3339))
	params.Set("limit", "500")

	var result ExpensesResponse
	if err := c.get("/get_expenses?"+params.Encode(), &result); err != nil {
		return nil, err
	}

	active := make([]Expense, 0, len(result.Expenses))
	for _, e := range result.Expenses {
		if e.DeletedAt == nil {
			active = append(active, e)
		}
	}
	return active, nil
}
