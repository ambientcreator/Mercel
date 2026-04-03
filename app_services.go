package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func formatMoney(value int) string {
	return displayMoney(value)
}

// RU: Р¤СѓРЅРєС†РёСЏ `displayMoney`.
// EN: Function `displayMoney`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: displayMoney is a UI-facing helper that mirrors money formatting where a separate name reads clearer.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func displayMoney(value int) string {
	return fmt.Sprintf("%d СЂ", value)
}

func archiveDateTitle(now time.Time) string {
	return now.Format("02.01.2006")
}

// RU: РњРµС‚РѕРґ `saveServiceForOwner`.
// EN: Method `saveServiceForOwner`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РѕРґРёРЅ РёР· РєР»СЋС‡РµРІС‹С… С€Р°РіРѕРІ backend-Р»РѕРіРёРєРё РІРЅСѓС‚СЂРё РїСЂРёР»РѕР¶РµРЅРёСЏ.
// EN: What it does: saveServiceForOwner performs the actual insert/update of a service for a concrete user owner.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) saveServiceForOwner(req UpsertServiceRequest, owner string) (Service, error) {
	name := strings.TrimSpace(req.Name)
	unit := normalizeUnit(req.Unit)
	if err := validateServiceName(name); err != nil {
		return Service{}, err
	}
	if unit == "" || req.Rate <= 0 || req.Rate > 100000000 {
		return Service{}, errors.New("РЈРєР°Р¶РёС‚Рµ РєРѕСЂСЂРµРєС‚РЅС‹Рµ РµРґРёРЅРёС†Сѓ Рё СЃС‚РѕРёРјРѕСЃС‚СЊ СѓСЃР»СѓРіРё.")
	}
	category := normalizeCategory(req.Category)
	var allocation interface{}
	if req.AllocationPercent != nil {
		if *req.AllocationPercent < 0 || *req.AllocationPercent > 100 {
			return Service{}, errors.New("РџСЂРѕС†РµРЅС‚ СѓСЃР»СѓРіРё РґРѕР»Р¶РµРЅ Р±С‹С‚СЊ РІ РґРёР°РїР°Р·РѕРЅРµ РѕС‚ 0 РґРѕ 100.")
		}
		allocation = *req.AllocationPercent
	}
	code := generateServiceCode(owner + "-" + name)
	createdAt := time.Now().Format(time.RFC3339)

	if req.ID == 0 {
		result, err := a.db.Exec(`INSERT INTO services(code, name, unit, rate, category, allocation_percent, created_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, code, name, unit, req.Rate, category, allocation, owner, createdAt)
		if err != nil {
			return Service{}, fmt.Errorf("create service: %w", err)
		}
		id, _ := result.LastInsertId()
		return a.getServiceByIDForOwner(id, owner)
	}

	_, err := a.db.Exec(`UPDATE services SET name = ?, unit = ?, rate = ?, category = ?, allocation_percent = ? WHERE id = ? AND created_by = ?`, name, unit, req.Rate, category, allocation, req.ID, owner)
	if err != nil {
		return Service{}, fmt.Errorf("update service: %w", err)
	}
	return a.getServiceByIDForOwner(req.ID, owner)
}

func (a *App) saveService(req UpsertServiceRequest) (Service, error) {
	user, err := a.requireAuth()
	if err != nil {
		return Service{}, err
	}
	return a.saveServiceForOwner(req, user.Username)
}

// RU: РњРµС‚РѕРґ `getServiceByIDForOwner`.
// EN: Method `getServiceByIDForOwner`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РѕРґРёРЅ РёР· РєР»СЋС‡РµРІС‹С… С€Р°РіРѕРІ backend-Р»РѕРіРёРєРё РІРЅСѓС‚СЂРё РїСЂРёР»РѕР¶РµРЅРёСЏ.
// EN: What it does: getServiceByIDForOwner fetches one service while enforcing ownership boundaries.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) getServiceByIDForOwner(id int64, owner string) (Service, error) {
	var item Service
	var allocation sql.NullFloat64
	err := a.db.QueryRow(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE id = ? AND created_by = ?`, id, owner).Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt)
	if err != nil {
		return Service{}, err
	}
	item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
	if allocation.Valid {
		value := allocation.Float64
		item.AllocationPercent = &value
	}
	return item, nil
}

