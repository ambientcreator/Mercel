package appcore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (a *App) SaveCalculation(req SaveCalculationRequest) (SavedCalculation, error) {
	user, err := a.requireAuth()
	if err != nil {
		return SavedCalculation{}, err
	}
	if req.TargetAmount <= 0 || len(req.Items) == 0 {
		return SavedCalculation{}, errors.New("Р СњР ВµР В»РЎРЉР В·РЎРЏ РЎРѓР С•РЎвЂ¦РЎР‚Р В°Р Р…Р С‘РЎвЂљРЎРЉ Р С—РЎС“РЎРѓРЎвЂљР С•Р в„– РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљ.")
	}
	now := time.Now()
	title := archiveDateTitle(now)
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
		return errors.New("Р СњР ВµР С”Р С•РЎР‚РЎР‚Р ВµР С”РЎвЂљР Р…РЎвЂ№Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљР В°.")
	}
	if normalizeRole(user.Role) == RoleAdmin || roleCanManageUsers(user.Role) {
		_, err = a.db.Exec(`DELETE FROM calculations WHERE id = ?`, id)
		return err
	}
	_, err = a.db.Exec(`DELETE FROM calculations WHERE id = ? AND created_by = ?`, id, user.Username)
	return err
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `CopyArchiveServicesToAdmin`.
// EN: Method `CopyArchiveServicesToAdmin`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р С—Р С•Р В·Р Р†Р С•Р В»РЎРЏР ВµРЎвЂљ Р В°Р Т‘Р СР С‘Р Р…Р С‘РЎРѓРЎвЂљРЎР‚Р В°РЎвЂљР С•РЎР‚РЎС“ Р Р†Р В·РЎРЏРЎвЂљРЎРЉ РЎС“РЎРѓР В»РЎС“Р С–Р С‘ Р С‘Р В· Р Р†РЎвЂ№Р В±РЎР‚Р В°Р Р…Р Р…Р С•Р С–Р С• Р В°РЎР‚РЎвЂ¦Р С‘Р Р†Р Р…Р С•Р С–Р С• РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљР В° Р С‘ РЎРѓР С”Р С•Р С—Р С‘РЎР‚Р С•Р Р†Р В°РЎвЂљРЎРЉ Р С‘РЎвЂ¦ Р Р† РЎРѓР Р†Р С•Р в„– РЎРѓР С•Р В±РЎРѓРЎвЂљР Р†Р ВµР Р…Р Р…РЎвЂ№Р в„– РЎРѓР С—Р С‘РЎРѓР С•Р С” РЎС“РЎРѓР В»РЎС“Р С–.
// EN: What it does: CopyArchiveServicesToAdmin lets an administrator import services from a selected archived calculation into the admin-owned service list.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Т‘Р С•РЎРѓРЎвЂљРЎС“Р С—Р ВµР Р… РЎвЂљР С•Р В»РЎРЉР С”Р С• РЎР‚Р С•Р В»Р С‘ admin; Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљ upsert Р С—Р С• Р С‘Р СР ВµР Р…Р С‘ РЎС“РЎРѓР В»РЎС“Р С–Р С‘, РЎвЂЎРЎвЂљР С•Р В±РЎвЂ№ Р Р…Р Вµ Р С—Р В»Р С•Р Т‘Р С‘РЎвЂљРЎРЉ Р Т‘РЎС“Р В±Р В»Р С‘Р С”Р В°РЎвЂљРЎвЂ№ РЎС“ Р В°Р Т‘Р СР С‘Р Р…Р С‘РЎРѓРЎвЂљРЎР‚Р В°РЎвЂљР С•РЎР‚Р В°; Р С”Р С•Р С—Р С‘РЎР‚РЎС“Р ВµРЎвЂљ РЎвЂљР В°РЎР‚Р С‘РЎвЂћ, Р ВµР Т‘Р С‘Р Р…Р С‘РЎвЂ РЎС“, Р С–РЎР‚РЎС“Р С—Р С—РЎС“ Р С‘ Р С‘Р Р…Р Т‘Р С‘Р Р†Р С‘Р Т‘РЎС“Р В°Р В»РЎРЉР Р…РЎвЂ№Р в„– Р С—РЎР‚Р С•РЎвЂ Р ВµР Р…РЎвЂљ РЎС“РЎРѓР В»РЎС“Р С–Р С‘ Р С‘Р В· Р В°РЎР‚РЎвЂ¦Р С‘Р Р†Р В°.
// EN: Key points: available only to the admin role; performs an upsert by service name to avoid duplicate admin services; copies rate, unit, category and per-service allocation from the archive.
func (a *App) CopyArchiveServicesToAdmin(calculationID int64) (CopyArchiveServicesResult, error) {
	admin, err := a.requireAdmin()
	if err != nil {
		return CopyArchiveServicesResult{}, err
	}
	if calculationID <= 0 {
		return CopyArchiveServicesResult{}, errors.New("Р СњР ВµР С”Р С•РЎР‚РЎР‚Р ВµР С”РЎвЂљР Р…РЎвЂ№Р в„– Р С‘Р Т‘Р ВµР Р…РЎвЂљР С‘РЎвЂћР С‘Р С”Р В°РЎвЂљР С•РЎР‚ Р В°РЎР‚РЎвЂ¦Р С‘Р Р†Р Р…Р С•Р С–Р С• РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљР В°.")
	}

	var payload string
	err = a.db.QueryRow(`SELECT items_json FROM calculations WHERE id = ?`, calculationID).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CopyArchiveServicesResult{}, errors.New("Р С’РЎР‚РЎвЂ¦Р С‘Р Р†Р Р…РЎвЂ№Р в„– РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљ Р Р…Р Вµ Р Р…Р В°Р в„–Р Т‘Р ВµР Р….")
		}
		return CopyArchiveServicesResult{}, fmt.Errorf("load calculation for copy: %w", err)
	}

	var items []CalculationItem
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return CopyArchiveServicesResult{}, fmt.Errorf("parse calculation items for copy: %w", err)
	}
	if len(items) == 0 {
		return CopyArchiveServicesResult{}, errors.New("Р вЂ™ Р В°РЎР‚РЎвЂ¦Р С‘Р Р†Р Р…Р С•Р С РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљР Вµ Р Р…Р ВµРЎвЂљ РЎС“РЎРѓР В»РЎС“Р С– Р Т‘Р В»РЎРЏ Р С”Р С•Р С—Р С‘РЎР‚Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.")
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
		return CopyArchiveServicesResult{}, errors.New("Р вЂ™ Р В°РЎР‚РЎвЂ¦Р С‘Р Р†Р Р…Р С•Р С РЎР‚Р В°РЎРѓРЎвЂЎРЎвЂРЎвЂљР Вµ Р Р…Р Вµ Р Р…Р В°Р в„–Р Т‘Р ВµР Р…Р С• РЎС“РЎРѓР В»РЎС“Р С– Р Т‘Р В»РЎРЏ Р С”Р С•Р С—Р С‘РЎР‚Р С•Р Р†Р В°Р Р…Р С‘РЎРЏ.")
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

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `ListCalculations`.
// EN: Method `ListCalculations`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎвЂЎР С‘РЎвЂљР В°Р ВµРЎвЂљ Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ РЎвЂћРЎР‚Р С•Р Р…РЎвЂљР ВµР Р…Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р ВµР Р…Р Р…Р ВµР в„– Р В»Р С•Р С–Р С‘Р С”Р С‘.
// EN: What it does: ListCalculations returns the archive visible to the current role, already filtered by ownership rules.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ Р Р† Р С•РЎвЂљРЎР‚Р С‘РЎРѓР С•Р Р†Р С”Р Вµ Р С‘Р Р…РЎвЂљР ВµРЎР‚РЎвЂћР ВµР в„–РЎРѓР В°; Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°РЎвЂљРЎРЉ РЎвЂљР С•Р В»РЎРЉР С”Р С• РЎР‚Р В°Р В·РЎР‚Р ВµРЎв‚¬РЎвЂР Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ; Р С—Р С•РЎР‚РЎРЏР Т‘Р С•Р С” Р С‘ РЎвЂћР С‘Р В»РЎРЉРЎвЂљРЎР‚Р В°РЎвЂ Р С‘РЎРЏ Р Р†Р В°Р В¶Р Р…РЎвЂ№ Р Т‘Р В»РЎРЏ UX.
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
