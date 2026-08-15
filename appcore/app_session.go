package appcore

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

// EN: Method `requireAuth`.
//
// EN: What it does: requireAuth returns the current session user or an authorization error when nobody is logged in.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAuth() (*User, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.currentSession == nil {
		return nil, errors.New("\u0421\u043d\u0430\u0447\u0430\u043b\u0430 \u0432\u043e\u0439\u0434\u0438\u0442\u0435 \u0432 \u0441\u0438\u0441\u0442\u0435\u043c\u0443.")
	}
	copyUser := *a.currentSession
	return &copyUser, nil
}

// EN: Method `requireManage`.
//
// EN: What it does: requireManage ensures the caller has management-level rights before continuing.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireManage() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if !roleCanManageUsers(user.Role) {
		return nil, errors.New("\u041d\u0435\u0434\u043e\u0441\u0442\u0430\u0442\u043e\u0447\u043d\u043e \u043f\u0440\u0430\u0432 \u0434\u043b\u044f \u044d\u0442\u043e\u0433\u043e \u0434\u0435\u0439\u0441\u0442\u0432\u0438\u044f.")
	}
	return user, nil
}

// EN: Method `requireAdmin`.
//
// EN: What it does: requireAdmin is the strictest guard and allows only the protected admin account.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) requireAdmin() (*User, error) {
	user, err := a.requireAuth()
	if err != nil {
		return nil, err
	}
	if user.Role != RoleAdmin {
		return nil, errors.New("\u0414\u043e\u0441\u0442\u0443\u043f\u043d\u043e \u0442\u043e\u043b\u044c\u043a\u043e \u0430\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440\u0443.")
	}
	return user, nil
}

// EN: Method `GetBootstrap`.
//
// EN: What it does: GetBootstrap returns the initial application payload used to hydrate the frontend state.
//
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

// EN: Method `GetAppInfo`.
//
// EN: What it does: GetAppInfo reports the version, the resolved database file and the support contact for the settings dialog.
//
// EN: Key points: kept out of GetBootstrap because it never changes during a session, while the bootstrap payload is refetched after every mutation; gated on an open session so the login screen cannot read the storage layout.
func (a *App) GetAppInfo() (AppInfo, error) {
	if _, err := a.requireAuth(); err != nil {
		return AppInfo{}, err
	}

	a.mu.RLock()
	databasePath := a.dbPath
	a.mu.RUnlock()

	if databasePath == "" {
		resolved, err := resolveDatabasePath()
		if err != nil {
			return AppInfo{}, fmt.Errorf("resolve database path: %w", err)
		}
		databasePath = resolved
	}

	return AppInfo{
		Version:        AppVersion,
		DatabasePath:   databasePath,
		SupportContact: AppSupportContact,
	}, nil
}

// EN: Method `Login`.
//
// EN: What it does: Login validates credentials, opens a session and returns capability flags for the current user.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Login(req LoginRequest) (SessionState, error) {
	username := stripSpaces(req.Username)
	password := stripSpaces(req.Password)
	if username == "" || password == "" {
		return SessionState{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043b\u043e\u0433\u0438\u043d \u0438 \u043f\u0430\u0440\u043e\u043b\u044c.")
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
			return SessionState{}, errors.New("\u041f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u044c \u043d\u0435 \u043d\u0430\u0439\u0434\u0435\u043d.")
		}
		return SessionState{}, fmt.Errorf("login query: %w", err)
	}
	if !verifyPassword(password, passwordHash) {
		return SessionState{}, errors.New("\u041d\u0435\u0432\u0435\u0440\u043d\u044b\u0439 \u043b\u043e\u0433\u0438\u043d \u0438\u043b\u0438 \u043f\u0430\u0440\u043e\u043b\u044c.")
	}
	if isLegacyPasswordHash(passwordHash) {
		_, _ = a.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hashPassword(password), user.ID)
	}

	a.mu.Lock()
	a.currentSession = &user
	state := a.sessionStateLocked("\u0412\u0445\u043e\u0434 \u0432\u044b\u043f\u043e\u043b\u043d\u0435\u043d \u0443\u0441\u043f\u0435\u0448\u043d\u043e.")
	a.mu.Unlock()
	return state, nil
}

// EN: Method `Logout`.
//
// EN: What it does: Logout clears the in-memory session and returns a locked guest state to the frontend.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) Logout() SessionState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.currentSession = nil
	return a.sessionStateLocked("\u0412\u044b \u0432\u044b\u0448\u043b\u0438 \u0438\u0437 \u0441\u0438\u0441\u0442\u0435\u043c\u044b.")
}

// EN: Method `GetSession`.
//
// EN: What it does: GetSession exposes the current session snapshot without reloading the full bootstrap payload.
//
// EN: Key points: is consumed by the frontend during rendering; must return only permitted data; ordering and filtering matter for the UI.
func (a *App) GetSession() SessionState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sessionStateLocked("\u0412\u044b \u0432\u044b\u0448\u043b\u0438 \u0438\u0437 \u0441\u0438\u0441\u0442\u0435\u043c\u044b.")
}

// EN: Method `GetServices`.
//
// EN: What it does: GetServices returns only the services belonging to the currently authenticated user.
//
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

// EN: Function `formatMoney`.
//
// EN: What it does: formatMoney produces a compact Russian-currency string for generated descriptions and archive titles.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.

// EN: Method `ListContractTemplates`.
//
// EN: What it does: ListContractTemplates returns every act blank the exporter supports, so the UI never duplicates their names.
//
// EN: Key points: derived from resolveContractInfo, which stays the only description of a blank.
func (a *App) ListContractTemplates() ([]ContractTemplate, error) {
	if _, err := a.requireAuth(); err != nil {
		return nil, err
	}

	templates := make([]ContractTemplate, 0, len(contractTemplateCodes))
	for _, code := range contractTemplateCodes {
		info, err := resolveContractInfo(code)
		if err != nil {
			return nil, err
		}
		templates = append(templates, ContractTemplate{
			Code:          info.Code,
			Title:         info.Title,
			CustomerName:  info.CustomerName,
			DirectorShort: info.CustomerDirectorShort,
		})
	}
	return templates, nil
}

// EN: Method `SetPreferredContractTemplate`.
//
// EN: What it does: SetPreferredContractTemplate switches which act blank the signed-in user exports with.
//
// EN: Key points: touches only the blank column, so it works before the contract number, date and full name are filled in — unlike UpdateUserContractDetails, which validates all of them.
func (a *App) SetPreferredContractTemplate(code string) (User, error) {
	current, err := a.requireAuth()
	if err != nil {
		return User{}, err
	}

	code = strings.TrimSpace(code)
	if _, err := resolveContractInfo(code); err != nil {
		return User{}, err
	}

	if _, err := a.db.Exec(`UPDATE users SET preferred_contract_code = ? WHERE id = ?`, code, current.ID); err != nil {
		return User{}, fmt.Errorf("update preferred contract template: %w", err)
	}

	updated, err := a.getUserByID(current.ID)
	if err != nil {
		return User{}, err
	}

	a.mu.Lock()
	if a.currentSession != nil && a.currentSession.ID == updated.ID {
		a.currentSession.PreferredContractCode = updated.PreferredContractCode
	}
	a.mu.Unlock()

	return updated, nil
}
