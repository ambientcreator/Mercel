package appcore

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
		return User{}, errors.New("Р В РЎСљР В Р’ВµР В РўвЂР В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂў Р В РЎвЂ”Р РЋР вЂљР В Р’В°Р В Р вЂ  Р В РўвЂР В Р’В»Р РЋР РЏ Р РЋР С“Р В РЎвЂўР В Р’В·Р В РўвЂР В Р’В°Р В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р В Р’ВµР В РІвЂћвЂ“.")
	}
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return User{}, errors.New("Р В Р в‚¬Р В РЎвЂќР В Р’В°Р В Р’В¶Р В РЎвЂР РЋРІР‚С™Р В Р’Вµ Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В Р вЂ¦ Р В РЎвЂ Р В РЎвЂ”Р В Р’В°Р РЋР вЂљР В РЎвЂўР В Р’В»Р РЋР Р‰ Р В Р вЂ¦Р В РЎвЂўР В Р вЂ Р В РЎвЂўР В РЎвЂ“Р В РЎвЂў Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР РЏ.")
	}
	if err := validateUsername(username); err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}
	role := normalizeAssignableRole(req.Role)
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Р В РІР‚в„ўР РЋРІР‚в„– Р В Р вЂ¦Р В Р’Вµ Р В РЎВР В РЎвЂўР В Р’В¶Р В Р’ВµР РЋРІР‚С™Р В Р’Вµ Р РЋР С“Р В РЎвЂўР В Р’В·Р В РўвЂР В Р’В°Р РЋРІР‚С™Р РЋР Р‰ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р РЋР РЏ Р РЋР С“ Р РЋР РЉР РЋРІР‚С™Р В РЎвЂўР В РІвЂћвЂ“ Р РЋР вЂљР В РЎвЂўР В Р’В»Р РЋР Р‰Р РЋР вЂ№.")
	}
	createdAt := time.Now().Format(time.RFC3339)
	result, err := a.db.Exec(`INSERT INTO users(username, password_hash, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, username, hashPassword(password), "", 1, "1", "", "", "", role, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	id, _ := result.LastInsertId()
	return a.getUserByID(id)
}

func (a *App) getUserByID(id int64) (User, error) {
	var user User
	err := a.db.QueryRow(`SELECT id, username, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Username, &user.FullName, &user.LastActNumber, &user.PreferredContractCode, &user.ContractSPBKSNumber, &user.ContractGrizablNumber, &user.ContractSignedAt, &user.Role, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// RU: Р В РЎС™Р В Р’ВµР РЋРІР‚С™Р В РЎвЂўР В РўвЂ `ListUsers`.
// EN: Method `ListUsers`.
//
// RU: Р В Р’В§Р РЋРІР‚С™Р В РЎвЂў Р В РўвЂР В Р’ВµР В Р’В»Р В Р’В°Р В Р’ВµР РЋРІР‚С™: Р РЋРІР‚РЋР В РЎвЂР РЋРІР‚С™Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РЎвЂ Р В Р вЂ Р В РЎвЂўР В Р’В·Р В Р вЂ Р РЋР вЂљР В Р’В°Р РЋРІР‚В°Р В Р’В°Р В Р’ВµР РЋРІР‚С™ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р В Р’Вµ Р В РўвЂР В Р’В»Р РЋР РЏ Р РЋРІР‚С›Р РЋР вЂљР В РЎвЂўР В Р вЂ¦Р РЋРІР‚С™Р В Р’ВµР В Р вЂ¦Р В РўвЂР В Р’В° Р В РЎвЂР В Р’В»Р В РЎвЂ Р В Р вЂ Р В Р вЂ¦Р РЋРЎвЂњР РЋРІР‚С™Р РЋР вЂљР В Р’ВµР В Р вЂ¦Р В Р вЂ¦Р В Р’ВµР В РІвЂћвЂ“ Р В Р’В»Р В РЎвЂўР В РЎвЂ“Р В РЎвЂР В РЎвЂќР В РЎвЂ.
// EN: What it does: ListUsers returns the visible users for the current actor, already filtered and sorted for the UI.
//
// RU: Р В РЎв„ўР В Р’В»Р РЋР вЂ№Р РЋРІР‚РЋР В Р’ВµР В Р вЂ Р РЋРІР‚в„–Р В Р’Вµ Р В РЎВР В РЎвЂўР В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚С™Р РЋРІР‚в„–: Р В РЎвЂР РЋР С“Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р РЋРЎвЂњР В Р’ВµР РЋРІР‚С™Р РЋР С“Р РЋР РЏ Р В Р вЂ  Р В РЎвЂўР РЋРІР‚С™Р РЋР вЂљР В РЎвЂР РЋР С“Р В РЎвЂўР В Р вЂ Р В РЎвЂќР В Р’Вµ Р В РЎвЂР В Р вЂ¦Р РЋРІР‚С™Р В Р’ВµР РЋР вЂљР РЋРІР‚С›Р В Р’ВµР В РІвЂћвЂ“Р РЋР С“Р В Р’В°; Р В РўвЂР В РЎвЂўР В Р’В»Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В Р вЂ Р В РЎвЂўР В Р’В·Р В Р вЂ Р РЋР вЂљР В Р’В°Р РЋРІР‚В°Р В Р’В°Р РЋРІР‚С™Р РЋР Р‰ Р РЋРІР‚С™Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В РЎвЂќР В РЎвЂў Р РЋР вЂљР В Р’В°Р В Р’В·Р РЋР вЂљР В Р’ВµР РЋРІвЂљВ¬Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р В Р’Вµ Р В РўвЂР В Р’В°Р В Р вЂ¦Р В Р вЂ¦Р РЋРІР‚в„–Р В Р’Вµ; Р В РЎвЂ”Р В РЎвЂўР РЋР вЂљР РЋР РЏР В РўвЂР В РЎвЂўР В РЎвЂќ Р В РЎвЂ Р РЋРІР‚С›Р В РЎвЂР В Р’В»Р РЋР Р‰Р РЋРІР‚С™Р РЋР вЂљР В Р’В°Р РЋРІР‚В Р В РЎвЂР РЋР РЏ Р В Р вЂ Р В Р’В°Р В Р’В¶Р В Р вЂ¦Р РЋРІР‚в„– Р В РўвЂР В Р’В»Р РЋР РЏ UX.
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListUsers() ([]User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []User{}, nil
	}
	if !canViewManagedUsers(user.Role) {
		return []User{}, nil
	}
	rows, err := a.db.Query(`SELECT id, username, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, role, created_at FROM users WHERE username <> 'admin' ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.FullName, &item.LastActNumber, &item.PreferredContractCode, &item.ContractSPBKSNumber, &item.ContractGrizablNumber, &item.ContractSignedAt, &item.Role, &item.CreatedAt); err != nil {
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
		return User{}, errors.New("Р В РЎСљР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р’В·Р РЋР РЏ Р В РЎВР В Р’ВµР В Р вЂ¦Р РЋР РЏР РЋРІР‚С™Р РЋР Р‰ Р РЋР вЂљР В РЎвЂўР В Р’В»Р РЋР Р‰ Р РЋРЎвЂњ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚В°Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р РЋРЎвЂњР РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂ.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return User{}, errors.New("Р В РЎСљР В Р’ВµР В РўвЂР В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂў Р В РЎвЂ”Р РЋР вЂљР В Р’В°Р В Р вЂ  Р В РўвЂР В Р’В»Р РЋР РЏ Р В РЎвЂР В Р’В·Р В РЎВР В Р’ВµР В Р вЂ¦Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р РЋР вЂљР В РЎвЂўР В Р’В»Р В РЎвЂ Р РЋР РЉР РЋРІР‚С™Р В РЎвЂўР В РІвЂћвЂ“ Р РЋРЎвЂњР РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂ.")
	}
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Р В РІР‚в„ўР РЋРІР‚в„– Р В Р вЂ¦Р В Р’Вµ Р В РЎВР В РЎвЂўР В Р’В¶Р В Р’ВµР РЋРІР‚С™Р В Р’Вµ Р В Р вЂ¦Р В Р’В°Р В Р’В·Р В Р вЂ¦Р В Р’В°Р РЋРІР‚РЋР В РЎвЂР РЋРІР‚С™Р РЋР Р‰ Р РЋР РЉР РЋРІР‚С™Р РЋРЎвЂњ Р РЋР вЂљР В РЎвЂўР В Р’В»Р РЋР Р‰.")
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
		return User{}, errors.New("Р В РЎСљР В РЎвЂўР В Р вЂ Р РЋРІР‚в„–Р В РІвЂћвЂ“ Р В РЎвЂ”Р В Р’В°Р РЋР вЂљР В РЎвЂўР В Р’В»Р РЋР Р‰ Р В Р вЂ¦Р В Р’Вµ Р В РўвЂР В РЎвЂўР В Р’В»Р В Р’В¶Р В Р’ВµР В Р вЂ¦ Р В Р’В±Р РЋРІР‚в„–Р РЋРІР‚С™Р РЋР Р‰ Р В РЎвЂ”Р РЋРЎвЂњР РЋР С“Р РЋРІР‚С™Р РЋРІР‚в„–Р В РЎВ.")
	}
	user, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if !canResetUserPassword(*current, user) {
		return User{}, errors.New("Р В РЎСљР В Р’ВµР В РўвЂР В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂў Р В РЎвЂ”Р РЋР вЂљР В Р’В°Р В Р вЂ  Р В РўвЂР В Р’В»Р РЋР РЏ Р РЋР С“Р В РЎВР В Р’ВµР В Р вЂ¦Р РЋРІР‚в„– Р В РЎвЂ”Р В Р’В°Р РЋР вЂљР В РЎвЂўР В Р’В»Р РЋР РЏ Р РЋР РЉР РЋРІР‚С™Р В РЎвЂўР В РІвЂћвЂ“ Р РЋРЎвЂњР РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂ.")
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
		return User{}, errors.New("Р Р€Р С”Р В°Р В¶Р С‘РЎвЂљР Вµ Р В¤Р ВР С› РЎРѓР С•РЎвЂљРЎР‚РЎС“Р Т‘Р Р…Р С‘Р С”Р В°.")
	}
	if len([]rune(fullName)) > 180 {
		return User{}, errors.New("Р В¤Р ВР С› РЎРѓР С•РЎвЂљРЎР‚РЎС“Р Т‘Р Р…Р С‘Р С”Р В° Р Р…Р Вµ Р Т‘Р С•Р В»Р В¶Р Р…Р С• Р С—РЎР‚Р ВµР Р†РЎвЂ№РЎв‚¬Р В°РЎвЂљРЎРЉ 180 РЎРѓР С‘Р СР Р†Р С•Р В»Р С•Р Р†.")
	}
	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)
	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("Р СњР ВµР Т‘Р С•РЎРѓРЎвЂљР В°РЎвЂљР С•РЎвЂЎР Р…Р С• Р С—РЎР‚Р В°Р Р† Р Т‘Р В»РЎРЏ Р С‘Р В·Р СР ВµР Р…Р ВµР Р…Р С‘РЎРЏ Р В¤Р ВР С› РЎРѓР С•РЎвЂљРЎР‚РЎС“Р Т‘Р Р…Р С‘Р С”Р В°.")
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
		return User{}, errors.New("РЈРєР°Р¶РёС‚Рµ РІР°С€Рµ Р¤РРћ.")
	}
	if len([]rune(fullName)) > 180 {
		return User{}, errors.New("Р¤РРћ РЅРµ РґРѕР»Р¶РЅРѕ РїСЂРµРІС‹С€Р°С‚СЊ 180 СЃРёРјРІРѕР»РѕРІ.")
	}

	signedAt := strings.TrimSpace(req.ContractSignedAt)
	if signedAt == "" {
		return User{}, errors.New("РЈРєР°Р¶РёС‚Рµ РґР°С‚Сѓ РїРѕРґРїРёСЃР°РЅРёСЏ РґРѕРіРѕРІРѕСЂР°.")
	}
	if _, err := time.Parse("2006-01-02", signedAt); err != nil {
		return User{}, errors.New("Р”Р°С‚Р° РїРѕРґРїРёСЃР°РЅРёСЏ РґРѕРіРѕРІРѕСЂР° СѓРєР°Р·Р°РЅР° РІ РЅРµРІРµСЂРЅРѕРј С„РѕСЂРјР°С‚Рµ.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}

	canEdit := target.Username == current.Username || canSeeUser(current.Role, normalizeRole(target.Role)) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ РёР·РјРµРЅРµРЅРёСЏ РґР°РЅРЅС‹С… РґРѕРіРѕРІРѕСЂР°.")
	}

	spbksNumber := strings.TrimSpace(req.ContractSPBKSNumber)
	grizablNumber := strings.TrimSpace(req.ContractGrizablNumber)
	preferredContractCode := strings.TrimSpace(req.PreferredContractCode)
	if preferredContractCode != "1" && preferredContractCode != "2" {
		preferredContractCode = "1"
	}

	if _, err := a.db.Exec(
		`UPDATE users SET full_name = ?, preferred_contract_code = ?, contract_spbks_number = ?, contract_grizabl_number = ?, contract_signed_at = ? WHERE id = ?`,
		fullName, preferredContractCode, spbksNumber, grizablNumber, signedAt, req.UserID,
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
			a.currentSession.PreferredContractCode = updated.PreferredContractCode
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
		return User{}, errors.New("РЈРєР°Р¶РёС‚Рµ РЅРѕРјРµСЂ Р°РєС‚Р° Р±РѕР»СЊС€Рµ 0.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)

	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("РќРµРґРѕСЃС‚Р°С‚РѕС‡РЅРѕ РїСЂР°РІ РґР»СЏ РёР·РјРµРЅРµРЅРёСЏ РЅРѕРјРµСЂР° Р°РєС‚Р°.")
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
		return errors.New("Р В РЎСљР В Р’ВµР В РўвЂР В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂў Р В РЎвЂ”Р РЋР вЂљР В Р’В°Р В Р вЂ  Р В РўвЂР В Р’В»Р РЋР РЏ Р РЋРЎвЂњР В РўвЂР В Р’В°Р В Р’В»Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р В РЎвЂ”Р В РЎвЂўР В Р’В»Р РЋР Р‰Р В Р’В·Р В РЎвЂўР В Р вЂ Р В Р’В°Р РЋРІР‚С™Р В Р’ВµР В Р’В»Р В Р’ВµР В РІвЂћвЂ“.")
	}
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if user.Username == "admin" || user.Username == current.Username {
		return errors.New("Р В РЎСљР В Р’ВµР В Р’В»Р РЋР Р‰Р В Р’В·Р РЋР РЏ Р РЋРЎвЂњР В РўвЂР В Р’В°Р В Р’В»Р В РЎвЂР РЋРІР‚С™Р РЋР Р‰ Р В Р’В·Р В Р’В°Р РЋРІР‚В°Р В РЎвЂР РЋРІР‚В°Р РЋРІР‚ВР В Р вЂ¦Р В Р вЂ¦Р РЋРЎвЂњР РЋР вЂ№ Р РЋРЎвЂњР РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р вЂ¦Р РЋРЎвЂњР РЋР вЂ№ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р РЋР Р‰.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return errors.New("Р В РЎСљР В Р’ВµР В РўвЂР В РЎвЂўР РЋР С“Р РЋРІР‚С™Р В Р’В°Р РЋРІР‚С™Р В РЎвЂўР РЋРІР‚РЋР В Р вЂ¦Р В РЎвЂў Р В РЎвЂ”Р РЋР вЂљР В Р’В°Р В Р вЂ  Р В РўвЂР В Р’В»Р РЋР РЏ Р РЋРЎвЂњР В РўвЂР В Р’В°Р В Р’В»Р В Р’ВµР В Р вЂ¦Р В РЎвЂР РЋР РЏ Р РЋР РЉР РЋРІР‚С™Р В РЎвЂўР В РІвЂћвЂ“ Р РЋРЎвЂњР РЋРІР‚РЋР РЋРІР‚ВР РЋРІР‚С™Р В Р вЂ¦Р В РЎвЂўР В РІвЂћвЂ“ Р В Р’В·Р В Р’В°Р В РЎвЂ”Р В РЎвЂР РЋР С“Р В РЎвЂ.")
	}
	_, err = a.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
