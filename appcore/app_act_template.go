package appcore

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
)

// EN: Act blank identifiers.
//
// EN: What it does: every employee is permanently bound to one of the printable act blanks, so two employees
// EN: from the same department hand in visually different documents for the very same calculation.
//
// EN: Key points: "standard" is the blank Mercel shipped from the start, the numbered ones repeat the source file
// EN: numbering (Акт_бланк_1, Акт_бланк_3); values are stored in users.act_template and must stay stable, because
// EN: changing them would silently reassign existing employees.
const (
	ActTemplateStandard    = "standard"
	ActTemplateFormal      = "blank1"
	ActTemplateTypographic = "blank3"
	ActTemplateContract    = "blank4"
	ActTemplateTabular     = "blank5"
)

// EN: Variable `legacyActTemplateIDs`.
//
// EN: What it does: legacyActTemplateIDs maps the first generation of identifiers onto the current ones.
//
// EN: Key points: the very first version stored "1" for the standard blank and "3" for the typographic one; those
// EN: rows are rewritten by the migration, and the mapping here keeps any stale value working in the meantime.
var legacyActTemplateIDs = map[string]string{
	"1": ActTemplateStandard,
	"3": ActTemplateTypographic,
}

// EN: Variable `actTemplateIDs`.
//
// EN: What it does: actTemplateIDs lists every blank that can be handed out to an employee.
//
// EN: Key points: the random assignment picks from this slice, so adding a new blank here is enough to include it
// EN: in the rotation; the order is not significant.
var actTemplateIDs = []string{ActTemplateStandard, ActTemplateFormal, ActTemplateTypographic, ActTemplateContract, ActTemplateTabular}

// EN: Function `actTemplateLabel`.
//
// EN: What it does: actTemplateLabel returns the human readable name of a blank for the UI and error messages.
//
// EN: Key points: unknown identifiers fall back to the classic label, so the caller never renders an empty string.
func actTemplateLabel(id string) string {
	switch normalizeActTemplate(id) {
	case ActTemplateFormal:
		return "Бланк №1 — типовой"
	case ActTemplateTypographic:
		return "Бланк №3 — типографский"
	case ActTemplateContract:
		return "Бланк №4 — договорный"
	case ActTemplateTabular:
		return "Бланк №5 — табличный"
	default:
		return "Стандартный бланк"
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
	if mapped, ok := legacyActTemplateIDs[trimmed]; ok {
		trimmed = mapped
	}
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
	return ActTemplateStandard
}

// EN: Function `ActTemplateOptions`.
//
// EN: What it does: ActTemplateOptions lists every blank with its label so the settings screen can offer a choice.
//
// EN: Key points: the order matches actTemplateIDs, which is also the order used by the random assignment.
func ActTemplateOptions() []ActTemplateOption {
	options := make([]ActTemplateOption, 0, len(actTemplateIDs))
	for _, id := range actTemplateIDs {
		options = append(options, ActTemplateOption{ID: id, Label: actTemplateLabel(id)})
	}
	return options
}

// EN: Method `ListActTemplates`.
//
// EN: What it does: ListActTemplates exposes the available act blanks to the frontend.
//
// EN: Key points: the catalog is static, so it needs no authorization and never fails.
func (a *App) ListActTemplates() []ActTemplateOption {
	return ActTemplateOptions()
}

// EN: Function `randomActTemplate`.
//
// EN: What it does: randomActTemplate picks one blank with equal probability.
//
// EN: Key points: used as the fallback when the current distribution cannot be read from the database.
func randomActTemplate() string {
	return actTemplateIDs[rand.IntN(len(actTemplateIDs))]
}

// EN: Method `nextActTemplate`.
//
// EN: What it does: nextActTemplate picks the blank for a new employee, keeping the blanks evenly distributed.
//
// EN: Key points: a plain random draw makes one blank dominate on small teams, so the least used blank wins and ties
// EN: are broken randomly; with N blanks the accounts end up split as evenly as N allows.
func (a *App) nextActTemplate() string {
	counts := make(map[string]int, len(actTemplateIDs))
	for _, id := range actTemplateIDs {
		counts[id] = 0
	}

	rows, err := a.db.Query(`SELECT act_template, COUNT(*) FROM users GROUP BY act_template`)
	if err != nil {
		return randomActTemplate()
	}
	defer rows.Close()
	for rows.Next() {
		var template string
		var count int
		if err := rows.Scan(&template, &count); err != nil {
			return randomActTemplate()
		}
		if id := normalizeActTemplate(template); id != "" {
			counts[id] += count
		}
	}
	if err := rows.Err(); err != nil {
		return randomActTemplate()
	}

	least := -1
	candidates := make([]string, 0, len(actTemplateIDs))
	for _, id := range actTemplateIDs {
		switch {
		case least < 0 || counts[id] < least:
			least = counts[id]
			candidates = append(candidates[:0], id)
		case counts[id] == least:
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return randomActTemplate()
	}
	return candidates[rand.IntN(len(candidates))]
}

// EN: Method `UpdateUserActTemplate`.
//
// EN: What it does: UpdateUserActTemplate switches the act blank of a user to a manually chosen one.
//
// EN: Key points: everyone may change their own blank; changing somebody else's blank follows the same rules as the
// EN: other user edits, so an administrator can fix any account.
func (a *App) UpdateUserActTemplate(req UpdateUserActTemplateRequest) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	template := normalizeActTemplate(req.ActTemplate)
	if template == "" {
		return User{}, errors.New("Выберите бланк акта из списка.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)

	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("Недостаточно прав для изменения бланка акта.")
	}

	if _, err := a.db.Exec(`UPDATE users SET act_template = ? WHERE id = ?`, template, req.UserID); err != nil {
		return User{}, fmt.Errorf("update act template: %w", err)
	}

	updated, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, err
	}
	a.mu.Lock()
	if a.currentSession != nil && a.currentSession.ID == updated.ID {
		a.currentSession.ActTemplate = updated.ActTemplate
	}
	a.mu.Unlock()
	return updated, nil
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
	assigned := a.nextActTemplate()
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
		// The counts are re-read on every step, so the backfilled accounts land on
		// the least used blank one by one instead of piling up on a single one.
		if _, err := a.db.Exec(`UPDATE users SET act_template = ? WHERE id = ?`, a.nextActTemplate(), id); err != nil {
			return fmt.Errorf("backfill act template: %w", err)
		}
	}
	return nil
}
