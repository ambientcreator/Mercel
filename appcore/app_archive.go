package appcore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

func (a *App) SaveCalculation(req SaveCalculationRequest) (SavedCalculation, error) {
	user, err := a.requireAuth()
	if err != nil {
		return SavedCalculation{}, err
	}
	if req.TargetAmount <= 0 || len(req.Items) == 0 {
		return SavedCalculation{}, errors.New("\u041d\u0435\u043b\u044c\u0437\u044f \u0441\u043e\u0445\u0440\u0430\u043d\u0438\u0442\u044c \u043f\u0443\u0441\u0442\u043e\u0439 \u0440\u0430\u0441\u0447\u0451\u0442.")
	}
	now := time.Now()
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = archiveDateTitle(now)
	}
	total := 0
	for _, item := range req.Items {
		total += item.LineTotal
	}
	payload, err := json.Marshal(req.Items)
	if err != nil {
		return SavedCalculation{}, fmt.Errorf("marshal calculation items: %w", err)
	}
	createdAt := now.Format(time.RFC3339)

	var existingID int64
	err = a.db.QueryRow(`SELECT id FROM calculations WHERE created_by = ? AND title = ? LIMIT 1`, user.Username, title).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return SavedCalculation{}, fmt.Errorf("find daily calculation: %w", err)
	}
	if err == nil {
		_, err = a.db.Exec(`UPDATE calculations SET target_amount = ?, total_amount = ?, items_json = ?, created_at = ?, created_role = ? WHERE id = ?`, req.TargetAmount, total, string(payload), createdAt, user.Role, existingID)
		if err != nil {
			return SavedCalculation{}, fmt.Errorf("update daily calculation: %w", err)
		}
		return SavedCalculation{ID: existingID, Title: title, TargetAmount: req.TargetAmount, TotalAmount: total, Items: req.Items, CreatedAt: createdAt, CreatedBy: user.Username, CreatedRole: user.Role}, nil
	}

	result, err := a.db.Exec(`INSERT INTO calculations(title, target_amount, total_amount, items_json, created_at, created_by, created_role) VALUES(?, ?, ?, ?, ?, ?, ?)`, title, req.TargetAmount, total, string(payload), createdAt, user.Username, user.Role)
	if err != nil {
		return SavedCalculation{}, fmt.Errorf("save calculation: %w", err)
	}
	id, _ := result.LastInsertId()
	return SavedCalculation{ID: id, Title: title, TargetAmount: req.TargetAmount, TotalAmount: total, Items: req.Items, CreatedAt: createdAt, CreatedBy: user.Username, CreatedRole: user.Role}, nil
}

