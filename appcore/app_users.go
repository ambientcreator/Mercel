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
		return User{}, errors.New("Недостаточно прав для создания пользователей.")
	}
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return User{}, errors.New("Укажите имя пользователя и пароль.")
	}
	if err := validateUsername(username); err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}
	role := normalizeAssignableRole(req.Role)
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Недостаточно прав для назначения этой роли.")
	}
	createdAt := time.Now().Format(time.RFC3339)
	result, err := a.db.Exec(`INSERT INTO users(username, password_hash, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, act_template, role, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, username, hashPassword(password), "", 1, "1", "", "", "", a.nextActTemplate(), role, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("resolve created user id: %w", err)
	}
	return a.getUserByID(id)
}

func (a *App) getUserByID(id int64) (User, error) {
	var user User
	err := a.db.QueryRow(`SELECT id, username, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, act_template, role, created_at FROM users WHERE id = ?`, id).
		Scan(&user.ID, &user.Username, &user.FullName, &user.LastActNumber, &user.PreferredContractCode, &user.ContractSPBKSNumber, &user.ContractGrizablNumber, &user.ContractSignedAt, &user.ActTemplate, &user.Role, &user.CreatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// EN: Method `ListUsers`.
//
// EN: What it does: ListUsers returns the visible users for the current actor, already filtered and sorted for the UI.
//
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) ListUsers() ([]User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return []User{}, nil
	}
	if !canViewManagedUsers(user.Role) {
		return []User{}, nil
	}
	rows, err := a.db.Query(`SELECT id, username, full_name, last_act_number, preferred_contract_code, contract_spbks_number, contract_grizabl_number, contract_signed_at, act_template, role, created_at FROM users WHERE username <> 'admin' ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var item User
		if err := rows.Scan(&item.ID, &item.Username, &item.FullName, &item.LastActNumber, &item.PreferredContractCode, &item.ContractSPBKSNumber, &item.ContractGrizablNumber, &item.ContractSignedAt, &item.ActTemplate, &item.Role, &item.CreatedAt); err != nil {
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
		return User{}, errors.New("Нельзя изменить роль защищённого пользователя или свою собственную роль.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return User{}, errors.New("Недостаточно прав для изменения этого пользователя.")
	}
	if !canCreateRole(current.Role, role) {
		return User{}, errors.New("Недостаточно прав для назначения этой роли.")
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
		return User{}, errors.New("Укажите новый пароль.")
	}
	user, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if !canResetUserPassword(*current, user) {
		return User{}, errors.New("Недостаточно прав для сброса пароля этого пользователя.")
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
		return User{}, errors.New("Укажите ФИО пользователя.")
	}
	if len([]rune(fullName)) > 180 {
		return User{}, errors.New("ФИО пользователя не должно превышать 180 символов.")
	}
	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)
	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("Недостаточно прав для изменения ФИО пользователя.")
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
		return User{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0432\u0430\u0448\u0435 \u0424\u0418\u041e.")
	}
	if len([]rune(fullName)) > 180 {
		return User{}, errors.New("\u0424\u0418\u041e \u043d\u0435 \u0434\u043e\u043b\u0436\u043d\u043e \u043f\u0440\u0435\u0432\u044b\u0448\u0430\u0442\u044c 180 \u0441\u0438\u043c\u0432\u043e\u043b\u043e\u0432.")
	}

	signedAt := strings.TrimSpace(req.ContractSignedAt)
	if signedAt == "" {
		return User{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.")
	}
	if _, err := time.Parse("2006-01-02", signedAt); err != nil {
		return User{}, errors.New("\u0414\u0430\u0442\u0430 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430 \u0443\u043a\u0430\u0437\u0430\u043d\u0430 \u0432 \u043d\u0435\u0432\u0435\u0440\u043d\u043e\u043c \u0444\u043e\u0440\u043c\u0430\u0442\u0435.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}

	canEdit := target.Username == current.Username || canSeeUser(current.Role, normalizeRole(target.Role)) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("\u041d\u0435\u0434\u043e\u0441\u0442\u0430\u0442\u043e\u0447\u043d\u043e \u043f\u0440\u0430\u0432 \u0434\u043b\u044f \u0438\u0437\u043c\u0435\u043d\u0435\u043d\u0438\u044f \u0434\u0430\u043d\u043d\u044b\u0445 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.")
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
		return User{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0430\u043a\u0442\u0430 \u0431\u043e\u043b\u044c\u0448\u0435 0.")
	}

	target, err := a.getUserByID(req.UserID)
	if err != nil {
		return User{}, fmt.Errorf("find user: %w", err)
	}
	target.Role = normalizeRole(target.Role)

	canEdit := target.Username == current.Username || canSeeUser(current.Role, target.Role) || normalizeRole(current.Role) == RoleAdmin
	if !canEdit {
		return User{}, errors.New("\u041d\u0435\u0434\u043e\u0441\u0442\u0430\u0442\u043e\u0447\u043d\u043e \u043f\u0440\u0430\u0432 \u0434\u043b\u044f \u0438\u0437\u043c\u0435\u043d\u0435\u043d\u0438\u044f \u043d\u043e\u043c\u0435\u0440\u0430 \u0430\u043a\u0442\u0430.")
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
		return errors.New("Недостаточно прав для удаления пользователей.")
	}
	user, err := a.getUserByID(userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	user.Role = normalizeRole(user.Role)
	if user.Username == "admin" || user.Username == current.Username {
		return errors.New("Нельзя удалить защищённого пользователя или свою собственную учётную запись.")
	}
	if !canSeeUser(current.Role, user.Role) {
		return errors.New("Недостаточно прав для удаления этого пользователя.")
	}
	_, err = a.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
