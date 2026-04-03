package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func (a *App) sessionStateLocked(message string) SessionState {
	state := SessionState{Authenticated: a.currentSession != nil, Message: message}
	if a.currentSession != nil {
		copyUser := *a.currentSession
		state.User = &copyUser
		state.CanManage = roleCanManageUsers(copyUser.Role)
		state.CanAdmin = normalizeRole(copyUser.Role) == RoleAdmin
		state.CanModerate = canModerateArchives(copyUser.Role)
	}
	return state
}

// RU: РњРµС‚РѕРґ `requireAuth`.
// EN: Method `requireAuth`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РѕРґРёРЅ РёР· РєР»СЋС‡РµРІС‹С… С€Р°РіРѕРІ backend-Р»РѕРіРёРєРё РІРЅСѓС‚СЂРё РїСЂРёР»РѕР¶РµРЅРёСЏ.
// EN: What it does: requireAuth returns the current session user or an authorization error when nobody is logged in.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAuth() (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.currentSession == nil {
		return nil, errors.New("РЎРЅР°С‡Р°Р»Р° РІРѕР№РґРёС‚Рµ РІ СЃРёСЃС‚РµРјСѓ.")
	}
	copyUser := *a.currentSession
	return &copyUser, nil
}

// RU: РњРµС‚РѕРґ `requireManage`.
// EN: Method `requireManage`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РѕРґРёРЅ РёР· РєР»СЋС‡РµРІС‹С… С€Р°РіРѕРІ backend-Р»РѕРіРёРєРё РІРЅСѓС‚СЂРё РїСЂРёР»РѕР¶РµРЅРёСЏ.
// EN: What it does: requireManage ensures the caller has management-level rights before continuing.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireManage() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if !roleCanManageUsers(user.Role) {
		return nil, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ СЌС‚РѕРіРѕ РґРµР№СЃС‚РІРёСЏ.")
	}
	return user, nil
}

// RU: РњРµС‚РѕРґ `requireAdmin`.
// EN: Method `requireAdmin`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РѕРґРёРЅ РёР· РєР»СЋС‡РµРІС‹С… С€Р°РіРѕРІ backend-Р»РѕРіРёРєРё РІРЅСѓС‚СЂРё РїСЂРёР»РѕР¶РµРЅРёСЏ.
// EN: What it does: requireAdmin is the strictest guard and allows only the protected admin account.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAdmin() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if user.Role != RoleAdmin {
		return nil, errors.New("Р”РѕСЃС‚СѓРїРЅРѕ С‚РѕР»СЊРєРѕ Р°РґРјРёРЅРёСЃС‚СЂР°С‚РѕСЂСѓ.")
	}
	return user, nil
}

// RU: РњРµС‚РѕРґ `GetBootstrap`.
// EN: Method `GetBootstrap`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: С‡РёС‚Р°РµС‚ Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РґР°РЅРЅС‹Рµ РґР»СЏ С„СЂРѕРЅС‚РµРЅРґР° РёР»Рё РІРЅСѓС‚СЂРµРЅРЅРµР№ Р»РѕРіРёРєРё.
// EN: What it does: GetBootstrap returns the initial application payload used to hydrate the frontend state.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ РІ РѕС‚СЂРёСЃРѕРІРєРµ РёРЅС‚РµСЂС„РµР№СЃР°; РґРѕР»Р¶РµРЅ РІРѕР·РІСЂР°С‰Р°С‚СЊ С‚РѕР»СЊРєРѕ СЂР°Р·СЂРµС€С‘РЅРЅС‹Рµ РґР°РЅРЅС‹Рµ; РїРѕСЂСЏРґРѕРє Рё С„РёР»СЊС‚СЂР°С†РёСЏ РІР°Р¶РЅС‹ РґР»СЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetBootstrap() (AppBootstrap, error) {
	services, err := a.GetServices()
	if err != nil {
		return AppBootstrap{}, err
	}
	calculations, err := a.ListCalculations()
	if err != nil {
		return AppBootstrap{}, err
	}
	users, err := a.ListUsers()
	if err != nil {
		return AppBootstrap{}, err
	}

	a.mu.RLock()
	session := a.sessionStateLocked("")
	a.mu.RUnlock()

	return AppBootstrap{
		Session:             session,
		Services:            services,
		Users:               users,
		SavedCalculations:   calculations,
		DefaultGroupPercent: defaultGroupPercent(),
	}, nil
}