func (a *App) DeleteCalculation(id int64) error {
	user, err := a.requireAuth()
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("Укажите архивный расчёт для удаления.")
	}
	var createdBy string
	var createdRole string
	err = a.db.QueryRow(`SELECT created_by, created_role FROM calculations WHERE id = ? LIMIT 1`, id).Scan(&createdBy, &createdRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("Архивный расчёт не найден.")
		}
		return fmt.Errorf("load archived calculation: %w", err)
	}

	createdRole = normalizeRole(createdRole)
	canDelete := normalizeRole(user.Role) == RoleAdmin || roleCanManageUsers(user.Role) || createdBy == user.Username
	if !canDelete || !canViewArchiveRole(user, createdRole, createdBy) {
		return errors.New("Недостаточно прав для удаления этого архивного расчёта.")
	}

	result, err := a.db.Exec(`DELETE FROM calculations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete archived calculation: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return errors.New("Архивный расчёт не найден.")
	}
	return nil
}

func (a *App) UpdateCalculationAsAdmin(req UpdateSavedCalculationRequest) (SavedCalculation, error) {
	admin, err := a.requireAdmin()
	if err != nil {
		return SavedCalculation{}, err
	}
	if req.ID <= 0 {
		return SavedCalculation{}, errors.New("Укажите архивный расчёт для редактирования.")
	}
	if len(req.Items) == 0 {
		return SavedCalculation{}, errors.New("Архивный расчёт не может быть пустым.")
	}

	var current SavedCalculation
	var payload string
	err = a.db.QueryRow(`SELECT id, title, target_amount, total_amount, items_json, created_at, created_by, created_role FROM calculations WHERE id = ? LIMIT 1`, req.ID).
		Scan(&current.ID, &current.Title, &current.TargetAmount, &current.TotalAmount, &payload, &current.CreatedAt, &current.CreatedBy, &current.CreatedRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SavedCalculation{}, errors.New("Архивный расчёт не найден.")
		}
		return SavedCalculation{}, fmt.Errorf("load archived calculation: %w", err)
	}

	items := make([]CalculationItem, 0, len(req.Items))
	total := 0
	for index, item := range req.Items {
		name := sanitizeStoredText(item.Name)
		if name == "" {
			return SavedCalculation{}, fmt.Errorf("Укажите название услуги в строке %d.", index+1)
		}
		unit := normalizeUnit(item.Unit)
		if unit == "" {
			return SavedCalculation{}, fmt.Errorf("Укажите корректную единицу в строке %d.", index+1)
		}
		category := normalizeCategory(item.Category)
		rate := item.Rate
		if rate <= 0 {
			return SavedCalculation{}, fmt.Errorf("Укажите положительную стоимость в строке %d.", index+1)
		}
		quantity := item.Quantity
		if quantity < 0 {
			return SavedCalculation{}, fmt.Errorf("Количество не может быть отрицательным в строке %d.", index+1)
		}

		var allocationPercent *float64
		if item.AllocationPercent != nil {
			value := math.Round((*item.AllocationPercent)*1000) / 1000
			if value < 0 {
				return SavedCalculation{}, fmt.Errorf("Процент не может быть отрицательным в строке %d.", index+1)
			}
			allocationPercent = &value
		}

		serviceCode := strings.TrimSpace(item.ServiceCode)
		if serviceCode == "" {
			serviceCode = fmt.Sprintf("archive-%d-%d", req.ID, index+1)
		}

		lineTotal := rate * quantity
		total += lineTotal
		items = append(items, CalculationItem{
			ServiceID:         item.ServiceID,
			ServiceCode:       serviceCode,
			Name:              name,
			Unit:              unit,
			Rate:              rate,
			Quantity:          quantity,
			LineTotal:         lineTotal,
			Description:       sanitizeStoredText(item.Description),
			Weight:            clampWeight(item.Weight),
			Category:          category,
			AllocationPercent: allocationPercent,
		})
	}

	encoded, err := json.Marshal(items)
	if err != nil {
		return SavedCalculation{}, fmt.Errorf("marshal updated archive items: %w", err)
	}
	if _, err := a.db.Exec(`UPDATE calculations SET target_amount = ?, total_amount = ?, items_json = ? WHERE id = ?`, total, total, string(encoded), req.ID); err != nil {
		return SavedCalculation{}, fmt.Errorf("update archived calculation: %w", err)
	}

	current.TargetAmount = total
	current.TotalAmount = total
	current.Items = items
	current.CreatedRole = normalizeRole(current.CreatedRole)
	if current.CreatedBy == "" {
		current.CreatedBy = admin.Username
	}
	return current, nil
}

// EN: Method `CopyArchiveServicesToAdmin`.
//
// EN: What it does: CopyArchiveServicesToAdmin lets an administrator import services from a selected archived calculation into the admin-owned service list.
//
// EN: Key points: available only to the admin role; performs an upsert by service name to avoid duplicate admin services; copies rate, unit, category and per-service allocation from the archive.
func (a *App) CopyArchiveServicesToAdmin(calculationID int64) (CopyArchiveServicesResult, error) {
	admin, err := a.requireAdmin()
	if err != nil {
		return CopyArchiveServicesResult{}, err
	}
	if calculationID <= 0 {
		return CopyArchiveServicesResult{}, errors.New("Укажите архивный расчёт для копирования услуг.")
	}

	var payload string
	err = a.db.QueryRow(`SELECT items_json FROM calculations WHERE id = ?`, calculationID).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CopyArchiveServicesResult{}, errors.New("Архивный расчёт не найден.")
		}
		return CopyArchiveServicesResult{}, fmt.Errorf("load calculation for copy: %w", err)
	}

	var items []CalculationItem
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return CopyArchiveServicesResult{}, fmt.Errorf("parse calculation items for copy: %w", err)
	}
	if len(items) == 0 {
		return CopyArchiveServicesResult{}, errors.New("В архивном расчёте нет услуг для копирования.")
	}

	result := CopyArchiveServicesResult{}
	seenNames := make(map[string]struct{})
	requests := make([]UpsertServiceRequest, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		if _, exists := seenNames[name]; exists {
			continue
		}
		seenNames[name] = struct{}{}
		requests = append(requests, UpsertServiceRequest{
			Name:              name,
			Unit:              normalizeUnit(item.Unit),
			Rate:              item.Rate,
			Category:          normalizeCategory(item.Category),
			AllocationPercent: item.AllocationPercent,
		})
	}

	if len(requests) == 0 {
		return CopyArchiveServicesResult{}, errors.New("В архивном расчёте нет подходящих услуг для копирования.")
	}
	if _, err := a.db.Exec(`DELETE FROM services WHERE created_by = ?`, admin.Username); err != nil {
		return CopyArchiveServicesResult{}, fmt.Errorf("clear admin services before copy: %w", err)
	}
	for _, req := range requests {
		if _, err := a.saveServiceForOwner(req, admin.Username); err != nil {
			return CopyArchiveServicesResult{}, fmt.Errorf("copy archive service %q: %w", req.Name, err)
		}
		result.Created++
	}
	return result, nil
}

// EN: Method `ListCalculations`.
//
// EN: What it does: ListCalculations returns the archive visible to the current role, already filtered by ownership rules.
//
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListCalculations() ([]SavedCalculation, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []SavedCalculation{}, nil
	}
	rows, err := a.db.Query(`SELECT c.id, c.title, c.target_amount, c.total_amount, c.items_json, c.created_at, c.created_by, c.created_role FROM calculations c ORDER BY c.id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list calculations: %w", err)
	}
	defer rows.Close()
	result := make([]SavedCalculation, 0)
	for rows.Next() {
		var item SavedCalculation
		var payload string
		if err := rows.Scan(&item.ID, &item.Title, &item.TargetAmount, &item.TotalAmount, &payload, &item.CreatedAt, &item.CreatedBy, &item.CreatedRole); err != nil {
			return nil, fmt.Errorf("scan calculation: %w", err)
		}
		item.CreatedRole = normalizeRole(item.CreatedRole)
		if !canViewArchiveRole(user, item.CreatedRole, item.CreatedBy) {
			continue
		}
		item.Title = sanitizeStoredText(item.Title)
		if item.Title == "" {
			item.Title = archiveDateTitle(time.Now())
		}
		if err := json.Unmarshal([]byte(payload), &item.Items); err != nil {
			return nil, fmt.Errorf("parse calculation items: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
