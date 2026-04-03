package main

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

func (a *App) CreateUser(req UserWithPassword) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	if !canCreateUsers(current.Role) {
		return User{}, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ СЃРѕР·РґР°РЅРёСЏ РїРѕР»СЊР·РѕРІР°С‚РµР»РµР№.")
	}
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return User{}, errors.New("РЈРєР°Р¶РёС‚Рµ Р»РѕРіРёРЅ Рё РїР°СЂРѕР»СЊ РЅРѕРІРѕРіРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ.")
	}
	if err := validateUsername(username); err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}
	role := normalizeAssignableRole(req.Role)
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Р’С‹ РЅРµ РјРѕР¶РµС‚Рµ СЃРѕР·РґР°С‚СЊ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ СЃ СЌС‚РѕР№ СЂРѕР»СЊСЋ.")
	}
	createdAt := time.Now().Format(time.RFC3339)
	result, err := a.db.Exec(`INSERT INTO users(username, password_hash, role, created_at) VALUES(?, ?, ?, ?)`, username, hashPassword(password), role, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	id, _ := result.LastInsertId()
	return a.getUserByID(id)
}

func (a *App) getUserByID(id int64) (User, error) {
	var user User
	err := a.db.QueryRow(`SELECT id, username, role, created_at FROM users WHERE id = ?`, id).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// RU: РњРµС‚РѕРґ `ListUsers`.
// EN: Method `ListUsers`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: С‡РёС‚Р°РµС‚ Рё РІРѕР·РІСЂР°С‰Р°РµС‚ РґР°РЅРЅС‹Рµ РґР»СЏ С„СЂРѕРЅС‚РµРЅРґР° РёР»Рё РІРЅСѓС‚СЂРµРЅРЅРµР№ Р»РѕРіРёРєРё.
// EN: What it does: ListUsers returns the visible users for the current actor, already filtered and sorted for the UI.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РёСЃРїРѕР»СЊР·СѓРµС‚СЃСЏ РІ РѕС‚СЂРёСЃРѕРІРєРµ РёРЅС‚РµСЂС„РµР№СЃР°; РґРѕР»Р¶РµРЅ РІРѕР·РІСЂР°С‰Р°С‚СЊ С‚РѕР»СЊРєРѕ СЂР°Р·СЂРµС€С‘РЅРЅС‹Рµ РґР°РЅРЅС‹Рµ; РїРѕСЂСЏРґРѕРє Рё С„РёР»СЊС‚СЂР°С†РёСЏ РІР°Р¶РЅС‹ РґР»СЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListUsers() ([]User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []User{}, nil
	}
	if !canViewManagedUsers(user.Role) {
		return []User{}, nil
	}
	rows, err := a.db.Query(`SELECT id, username, role, created_at FROM users WHERE username <> 'admin' ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.Role, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		item.Role = normalizeRole(item.Role)
		if item.Username == user.Username {
			continue
		}
		if canSeeUser(user.Role, item.Role) {
			users = append(users, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(users, func(i, j int) bool {
		leftPriority := userSortPriority(users[i].Role)
		rightPriority := userSortPriority(users[j].Role)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		if users[i].CreatedAt != users[j].CreatedAt {
			return users[i].CreatedAt < users[j].CreatedAt
		}
		return users[i].Username < users[j].Username
	})
	return users, nil
}
func (a *App) UpdateUserRole(userID int64, role string) (User, error) {
	current, err := a.requireManage()
	if err != nil {
		return User{}, err
	}
	role = normalizeAssignableRole(role)
	user, err := a.getUserByID(userID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if user.Username == "admin" || user.Username == current.Username {
		return User{}, errors.New("РќРµР»СЊР·СЏ РјРµРЅСЏС‚СЊ СЂРѕР»СЊ Сѓ Р·Р°С‰РёС‰С‘РЅРЅРѕР№ СѓС‡С‘С‚РЅРѕР№ Р·Р°РїРёСЃРё.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return User{}, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ РёР·РјРµРЅРµРЅРёСЏ СЂРѕР»Рё СЌС‚РѕР№ СѓС‡С‘С‚РЅРѕР№ Р·Р°РїРёСЃРё.")
	}
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Р’С‹ РЅРµ РјРѕР¶РµС‚Рµ РЅР°Р·РЅР°С‡РёС‚СЊ СЌС‚Сѓ СЂРѕР»СЊ.")
	}

	_, err = a.db.Exec(`UPDATE users SET role = ? WHERE id = ?`, role, userID)
	if err != nil {
		return User{}, fmt.Errorf("update user role: %w", err)
	}
	updated, err := a.getUserByID(userID)
	if err != nil {
		return User{}, err
	}
	updated.Role = normalizeRole(updated.Role)
	return updated, nil
}

func (a *App) ResetUserPassword(req ResetUserPasswordRequest) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	password := stripSpaces(req.Password)
	if password == "" {
		return User{}, errors.New("РќРѕРІС‹Р№ РїР°СЂРѕР»СЊ РЅРµ РґРѕР»Р¶РµРЅ Р±С‹С‚СЊ РїСѓСЃС‚С‹Рј.")
	}
	user, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if !canResetUserPassword(*current, user) {
		return User{}, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ СЃРјРµРЅС‹ РїР°СЂРѕР»СЏ СЌС‚РѕР№ СѓС‡С‘С‚РЅРѕР№ Р·Р°РїРёСЃРё.")
	}
	if _, err := a.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hashPassword(password), req.UserID); err != nil {
		return User{}, fmt.Errorf("reset user password: %w", err)
	}
	return a.getUserByID(req.UserID)
}

func (a *App) DeleteUser(userID int64) error {
	current, err := a.requireAuth()
	if err != nil {
		return err
	}
	if !canViewManagedUsers(current.Role) {
		return errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ СѓРґР°Р»РµРЅРёСЏ РїРѕР»СЊР·РѕРІР°С‚РµР»РµР№.")
	}
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if user.Username == "admin" || user.Username == current.Username {
		return errors.New("РќРµР»СЊР·СЏ СѓРґР°Р»РёС‚СЊ Р·Р°С‰РёС‰С‘РЅРЅСѓСЋ СѓС‡С‘С‚РЅСѓСЋ Р·Р°РїРёСЃСЊ.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ СѓРґР°Р»РµРЅРёСЏ СЌС‚РѕР№ СѓС‡С‘С‚РЅРѕР№ Р·Р°РїРёСЃРё.")
	}
	_, err = a.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
