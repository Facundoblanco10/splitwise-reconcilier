package mapping

import (
	"testing"

	"github.com/facundo/splitwise-reconcilier/internal/splitwise"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func mkExpense(details string, users ...splitwise.ExpenseUser) splitwise.Expense {
	return splitwise.Expense{ID: 1, Cost: "1000.00", Details: details, Users: users}
}

func mkUser(id int, paid, owed string) splitwise.ExpenseUser {
	return splitwise.ExpenseUser{UserID: id, PaidShare: paid, OwedShare: owed}
}

func newTestMapper(myID int, members []Member, names map[int]string) *Mapper {
	cfg := &Config{}
	cfg.Splitwise.MyUserID = myID
	cfg.Members = members
	return NewMapper(cfg, names)
}

// ── resolveCard ───────────────────────────────────────────────────────────────

func TestResolveCard_Tag(t *testing.T) {
	names := map[int]string{1: "John", 2: "Jane"}
	members := []Member{
		{SplitwiseID: 1, DefaultEntity: "VISA"},
		{SplitwiseID: 2, DefaultEntity: "AMEX"},
	}
	m := newTestMapper(1, members, names)

	tests := []struct {
		name    string
		expense splitwise.Expense
		want    string
	}{
		{
			name:    "explicit ENTITY:PERSON tag",
			expense: mkExpense("[SCOTIA:FATI]", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00")),
			want:    "SCOTIA:FATI",
		},
		{
			name:    "entity is uppercased, person preserved",
			expense: mkExpense("[scotia:Fati]", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00")),
			want:    "SCOTIA:Fati",
		},
		{
			name:    "hyphen in entity",
			expense: mkExpense("[ITAU-D:John]", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00")),
			want:    "ITAU-D:John",
		},
		{
			name:    "lowercase tag with hyphen auto-fills payer",
			expense: mkExpense("[itau-d]", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00")),
			want:    "ITAU-D:John",
		},
		{
			name:    "tag without colon auto-fills payer",
			expense: mkExpense("[SCOTIA]", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00")),
			want:    "SCOTIA:John",
		},
		{
			name:    "tag with empty person auto-fills payer",
			expense: mkExpense("[SCOTIA:]", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00")),
			want:    "SCOTIA:John",
		},
		{
			name:    "tag auto-fills from second user when they paid",
			expense: mkExpense("[AMEX]", mkUser(2, "1000.00", "500.00"), mkUser(1, "0.00", "500.00")),
			want:    "AMEX:Jane",
		},
		{
			name:    "tag mid-sentence is found",
			expense: mkExpense("monthly bill [AMEX:Jane] installment", mkUser(1, "1000.00", "1000.00")),
			want:    "AMEX:Jane",
		},
		{
			name:    "first tag wins when multiple present",
			expense: mkExpense("[VISA:A] [AMEX:B]", mkUser(1, "1000.00", "1000.00")),
			want:    "VISA:A",
		},
		{
			name:    "tag with unknown payer returns entity only",
			expense: mkExpense("[SCOTIA]", mkUser(99, "1000.00", "1000.00")),
			want:    "SCOTIA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.resolveCard(tt.expense)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveCard_DefaultEntity(t *testing.T) {
	names := map[int]string{1: "John", 2: "Jane"}
	members := []Member{
		{SplitwiseID: 1, DefaultEntity: "VISA"},
		{SplitwiseID: 2, DefaultEntity: "AMEX"},
	}
	m := newTestMapper(1, members, names)

	t.Run("payer 1 uses their default entity", func(t *testing.T) {
		e := mkExpense("", mkUser(1, "1000.00", "500.00"), mkUser(2, "0.00", "500.00"))
		if got := m.resolveCard(e); got != "VISA:John" {
			t.Errorf("got %q, want VISA:John", got)
		}
	})

	t.Run("payer 2 uses their default entity", func(t *testing.T) {
		e := mkExpense("", mkUser(2, "1000.00", "500.00"), mkUser(1, "0.00", "500.00"))
		if got := m.resolveCard(e); got != "AMEX:Jane" {
			t.Errorf("got %q, want AMEX:Jane", got)
		}
	})
}

func TestResolveCard_Unassigned(t *testing.T) {
	names := map[int]string{1: "John"}

	t.Run("payer not in members", func(t *testing.T) {
		m := newTestMapper(1, []Member{{SplitwiseID: 1, DefaultEntity: "VISA"}}, names)
		e := mkExpense("", mkUser(99, "1000.00", "1000.00"))
		if got := m.resolveCard(e); got != Unassigned {
			t.Errorf("got %q, want %s", got, Unassigned)
		}
	})

	t.Run("payer has no default entity", func(t *testing.T) {
		m := newTestMapper(1, []Member{{SplitwiseID: 1}}, names)
		e := mkExpense("", mkUser(1, "1000.00", "1000.00"))
		if got := m.resolveCard(e); got != Unassigned {
			t.Errorf("got %q, want %s", got, Unassigned)
		}
	})

	t.Run("no users", func(t *testing.T) {
		m := newTestMapper(1, []Member{{SplitwiseID: 1, DefaultEntity: "VISA"}}, names)
		e := mkExpense("")
		if got := m.resolveCard(e); got != Unassigned {
			t.Errorf("got %q, want %s", got, Unassigned)
		}
	})
}

// ── resolveShares ─────────────────────────────────────────────────────────────

func TestResolveShares(t *testing.T) {
	m := newTestMapper(1, nil, nil)

	tests := []struct {
		name       string
		expense    splitwise.Expense
		wantMy     float64
		wantTheir  float64
	}{
		{
			name:      "two-way split",
			expense:   mkExpense("", mkUser(1, "1000.00", "600.00"), mkUser(2, "0.00", "400.00")),
			wantMy:    600.00,
			wantTheir: 400.00,
		},
		{
			name:      "multiple others are summed",
			expense:   mkExpense("", mkUser(1, "1000.00", "400.00"), mkUser(2, "0.00", "300.00"), mkUser(3, "0.00", "300.00")),
			wantMy:    400.00,
			wantTheir: 600.00,
		},
		{
			name:      "only me",
			expense:   mkExpense("", mkUser(1, "1000.00", "1000.00")),
			wantMy:    1000.00,
			wantTheir: 0.00,
		},
		{
			name:      "I paid nothing",
			expense:   mkExpense("", mkUser(2, "1000.00", "600.00"), mkUser(1, "0.00", "400.00")),
			wantMy:    400.00,
			wantTheir: 600.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			my, their := m.resolveShares(tt.expense)
			if my != tt.wantMy {
				t.Errorf("myShare = %.2f, want %.2f", my, tt.wantMy)
			}
			if their != tt.wantTheir {
				t.Errorf("theirShare = %.2f, want %.2f", their, tt.wantTheir)
			}
		})
	}
}

// ── MapAll ────────────────────────────────────────────────────────────────────

func TestMapAll(t *testing.T) {
	names := map[int]string{1: "John", 2: "Jane"}
	members := []Member{{SplitwiseID: 1, DefaultEntity: "VISA"}}
	m := newTestMapper(1, members, names)

	expenses := []splitwise.Expense{
		mkExpense("[AMEX:Jane]", mkUser(1, "500.00", "250.00"), mkUser(2, "0.00", "250.00")),
		mkExpense("",            mkUser(1, "800.00", "800.00")),
	}

	mapped := m.MapAll(expenses)

	if len(mapped) != 2 {
		t.Fatalf("MapAll returned %d results, want 2", len(mapped))
	}
	if mapped[0].Card != "AMEX:Jane" {
		t.Errorf("mapped[0].Card = %q, want AMEX:Jane", mapped[0].Card)
	}
	if mapped[1].Card != "VISA:John" {
		t.Errorf("mapped[1].Card = %q, want VISA:John", mapped[1].Card)
	}
	if mapped[0].MyShare != 250.00 {
		t.Errorf("mapped[0].MyShare = %.2f, want 250.00", mapped[0].MyShare)
	}
}

// ── FormatCardName ────────────────────────────────────────────────────────────

func TestFormatCardName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{Unassigned, "Unassigned"},
		{"SCOTIA:John", "SCOTIA:John"},
		{"AMEX", "AMEX"},
	}
	for _, tt := range tests {
		if got := FormatCardName(tt.input); got != tt.want {
			t.Errorf("FormatCardName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── ParseCost ─────────────────────────────────────────────────────────────────

func TestParseCost(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"5000.00", 5000.00},
		{"0.00", 0.00},
		{"", 0.00},
		{"abc", 0.00},
		{"-100.50", -100.50},
	}
	for _, tt := range tests {
		if got := ParseCost(tt.input); got != tt.want {
			t.Errorf("ParseCost(%q) = %.2f, want %.2f", tt.input, got, tt.want)
		}
	}
}
