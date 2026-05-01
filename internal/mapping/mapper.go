package mapping

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/facundo/splitwise-reconcilier/internal/splitwise"
)

const Unassigned = "UNASSIGNED"

// cardTagRe matches [ENTITY] or [ENTITY:PERSON] tags in the details field.
// Group 1 = entity (e.g. SCOTIA), group 2 = person (e.g. FATI — absent or empty triggers auto-fill).
var cardTagRe = regexp.MustCompile(`\[([A-Z0-9_]+)(?::([^\]]*))?\]`)

type MappedExpense struct {
	Expense    splitwise.Expense
	Card       string
	MyShare    float64
	TheirShare float64
}

type Mapper struct {
	cfg            *Config
	userFirstNames map[int]string // splitwise user ID → first name
}

func NewMapper(cfg *Config, userFirstNames map[int]string) *Mapper {
	return &Mapper{cfg: cfg, userFirstNames: userFirstNames}
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
	// Priority 1: [ENTITY:PERSON] tag in the details (notes) field.
	// If the person segment is empty, auto-fill with the payer's first name.
	for _, matches := range cardTagRe.FindAllStringSubmatch(e.Details, -1) {
		entity := matches[1]
		person := strings.TrimSpace(matches[2])
		if person == "" {
			person = m.payerFirstName(e)
		}
		if person != "" {
			return entity + ":" + person
		}
		return entity
	}

	// Priority 2: payer's default entity from member config.
	for _, u := range e.Users {
		if parseAmount(u.PaidShare) > 0 {
			for _, member := range m.cfg.Members {
				if member.SplitwiseID == u.UserID && member.DefaultEntity != "" {
					if name, ok := m.userFirstNames[u.UserID]; ok {
						return member.DefaultEntity + ":" + name
					}
				}
			}
		}
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

// payerFirstName returns the first name of the user with paid_share > 0.
// In split-payment expenses (multiple payers), returns the first match.
func (m *Mapper) payerFirstName(e splitwise.Expense) string {
	for _, u := range e.Users {
		if parseAmount(u.PaidShare) > 0 {
			if name, ok := m.userFirstNames[u.UserID]; ok {
				return name
			}
		}
	}
	return ""
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

// FormatCardName returns a human-readable label for a card key (ENTITY:PERSON).
func FormatCardName(cardID string) string {
	if cardID == Unassigned {
		return "Unassigned"
	}
	return cardID
}
