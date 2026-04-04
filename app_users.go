package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

func (a *App) CreateUser(req UserWithPassword) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	if !canCreateUsers(current.Role) {
		return User{}, errors.New("Р СњР ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎвЂљР С•РЎвЂЎР Р…Р С• Р С—РЎР‚Р В°Р Р† Р Т‘Р В»РЎРЏ РЎРѓР С•Р В·Р Т‘Р В°Р Р…Р С‘РЎРЏ Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»Р ВµР в„–.")
	}
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return User{}, errors.New("Р Р€Р С”Р В°Р В¶Р С‘РЎвЂљР Вµ Р В»Р С•Р С–Р С‘Р Р… Р С‘ Р С—Р В°РЎР‚Р С•Р В»РЎРЉ Р Р…Р С•Р Р†Р С•Р С–Р С• Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»РЎРЏ.")
	}
	if err := validateUsername(username); err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}
	role := normalizeAssignableRole(req.Role)
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Р вЂ™РЎвЂ№ Р Р…Р Вµ Р СР С•Р В¶Р ВµРЎвЂљР Вµ РЎРѓР С•Р В·Р Т‘Р В°РЎвЂљРЎРЉ Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»РЎРЏ РЎРѓ РЎРЊРЎвЂљР С•Р в„– РЎР‚Р С•Р В»РЎРЉРЎР‹.")
	}
	createdAt := time.Now().Format(time.RFC3339)
	result, err := a.db.Exec(`INSERT INTO users(username, password_hash, full_name, last_act_number, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, username, hashPassword(password), "", 1, "", "", "", role, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	id, _ := result.LastInsertId()
	return a.getUserByID(id)
}

func (a *App) getUserByID(id int64) (User, error) {
	var user User
	err := a.db.QueryRow(`SELECT id, username, full_name, last_act_number, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Username, &user.FullName, &user.LastActNumber, &user.ContractSPBKSNumber, &user.ContractGrizablNumber, &user.ContractSignedAt, &user.Role, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// RU: Р СљР ВµРЎвЂљР С•Р Т‘ `ListUsers`.
// EN: Method `ListUsers`.
//
// RU: Р В§РЎвЂљР С• Р Т‘Р ВµР В»Р В°Р ВµРЎвЂљ: РЎвЂЎР С‘РЎвЂљР В°Р ВµРЎвЂљ Р С‘ Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°Р ВµРЎвЂљ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В»РЎРЏ РЎвЂћРЎР‚Р С•Р Р…РЎвЂљР ВµР Р…Р Т‘Р В° Р С‘Р В»Р С‘ Р Р†Р Р…РЎС“РЎвЂљРЎР‚Р ВµР Р…Р Р…Р ВµР в„– Р В»Р С•Р С–Р С‘Р С”Р С‘.
// EN: What it does: ListUsers returns the visible users for the current actor, already filtered and sorted for the UI.
//
// RU: Р С™Р В»РЎР‹РЎвЂЎР ВµР Р†РЎвЂ№Р Вµ Р СР С•Р СР ВµР Р…РЎвЂљРЎвЂ№: Р С‘РЎРѓР С—Р С•Р В»РЎРЉР В·РЎС“Р ВµРЎвЂљРЎРѓРЎРЏ Р Р† Р С•РЎвЂљРЎР‚Р С‘РЎРѓР С•Р Р†Р С”Р Вµ Р С‘Р Р…РЎвЂљР ВµРЎР‚РЎвЂћР ВµР в„–РЎРѓР В°; Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р Р†Р С•Р В·Р Р†РЎР‚Р В°РЎвЂ°Р В°РЎвЂљРЎРЉ РЎвЂљР С•Р В»РЎРЉР С”Р С• РЎР‚Р В°Р В·РЎР‚Р ВµРЎв‚¬РЎвЂР Р…Р Р…РЎвЂ№Р Вµ Р Т‘Р В°Р Р…Р Р…РЎвЂ№Р Вµ; Р С—Р С•РЎР‚РЎРЏР Т‘Р С•Р С” Р С‘ РЎвЂћР С‘Р В»РЎРЉРЎвЂљРЎР‚Р В°РЎвЂ Р С‘РЎРЏ Р Р†Р В°Р В¶Р Р…РЎвЂ№ Р Т‘Р В»РЎРЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListUsers() ([]User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []User{}, nil
	}
	if !canViewManagedUsers(user.Role) {
		return []User{}, nil
	}
	rows, err := a.db.Query(`SELECT id, username, full_name, last_act_number, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at FROM users WHERE username <> 'admin' ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.FullName, &item.LastActNumber, &item.ContractSPBKSNumber, &item.ContractGrizablNumber, &item.ContractSignedAt, &item.Role, &item.CreatedAt); err != nil {
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
		return User{}, errors.New("Р СњР ВµР В»РЎРЉР В·РЎРЏ Р СР ВµР Р…РЎРЏРЎвЂљРЎРЉ РЎР‚Р С•Р В»РЎРЉ РЎС“ Р В·Р В°РЎвЂ°Р С‘РЎвЂ°РЎвЂР Р…Р Р…Р С•Р в„– РЎС“РЎвЂЎРЎвЂРЎвЂљР Р…Р С•Р в„– Р В·Р В°Р С—Р С‘РЎРѓР С‘.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return User{}, errors.New("Р СњР ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎвЂљР С•РЎвЂЎР Р…Р С• Р С—РЎР‚Р В°Р Р† Р Т‘Р В»РЎРЏ Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ РЎР‚Р С•Р В»Р С‘ РЎРЊРЎвЂљР С•Р в„– РЎС“РЎвЂЎРЎвЂРЎвЂљР Р…Р С•Р в„– Р В·Р В°Р С—Р С‘РЎРѓР С‘.")
	}
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Р вЂ™РЎвЂ№ Р Р…Р Вµ Р СР С•Р В¶Р ВµРЎвЂљР Вµ Р Р…Р В°Р В·Р Р…Р В°РЎвЂЎР С‘РЎвЂљРЎРЉ РЎРЊРЎвЂљРЎС“ РЎР‚Р С•Р В»РЎРЉ.")
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
		return User{}, errors.New("Р СњР С•Р Р†РЎвЂ№Р в„– Р С—Р В°РЎР‚Р С•Р В»РЎРЉ Р Р…Р Вµ Р Т‘Р С•Р В»Р В¶Р ВµР Р… Р В±РЎвЂ№РЎвЂљРЎРЉ Р С—РЎС“РЎРѓРЎвЂљРЎвЂ№Р С.")
	}
	user, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if !canResetUserPassword(*current, user) {
		return User{}, errors.New("Р СњР ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎвЂљР С•РЎвЂЎР Р…Р С• Р С—РЎР‚Р В°Р Р† Р Т‘Р В»РЎРЏ РЎРѓР СР ВµР Р…РЎвЂ№ Р С—Р В°РЎР‚Р С•Р В»РЎРЏ РЎРЊРЎвЂљР С•Р в„– РЎС“РЎвЂЎРЎвЂРЎвЂљР Р…Р С•Р в„– Р В·Р В°Р С—Р С‘РЎРѓР С‘.")
	}
	if _, err := a.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hashPassword(password), req.UserID); err != nil {
		return User{}, fmt.Errorf("reset user password: %w", err)
	}
	return a.getUserByID(req.UserID)
}

func (a *App) UpdateUserFullName(req UpdateUserFullNameRequest) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		return User{}, errors.New("РЈРєР°Р¶РёС‚Рµ Р¤РРћ СЃРѕС‚СЂСѓРґРЅРёРєР°.")
	}
	if len([]rune(fullName)) > 180 {
		return User{}, errors.New("Р¤РРћ СЃРѕС‚СЂСѓРґРЅРёРєР° РЅРµ РґРѕР»Р¶РЅРѕ РїСЂРµРІС‹С€Р°С‚СЊ 180 СЃРёРјРІРѕР»РѕРІ.")
	}
	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)
	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ РёР·РјРµРЅРµРЅРёСЏ Р¤РРћ СЃРѕС‚СЂСѓРґРЅРёРєР°.")
	}
	if _, err := a.db.Exec(`UPDATE users SET full_name = ? WHERE id = ?`, fullName, req.UserID); err != nil {
		return User{}, fmt.Errorf("update user full name: %w", err)
	}
	updated, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, err
	}
	if current.Username == updated.Username {
		a.mu.Lock()
		if a.currentSession != nil && a.currentSession.ID == updated.ID {
			a.currentSession.FullName = updated.FullName
		}
		a.mu.Unlock()
	}
	return updated, nil
}

func (a *App) UpdateUserContractDetails(req UpdateUserContractDetailsRequest) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}

	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		return User{}, errors.New("Укажите ваше ФИО.")
	}
	if len([]rune(fullName)) > 180 {
		return User{}, errors.New("ФИО не должно превышать 180 символов.")
	}

	signedAt := strings.TrimSpace(req.ContractSignedAt)
	if signedAt == "" {
		return User{}, errors.New("Укажите дату подписания договора.")
	}
	if _, err := time.Parse("2006-01-02", signedAt); err != nil {
		return User{}, errors.New("Дата подписания договора указана в неверном формате.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}

	canEdit := target.Username == current.Username || canSeeUser(current.Role, normalizeRole(target.Role)) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("Недостаточно прав для изменения данных договора.")
	}

	spbksNumber := strings.TrimSpace(req.ContractSPBKSNumber)
	grizablNumber := strings.TrimSpace(req.ContractGrizablNumber)

	if _, err := a.db.Exec(
		`UPDATE users SET full_name = ?, contract_spbks_number = ?, contract_grizabl_number = ?, contract_signed_at = ? WHERE id = ?`,
		fullName, spbksNumber, grizablNumber, signedAt, req.UserID,
	); err != nil {
		return User{}, fmt.Errorf("update user contract details: %w", err)
	}

	updated, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, err
	}

	if current.Username == updated.Username {
		a.mu.Lock()
		if a.currentSession != nil && a.currentSession.ID == updated.ID {
			a.currentSession.FullName = updated.FullName
			a.currentSession.ContractSPBKSNumber = updated.ContractSPBKSNumber
			a.currentSession.ContractGrizablNumber = updated.ContractGrizablNumber
			a.currentSession.ContractSignedAt = updated.ContractSignedAt
		}
		a.mu.Unlock()
	}

	return updated, nil
}

func (a *App) UpdateUserLastActNumber(req UpdateUserLastActNumberRequest) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}
	if req.LastActNumber <= 0 {
		return User{}, errors.New("Укажите номер акта больше 0.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)

	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("Недостаточно прав для изменения номера акта.")
	}

	if _, err := a.db.Exec(`UPDATE users SET last_act_number = ? WHERE id = ?`, req.LastActNumber, req.UserID); err != nil {
		return User{}, fmt.Errorf("update user last act number: %w", err)
	}

	updated, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, err
	}
	if current.Username == updated.Username {
		a.mu.Lock()
		if a.currentSession != nil && a.currentSession.ID == updated.ID {
			a.currentSession.LastActNumber = updated.LastActNumber
		}
		a.mu.Unlock()
	}
	return updated, nil
}

func (a *App) DeleteUser(userID int64) error {
	current, err := a.requireAuth()
	if err != nil {
		return err
	}
	if !canViewManagedUsers(current.Role) {
		return errors.New("Р СњР ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎвЂљР С•РЎвЂЎР Р…Р С• Р С—РЎР‚Р В°Р Р† Р Т‘Р В»РЎРЏ РЎС“Р Т‘Р В°Р В»Р ВµР Р…Р С‘РЎРЏ Р С—Р С•Р В»РЎРЉР В·Р С•Р Р†Р В°РЎвЂљР ВµР В»Р ВµР в„–.")
	}
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if user.Username == "admin" || user.Username == current.Username {
		return errors.New("Р СњР ВµР В»РЎРЉР В·РЎРЏ РЎС“Р Т‘Р В°Р В»Р С‘РЎвЂљРЎРЉ Р В·Р В°РЎвЂ°Р С‘РЎвЂ°РЎвЂР Р…Р Р…РЎС“РЎР‹ РЎС“РЎвЂЎРЎвЂРЎвЂљР Р…РЎС“РЎР‹ Р В·Р В°Р С—Р С‘РЎРѓРЎРЉ.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return errors.New("Р СњР ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎвЂљР С•РЎвЂЎР Р…Р С• Р С—РЎР‚Р В°Р Р† Р Т‘Р В»РЎРЏ РЎС“Р Т‘Р В°Р В»Р ВµР Р…Р С‘РЎРЏ РЎРЊРЎвЂљР С•Р в„– РЎС“РЎвЂЎРЎвЂРЎвЂљР Р…Р С•Р в„– Р В·Р В°Р С—Р С‘РЎРѓР С‘.")
	}
	_, err = a.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}


