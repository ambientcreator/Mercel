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

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `requireAuth`.
// EN: Method `requireAuth`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р С•Р Т‘Р С‘Р Р… Р С‘Р В· Р С”Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№РЎвЂ¦ РЎв‚¬Р В°Р С–Р С•Р Р† backend-Р В»Р С•Р С–Р С‘Р С”Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р С‘ Р С—РЎР‚Р С‘Р В»Р С•Р В¶Р ВµР Р…Р С‘РЎРЏ.
// EN: What it does: requireAuth returns the current session user or an authorization error when nobody is logged in.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAuth() (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.currentSession == nil {
		return nil, errors.New("Сначала войдите в систему.")
	}
	copyUser := *a.currentSession
	return &copyUser, nil
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `requireManage`.
// EN: Method `requireManage`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р С•Р Т‘Р С‘Р Р… Р С‘Р В· Р С”Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№РЎвЂ¦ РЎв‚¬Р В°Р С–Р С•Р Р† backend-Р В»Р С•Р С–Р С‘Р С”Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р С‘ Р С—РЎР‚Р С‘Р В»Р С•Р В¶Р ВµР Р…Р С‘РЎРЏ.
// EN: What it does: requireManage ensures the caller has management-level rights before continuing.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireManage() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if !roleCanManageUsers(user.Role) {
		return nil, errors.New("Недостаточно прав для этого действия.")
	}
	return user, nil
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `requireAdmin`.
// EN: Method `requireAdmin`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р С•Р Т‘Р С‘Р Р… Р С‘Р В· Р С”Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№РЎвЂ¦ РЎв‚¬Р В°Р С–Р С•Р Р† backend-Р В»Р С•Р С–Р С‘Р С”Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р С‘ Р С—РЎР‚Р С‘Р В»Р С•Р В¶Р ВµР Р…Р С‘РЎРЏ.
// EN: What it does: requireAdmin is the strictest guard and allows only the protected admin account.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAdmin() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if user.Role != RoleAdmin {
		return nil, errors.New("Доступно только администратору.")
	}
	return user, nil
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `GetBootstrap`.
// EN: Method `GetBootstrap`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎвЂЎР С‘РЎвЂљР В°Р ВµРЎвЂљ Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ РЎвЂћРЎР‚Р С•Р Р…РЎвЂљР ВµР Р…Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р ВµР Р…Р Р…Р ВµР в„– Р В»Р С•Р С–Р С‘Р С”Р С‘.
// EN: What it does: GetBootstrap returns the initial application payload used to hydrate the frontend state.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ Р Р† Р С•РЎвЂљРЎР‚Р С‘РЎРѓР С•Р Р†Р С”Р Вµ Р С‘Р Р…РЎвЂљР ВµРЎР‚РЎвЂћР ВµР в„–РЎРѓР В°; Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°РЎвЂљРЎРЉ РЎвЂљР С•Р В»РЎРЉР С”Р С• РЎР‚Р В°Р В·РЎР‚Р ВµРЎв‚¬РЎвЂР Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ; Р С—Р С•РЎР‚РЎРЏР Т‘Р С•Р С” Р С‘ РЎвЂћР С‘Р В»РЎРЉРЎвЂљРЎР‚Р В°РЎвЂ Р С‘РЎРЏ Р Р†Р В°Р В¶Р Р…РЎвЂ№ Р Т‘Р В»РЎРЏ UX.
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

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `Login`.
// EN: Method `Login`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎС“Р С—РЎР‚Р В°Р Р†Р В»РЎРЏР ВµРЎвЂљ РЎРѓР ВµРЎРѓРЎРѓР С‘Р ВµР в„– Р С‘ Р Р…Р В°Р В±Р С•РЎР‚Р С•Р С Р С—РЎР‚Р В°Р Р† РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР С–Р С• Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»РЎРЏ.
// EN: What it does: Login validates credentials, opens a session and returns capability flags for the current user.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Login(req LoginRequest) (SessionState, error) {
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return SessionState{}, errors.New("Укажите логин и пароль.")
	}
	if err := validateUsername(username); err != nil {
		return SessionState{}, err
	}
	if err := validatePassword(password); err != nil {
		return SessionState{}, err
	}

	var user User
	var passwordHash string
	err := a.db.QueryRow(`SELECT id, username, password_hash, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &passwordHash, &user.FullName, &user.LastActNumber, &user.PreferredContractCode, &user.ContractSPBKSNumber, &user.ContractGrizablNumber, &user.ContractSignedAt, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SessionState{}, errors.New("Пользователь не найден.")
		}
		return SessionState{}, fmt.Errorf("login query: %w", err)
	}
	if !verifyPassword(password, passwordHash) {
		return SessionState{}, errors.New("Неверный логин или пароль.")
	}
	if isLegacyPasswordHash(passwordHash) {
		_, _ = a.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hashPassword(password), user.ID)
	}

	a.mu.Lock()
	a.currentSession = &user
	state := a.sessionStateLocked("Вход выполнен успешно.")
	a.mu.Unlock()
	return state, nil
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `Logout`.
// EN: Method `Logout`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р В·Р В°Р Р†Р ВµРЎР‚РЎв‚¬Р В°Р ВµРЎвЂљ РЎРѓР ВµРЎРѓРЎРѓР С‘РЎР‹ Р С‘ Р С•РЎвЂЎР С‘РЎвЂ°Р В°Р ВµРЎвЂљ Р Р†РЎРѓР Вµ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ РЎвЂљР ВµР С”РЎС“РЎвЂ°Р ВµР в„– Р В°Р Р†РЎвЂљР С•РЎР‚Р С‘Р В·Р В°РЎвЂ Р С‘Р С‘.
// EN: What it does: Logout clears the in-memory session and returns a locked guest state to the frontend.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ Р С”Р С•РЎР‚РЎР‚Р ВµР С”РЎвЂљР Р…Р С•Р С–Р С• logout-РЎРѓРЎвЂ Р ВµР Р…Р В°РЎР‚Р С‘РЎРЏ; Р С—Р С•РЎРѓР В»Р Вµ Р Р†РЎвЂ№Р В·Р С•Р Р†Р В° Р С‘Р Р…РЎвЂљР ВµРЎР‚РЎвЂћР ВµР в„–РЎРѓ Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р С—Р ВµРЎР‚Р ВµР в„–РЎвЂљР С‘ Р Р† Р С–Р С•РЎРѓРЎвЂљР ВµР Р†Р С•Р Вµ РЎРѓР С•РЎРѓРЎвЂљР С•РЎРЏР Р…Р С‘Р Вµ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Logout() SessionState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentSession = nil
	return a.sessionStateLocked("Вы вышли из системы.")
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `GetSession`.
// EN: Method `GetSession`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎвЂЎР С‘РЎвЂљР В°Р ВµРЎвЂљ Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ РЎвЂћРЎР‚Р С•Р Р…РЎвЂљР ВµР Р…Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р ВµР Р…Р Р…Р ВµР в„– Р В»Р С•Р С–Р С‘Р С”Р С‘.
// EN: What it does: GetSession exposes the current session snapshot without reloading the full bootstrap payload.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ Р Р† Р С•РЎвЂљРЎР‚Р С‘РЎРѓР С•Р Р†Р С”Р Вµ Р С‘Р Р…РЎвЂљР ВµРЎР‚РЎвЂћР ВµР в„–РЎРѓР В°; Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°РЎвЂљРЎРЉ РЎвЂљР С•Р В»РЎРЉР С”Р С• РЎР‚Р В°Р В·РЎР‚Р ВµРЎв‚¬РЎвЂР Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ; Р С—Р С•РЎР‚РЎРЏР Т‘Р С•Р С” Р С‘ РЎвЂћР С‘Р В»РЎРЉРЎвЂљРЎР‚Р В°РЎвЂ Р С‘РЎРЏ Р Р†Р В°Р В¶Р Р…РЎвЂ№ Р Т‘Р В»РЎРЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetSession() SessionState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionStateLocked("Вы вышли из системы.")
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `GetServices`.
// EN: Method `GetServices`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎвЂЎР С‘РЎвЂљР В°Р ВµРЎвЂљ Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ РЎвЂћРЎР‚Р С•Р Р…РЎвЂљР ВµР Р…Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р ВµР Р…Р Р…Р ВµР в„– Р В»Р С•Р С–Р С‘Р С”Р С‘.
// EN: What it does: GetServices returns only the services belonging to the currently authenticated user.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ Р Р† Р С•РЎвЂљРЎР‚Р С‘РЎРѓР С•Р Р†Р С”Р Вµ Р С‘Р Р…РЎвЂљР ВµРЎР‚РЎвЂћР ВµР в„–РЎРѓР В°; Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°РЎвЂљРЎРЉ РЎвЂљР С•Р В»РЎРЉР С”Р С• РЎР‚Р В°Р В·РЎР‚Р ВµРЎв‚¬РЎвЂР Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ; Р С—Р С•РЎР‚РЎРЏР Т‘Р С•Р С” Р С‘ РЎвЂћР С‘Р В»РЎРЉРЎвЂљРЎР‚Р В°РЎвЂ Р С‘РЎРЏ Р Р†Р В°Р В¶Р Р…РЎвЂ№ Р Т‘Р В»РЎРЏ UX.
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

// RU: Р В¤РЎС“Р Р…Р С”РЎвЂ Р С‘РЎРЏ `formatMoney`.
// EN: Function `formatMoney`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: Р Р†РЎвЂ№Р С—Р С•Р В»Р Р…РЎРЏР ВµРЎвЂљ Р Р†РЎРѓР С—Р С•Р СР С•Р С–Р В°РЎвЂљР ВµР В»РЎРЉР Р…Р С•Р Вµ Р С—РЎР‚Р ВµР С•Р В±РЎР‚Р В°Р В·Р С•Р Р†Р В°Р Р…Р С‘Р Вµ, Р С—РЎР‚Р С•Р Р†Р ВµРЎР‚Р С”РЎС“ Р С‘Р В»Р С‘ Р С—Р С•Р Т‘Р С–Р С•РЎвЂљР С•Р Р†Р С”РЎС“ Р Т‘Р В°Р Р…Р Р…РЎвЂ№РЎвЂ¦.
// EN: What it does: formatMoney produces a compact Russian-currency string for generated descriptions and archive titles.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р Р†Р В°Р В¶Р ВµР Р… Р Т‘Р В»РЎРЏ РЎС“РЎРѓРЎвЂљР С•Р в„–РЎвЂЎР С‘Р Р†Р С•РЎРѓРЎвЂљР С‘ Р В»Р С•Р С–Р С‘Р С”Р С‘; Р СР С•Р В¶Р ВµРЎвЂљ Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљРЎРЉРЎРѓРЎРЏ РЎРѓРЎР‚Р В°Р В·РЎС“ Р Р† Р Р…Р ВµРЎРѓР С”Р С•Р В»РЎРЉР С”Р С‘РЎвЂ¦ Р СР ВµРЎРѓРЎвЂљР В°РЎвЂ¦; Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎРѓРЎвЂљР С•Р С‘РЎвЂљ Р Т‘Р ВµР В»Р В°РЎвЂљРЎРЉ Р С•РЎРѓР С•Р В·Р Р…Р В°Р Р…Р Р…Р С•.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.