func (a *App) getServiceByNameForOwner(name string, owner string) (Service, error) {
	var item Service
	var allocation sql.NullFloat64
	err := a.db.QueryRow(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE created_by = ? AND name = ? ORDER BY id ASC LIMIT 1`, owner, strings.TrimSpace(name)).Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt)
	if err != nil {
		return Service{}, err
	}
	item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
	if allocation.Valid {
		value := allocation.Float64
		item.AllocationPercent = &value
	}
	return item, nil
}

// RU: РњРµС‚РѕРґ `UpsertService`.
// EN: Method `UpsertService`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РґР°РЅРЅС‹С… РІ РїСЂРёР»РѕР¶РµРЅРёРё Рё РїСЂРѕРІРѕРґРёС‚ Р±РёР·РЅРµСЃ-РѕРїРµСЂР°С†РёСЋ.
// EN: What it does: UpsertService is the public service create/update API used by the settings screen.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РјРµРЅСЏРµС‚ СЃРѕСЃС‚РѕСЏРЅРёРµ SQLite РёР»Рё СЃРµСЃСЃРёРё; РѕРїРёСЂР°РµС‚СЃСЏ РЅР° РїСЂРѕРІРµСЂРєРё СЂРѕР»РµР№ Рё РІР»Р°РґРµРЅРёСЏ; РѕС€РёР±РєРё Р·РґРµСЃСЊ Р·Р°РјРµС‚РЅС‹ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ СЃСЂР°Р·Сѓ.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) UpsertService(req UpsertServiceRequest) (Service, error) {
	if _, err := a.requireAuth(); err != nil {
		return Service{}, err
	}
	return a.saveService(req)
}

// RU: РњРµС‚РѕРґ `DeleteService`.
// EN: Method `DeleteService`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РґР°РЅРЅС‹С… РІ РїСЂРёР»РѕР¶РµРЅРёРё Рё РїСЂРѕРІРѕРґРёС‚ Р±РёР·РЅРµСЃ-РѕРїРµСЂР°С†РёСЋ.
// EN: What it does: DeleteService removes one owned service after authorization and ownership checks.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РјРµРЅСЏРµС‚ СЃРѕСЃС‚РѕСЏРЅРёРµ SQLite РёР»Рё СЃРµСЃСЃРёРё; РѕРїРёСЂР°РµС‚СЃСЏ РЅР° РїСЂРѕРІРµСЂРєРё СЂРѕР»РµР№ Рё РІР»Р°РґРµРЅРёСЏ; РѕС€РёР±РєРё Р·РґРµСЃСЊ Р·Р°РјРµС‚РЅС‹ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ СЃСЂР°Р·Сѓ.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) DeleteService(id int64) error {
	user, err := a.requireAuth()
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("РќРµРєРѕСЂСЂРµРєС‚РЅС‹Р№ РёРґРµРЅС‚РёС„РёРєР°С‚РѕСЂ СѓСЃР»СѓРіРё.")
	}
	_, err = a.db.Exec(`DELETE FROM services WHERE id = ? AND created_by = ?`, id, user.Username)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// RU: РњРµС‚РѕРґ `CreateUser`.
// EN: Method `CreateUser`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РґР°РЅРЅС‹С… РІ РїСЂРёР»РѕР¶РµРЅРёРё Рё РїСЂРѕРІРѕРґРёС‚ Р±РёР·РЅРµСЃ-РѕРїРµСЂР°С†РёСЋ.
// EN: What it does: CreateUser creates a subordinate account according to the current actor permissions.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РјРµРЅСЏРµС‚ СЃРѕСЃС‚РѕСЏРЅРёРµ SQLite РёР»Рё СЃРµСЃСЃРёРё; РѕРїРёСЂР°РµС‚СЃСЏ РЅР° РїСЂРѕРІРµСЂРєРё СЂРѕР»РµР№ Рё РІР»Р°РґРµРЅРёСЏ; РѕС€РёР±РєРё Р·РґРµСЃСЊ Р·Р°РјРµС‚РЅС‹ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ СЃСЂР°Р·Сѓ.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
