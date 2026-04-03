package main

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
		return SavedCalculation{}, errors.New("РќРµР»СЊР·СЏ СЃРѕС…СЂР°РЅРёС‚СЊ РїСѓСЃС‚РѕР№ СЂР°СЃС‡С‘С‚.")
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
		return errors.New("РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ СЂР°СЃС‡С‘С‚Р°.")
	}
	if normalizeRole(user.Role) == RoleAdmin || roleCanManageUsers(user.Role) {
		_, err = a.db.Exec(`DELETE FROM calculations WHERE id = ?`, id)
		return err
	}
	_, err = a.db.Exec(`DELETE FROM calculations WHERE id = ? AND created_by = ?`, id, user.Username)
	return err
}

// RU: РњРµС‚РѕРґ `CopyArchiveServicesToAdmin`.
// EN: Method `CopyArchiveServicesToAdmin`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РїРѕР·РІРѕР»СЏРµС‚ Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂСѓ РІР·СЏС‚СЊ СѓСЃР»СѓРіРё РёР· РІС‹Р±СЂР°РЅРЅРѕРіРѕ Р°СЂС…РёРІРЅРѕРіРѕ СЂР°СЃС‡С‘С‚Р° Рё СЃРєРѕРїРёСЂРѕРІР°С‚СЊ РёС… РІ СЃРІРѕР№ СЃРѕР±СЃС‚РІРµРЅРЅС‹Р№ СЃРїРёСЃРѕРє СѓСЃР»СѓРі.
// EN: What it does: CopyArchiveServicesToAdmin lets an administrator import services from a selected archived calculation into the admin-owned service list.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РґРѕСЃС‚СѓРїРµРЅ С‚РѕР»СЊРєРѕ СЂРѕР»Рё admin; РёСЃРїРѕР»СЊР·СѓРµС‚ upsert РїРѕ РёРјРµРЅРё СѓСЃР»СѓРіРё, С‡С‚РѕР±С‹ РЅРµ РїР»РѕРґРёС‚СЊ РґСѓР±Р»РёРєР°С‚С‹ Сѓ Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂР°; РєРѕРїРёСЂСѓРµС‚ С‚Р°СЂРёС„, РµРґРёРЅРёС†Сѓ, РіСЂСѓРїРїСѓ Рё РёРЅРґРёРІРёРґСѓР°Р»СЊРЅС‹Р№ РїСЂРѕС†РµРЅС‚ СѓСЃР»СѓРіРё РёР· Р°СЂС…РёРІР°.
// EN: Key points: available only to the admin role; performs an upsert by service name to avoid duplicate admin services; copies rate, unit, category and per-service allocation from the archive.
func (a *App) CopyArchiveServicesToAdmin(calculationID int64) (CopyArchiveServicesResult, error) {
	admin, err := a.requireAdmin()
	if err != nil {
		return CopyArchiveServicesResult{}, err
	}
	if calculationID <= 0 {
		return CopyArchiveServicesResult{}, errors.New("РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ Р°СЂС…РёРІРЅРѕРіРѕ СЂР°СЃС‡С‘С‚Р°.")
	}

	var payload string
	err = a.db.QueryRow(`SELECT items_json FROM calculations WHERE id = ?`, calculationID).Scan(&payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return CopyArchiveServicesResult{}, errors.New("РђСЂС…РёРІРЅС‹Р№ СЂР°СЃС‡С‘С‚ РЅРµ РЅР°Р№РґРµРЅ.")
		}
		return CopyArchiveServicesResult{}, fmt.Errorf("load calculation for copy: %w", err)
	}

	var items []CalculationItem
	if err := json.Unmarshal([]byte(payload), &items); err != nil {
		return CopyArchiveServicesResult{}, fmt.Errorf("parse calculation items for copy: %w", err)
	}
	if len(items) == 0 {
		return CopyArchiveServicesResult{}, errors.New("Р’ Р°СЂС…РёРІРЅРѕРј СЂР°СЃС‡С‘С‚Рµ РЅРµС‚ СѓСЃР»СѓРі РґР»СЏ РєРѕРїРёСЂРѕРІР°РЅРёСЏ.")
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
		return CopyArchiveServicesResult{}, errors.New("Р’ Р°СЂС…РёРІРЅРѕРј СЂР°СЃС‡С‘С‚Рµ РЅРµ РЅР°Р№РґРµРЅРѕ СѓСЃР»СѓРі РґР»СЏ РєРѕРїРёСЂРѕРІР°РЅРёСЏ.")
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

// RU: РњРµС‚РѕРґ `ListCalculations`.
// EN: Method `ListCalculations`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: С‡РёС‚Р°РµС‚ Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РґР°РЅРЅС‹Рµ РґР»СЏ С„СЂРѕРЅС‚РµРЅРґР° РёР»Рё РІРЅСѓС‚СЂРµРЅРЅРµР№ Р»РѕРіРёРєРё.
// EN: What it does: ListCalculations returns the archive visible to the current role, already filtered by ownership rules.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ РІ РѕС‚СЂРёСЃРѕРІРєРµ РёРЅС‚РµСЂС„РµР№СЃР°; РґРѕР»Р¶РµРЅ РІРѕР·РІСЂР°С‰Р°С‚СЊ С‚РѕР»СЊРєРѕ СЂР°Р·СЂРµС€С‘РЅРЅС‹Рµ РґР°РЅРЅС‹Рµ; РїРѕСЂСЏРґРѕРє Рё С„РёР»СЊС‚СЂР°С†РёСЏ РІР°Р¶РЅС‹ РґР»СЏ UX.
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
