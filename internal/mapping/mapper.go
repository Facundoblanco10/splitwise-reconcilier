package mapping

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/facundo/splitwise-reconcilier/internal/splitwise"
)

const Unassigned = "UNASSIGNED"

var cardTagRe = regexp.MustCompile(`\[CARD:([A-Z0-9_]+)\]`)

type MappedExpense struct {
	Expense   splitwise.Expense
	Card      string
	MyShare   float64
	TheirShare float64
}

type Mapper struct {
	cfg *Config
}

func NewMapper(cfg *Config) *Mapper {
	return &Mapper{cfg: cfg}
}

func (m *Mapper) MapAll(expenses []splitwise.Expense) []MappedExpense {
	result := make([]MappedExpense, 0, len(expenses))
	for _, e := range expenses {
		result = append(result, m.mapOne(e))
	}
	return result
}

func (m *Mapper) mapOne(e splitwise.Expense) MappedExpense {
	card := m.resolveCard(e)
	myShare, theirShare := m.resolveShares(e)
	return MappedExpense{
		Expense:    e,
		Card:       card,
		MyShare:    myShare,
		TheirShare: theirShare,
	}
}

func (m *Mapper) resolveCard(e splitwise.Expense) string {
	// Priority 1: explicit override by expense ID
	if cardID, ok := m.cfg.OverridesByExpenseID[e.ID]; ok {
		return cardID
	}

	// Priority 2: [CARD:XXX] tag in details field
	if matches := cardTagRe.FindStringSubmatch(e.Details); len(matches) == 2 {
		return matches[1]
	}

	// Priority 3: rule by category name
	if cardID, ok := m.cfg.RulesByCategory[e.Category.Name]; ok {
		return cardID
	}

	return Unassigned
}

func (m *Mapper) resolveShares(e splitwise.Expense) (myShare, theirShare float64) {
	for _, u := range e.Users {
		share := parseAmount(u.OwedShare)
		if u.UserID == m.cfg.Splitwise.MyUserID {
			myShare = share
		} else {
			theirShare += share
		}
	}
	return
}

func parseAmount(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func ParseCost(s string) float64 {
	return parseAmount(s)
}

func FormatCardName(cfg *Config, cardID string) string {
	for _, c := range cfg.Cards {
		if c.ID == cardID {
			return c.Name
		}
	}
	if cardID == Unassigned {
		return "Unassigned"
	}
	return fmt.Sprintf("(%s)", cardID)
}