// RU: РњРµС‚РѕРґ `Login`.
// EN: Method `Login`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓРїСЂР°РІР»СЏРµС‚ СЃРµСЃСЃРёРµР№ Рё РЅР°Р±РѕСЂРѕРј РїСЂР°РІ С‚РµРєСѓС‰РµРіРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ.
// EN: What it does: Login validates credentials, opens a session and returns capability flags for the current user.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Login(req LoginRequest) (SessionState, error) {
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return SessionState{}, errors.New("РЈРєР°Р¶РёС‚Рµ Р»РѕРіРёРЅ Рё РїР°СЂРѕР»СЊ.")
	}
	if err := validateUsername(username); err != nil {
		return SessionState{}, err
	}
	if err := validatePassword(password); err != nil {
		return SessionState{}, err
	}

	var user User
	var passwordHash string
	err := a.db.QueryRow(`SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`, username).Scan(&user.ID, &user.Username, &passwordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionState{}, errors.New("РџРѕР»СЊР·РѕРІР°С‚РµР»СЊ РЅРµ РЅР°Р№РґРµРЅ.")
		}
		return SessionState{}, fmt.Errorf("login query: %w", err)
	}
	if !verifyPassword(password, passwordHash) {
		return SessionState{}, errors.New("РќРµРІРµСЂРЅС‹Р№ Р»РѕРіРёРЅ РёР»Рё РїР°СЂРѕР»СЊ.")
	}
	if isLegacyPasswordHash(passwordHash) {
		_, _ = a.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hashPassword(password), user.ID)
	}

	a.mu.Lock()
	a.currentSession = &user
	state := a.sessionStateLocked("Р’С…РѕРґ РІС‹РїРѕР»РЅРµРЅ СѓСЃРїРµС€РЅРѕ.")
	a.mu.Unlock()
	return state, nil
}

// RU: РњРµС‚РѕРґ `Logout`.
// EN: Method `Logout`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: Р·Р°РІРµСЂС€Р°РµС‚ СЃРµСЃСЃРёСЋ Рё РѕС‡РёС‰Р°РµС‚ РІСЃРµ РґР°РЅРЅС‹Рµ С‚РµРєСѓС‰РµР№ Р°РІС‚РѕСЂРёР·Р°С†РёРё.
// EN: What it does: Logout clears the in-memory session and returns a locked guest state to the frontend.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ РєРѕСЂСЂРµРєС‚РЅРѕРіРѕ logout-СЃС†РµРЅР°СЂРёСЏ; РїРѕСЃР»Рµ РІС‹Р·РѕРІР° РёРЅС‚РµСЂС„РµР№СЃ РґРѕР»Р¶РµРЅ РїРµСЂРµР№С‚Рё РІ РіРѕСЃС‚РµРІРѕРµ СЃРѕСЃС‚РѕСЏРЅРёРµ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Logout() SessionState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentSession = nil
	return a.sessionStateLocked("Р’С‹ РІС‹С€Р»Рё РёР· СЃРёСЃС‚РµРјС‹.")
}

// RU: РњРµС‚РѕРґ `GetSession`.
// EN: Method `GetSession`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: С‡РёС‚Р°РµС‚ Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РґР°РЅРЅС‹Рµ РґР»СЏ С„СЂРѕРЅС‚РµРЅРґР° РёР»Рё РІРЅСѓС‚СЂРµРЅРЅРµР№ Р»РѕРіРёРєРё.
// EN: What it does: GetSession exposes the current session snapshot without reloading the full bootstrap payload.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ РІ РѕС‚СЂРёСЃРѕРІРєРµ РёРЅС‚РµСЂС„РµР№СЃР°; РґРѕР»Р¶РµРЅ РІРѕР·РІСЂР°С‰Р°С‚СЊ С‚РѕР»СЊРєРѕ СЂР°Р·СЂРµС€С‘РЅРЅС‹Рµ РґР°РЅРЅС‹Рµ; РїРѕСЂСЏРґРѕРє Рё С„РёР»СЊС‚СЂР°С†РёСЏ РІР°Р¶РЅС‹ РґР»СЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetSession() SessionState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionStateLocked("")
}

// RU: РњРµС‚РѕРґ `GetServices`.
// EN: Method `GetServices`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: С‡РёС‚Р°РµС‚ Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РґР°РЅРЅС‹Рµ РґР»СЏ С„СЂРѕРЅС‚РµРЅРґР° РёР»Рё РІРЅСѓС‚СЂРµРЅРЅРµР№ Р»РѕРіРёРєРё.
// EN: What it does: GetServices returns only the services belonging to the currently authenticated user.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ РІ РѕС‚СЂРёСЃРѕРІРєРµ РёРЅС‚РµСЂС„РµР№СЃР°; РґРѕР»Р¶РµРЅ РІРѕР·РІСЂР°С‰Р°С‚СЊ С‚РѕР»СЊРєРѕ СЂР°Р·СЂРµС€С‘РЅРЅС‹Рµ РґР°РЅРЅС‹Рµ; РїРѕСЂСЏРґРѕРє Рё С„РёР»СЊС‚СЂР°С†РёСЏ РІР°Р¶РЅС‹ РґР»СЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetServices() ([]Service, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []Service{}, nil
	}
	rows, err := a.db.Query(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE created_by = ? ORDER BY id ASC`, user.Username)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	services := make([]Service, 0)
	for rows.Next() {
		var item Service
		var allocation sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
		if allocation.Valid {
			value := allocation.Float64
			item.AllocationPercent = &value
		}
		services = append(services, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services: %w", err)
	}
	return services, nil
}

// RU: Р¤СѓРЅРєС†РёСЏ `formatMoney`.
// EN: Function `formatMoney`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: formatMoney produces a compact Russian-currency string for generated descriptions and archive titles.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
