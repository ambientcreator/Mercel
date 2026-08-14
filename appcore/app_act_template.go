package appcore

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// EN: Act blank identifiers.
//
// EN: What it does: every employee is permanently bound to one of the printable act blanks, so two employees
// EN: from the same department hand in visually different documents for the very same calculation.
//
// EN: Key points: identifiers repeat the source file numbering (Акт_бланк_1, Акт_бланк_3); values are stored in
// EN: users.act_template and must stay stable, because changing them would silently reassign existing employees.
const (
	ActTemplateClassic     = "1"
	ActTemplateTypographic = "3"
)

// EN: Variable `actTemplateIDs`.
//
// EN: What it does: actTemplateIDs lists every blank that can be handed out to an employee.
//
// EN: Key points: the random assignment picks from this slice, so adding a new blank here is enough to include it
// EN: in the rotation; the order is not significant.
var actTemplateIDs = []string{ActTemplateClassic, ActTemplateTypographic}

// EN: Function `actTemplateLabel`.
//
// EN: What it does: actTemplateLabel returns the human readable name of a blank for the UI and error messages.
//
// EN: Key points: unknown identifiers fall back to the classic label, so the caller never renders an empty string.
func actTemplateLabel(id string) string {
	switch normalizeActTemplate(id) {
	case ActTemplateTypographic:
		return "Бланк №3 — типографский"
	default:
		return "Бланк №1 — классический"
	}
}

// EN: Function `normalizeActTemplate`.
//
// EN: What it does: normalizeActTemplate keeps only identifiers the renderer really knows about.
//
// EN: Key points: returns an empty string for unknown or missing values, which is the signal for callers that a
// EN: blank still has to be assigned.
func normalizeActTemplate(value string) string {
	trimmed := strings.TrimSpace(value)
	for _, id := range actTemplateIDs {
		if trimmed == id {
			return trimmed
		}
	}
	return ""
}

// EN: Function `actTemplateOrDefault`.
//
// EN: What it does: actTemplateOrDefault resolves the blank used for rendering when nothing was assigned yet.
//
// EN: Key points: rendering must never fail because of a missing blank, so the classic act is the safe fallback.
func actTemplateOrDefault(value string) string {
	if id := normalizeActTemplate(value); id != "" {
		return id
	}
	return ActTemplateClassic
}

// EN: Function `randomActTemplate`.
//
// EN: What it does: randomActTemplate picks the blank a freshly created employee will use from now on.
//
// EN: Key points: the choice is made once and then persisted, so the same employee always exports the same blank.
func randomActTemplate() string {
	return actTemplateIDs[rand.IntN(len(actTemplateIDs))]
}

// EN: Method `ensureUserActTemplate`.
//
// EN: What it does: ensureUserActTemplate returns the blank of a user and assigns a random one on first use.
//
// EN: Key points: covers accounts created before this feature existed; the assignment is written back to the
// EN: database and to the in-memory session, so the blank stays the same for every later export.
func (a *App) ensureUserActTemplate(userID int64, current string) (string, error) {
	if id := normalizeActTemplate(current); id != "" {
		return id, nil
	}
	assigned := randomActTemplate()
	if _, err := a.db.Exec(`UPDATE users SET act_template = ? WHERE id = ?`, assigned, userID); err != nil {
		return "", fmt.Errorf("assign act template: %w", err)
	}
	a.mu.Lock()
	if a.currentSession != nil && a.currentSession.ID == userID {
		a.currentSession.ActTemplate = assigned
	}
	a.mu.Unlock()
	return assigned, nil
}

// EN: Method `backfillActTemplates`.
//
// EN: What it does: backfillActTemplates hands a random blank to every user row that has no valid one yet.
//
// EN: Key points: runs as part of the lightweight migrations; each row is drawn separately so existing employees
// EN: end up spread across the available blanks instead of all sharing one.
func (a *App) backfillActTemplates() error {
	rows, err := a.db.Query(`SELECT id, act_template FROM users`)
	if err != nil {
		return fmt.Errorf("read act templates: %w", err)
	}
	defer rows.Close()

	pending := make([]int64, 0)
	for rows.Next() {
		var id int64
		var template string
		if err := rows.Scan(&id, &template); err != nil {
			return fmt.Errorf("scan act template: %w", err)
		}
		if normalizeActTemplate(template) == "" {
			pending = append(pending, id)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, id := range pending {
		if _, err := a.db.Exec(`UPDATE users SET act_template = ? WHERE id = ?`, randomActTemplate(), id); err != nil {
			return fmt.Errorf("backfill act template: %w", err)
		}
	}
	return nil
}
