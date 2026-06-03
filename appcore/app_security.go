package appcore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var safeSQLiteIdentifiers = map[string]struct{}{
	"users":        {},
	"services":     {},
	"calculations": {},
}

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9._@-]{3,32}$`)

func isSafeSQLiteIdentifier(name string) bool {
	_, ok := safeSQLiteIdentifiers[name]
	return ok
}

func validateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("Имя пользователя должно быть от 3 до 32 символов и содержать только буквы, цифры, '.', '_', '-' или '@'.")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 4 || len(password) > 128 {
		return errors.New("Пароль должен содержать от 4 до 128 символов.")
	}
	return nil
}

func validateServiceName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return errors.New("Название услуги не должно быть пустым.")
	}
	if len([]rune(trimmed)) > 120 {
		return errors.New("Название услуги не должно превышать 120 символов.")
	}
	for _, r := range trimmed {
		if r < 32 {
			return errors.New("Название услуги содержит недопустимые управляющие символы.")
		}
	}
	return nil
}

func legacyHashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func isLegacyPasswordHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return legacyHashPassword(password)
	}
	return string(hash)
}

func verifyPassword(password string, storedHash string) bool {
	if strings.HasPrefix(storedHash, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)) == nil
	}
	return legacyHashPassword(password) == storedHash
}

// EN: Function `stripSpaces`.
//
// EN: What it does: stripSpaces removes whitespace around and inside a string where credentials should not keep spaces.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func stripSpaces(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), "")
}

// EN: Function `looksLikeMojibake`.
//
// EN: What it does: looksLikeMojibake detects common UTF-8/Windows-1251 corruption markers in stored Russian text.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func looksLikeMojibake(value string) bool {
	patterns := []string{
		"\u0420\u00a0",
		"\u0420\u2019",
		"\u0420\u0402",
		"\u0420\u0452",
		"\u0420\u045c",
		"\u0420\u0408",
		"\u0420\u0459",
		"\u0420\u040e",
		"\u0420\u040b",
		"\u0421\u0403",
		"\u0421\u2021",
		"\u0421\u2039",
		"\u0421\u0152",
		"\u0421\u201a",
		"\u0432\u0402",
	}
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

// EN: Function `sanitizeStoredText`.
//
// EN: What it does: sanitizeStoredText normalizes archived titles and labels before returning them to the frontend.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func sanitizeStoredText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if looksLikeMojibake(value) {
		return ""
	}
	return value
}

// EN: Function `defaultGroupPercent`.
//
// EN: What it does: defaultGroupPercent returns the fallback allocation share for each service category.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func defaultGroupPercent() map[string]float64 {
	return map[string]float64{
		CategoryPrimary:   0.79,
		CategorySecondary: 0.20,
		CategoryClosing:   0.01,
	}
}

// EN: Function `rolePower`.
//
// EN: What it does: rolePower maps roles to a comparable numeric hierarchy for authorization checks.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func roleDepartment(role string) string {
	meta, ok := roleCatalog[normalizeRole(role)]
	if !ok {
		return DepartmentSupport
	}
	return meta.Department
}

func roleLevel(role string) int {
	meta, ok := roleCatalog[normalizeRole(role)]
	if !ok {
		return 0
	}
	return meta.Level
}

func rolePower(role string) int {
	return roleLevel(role)
}

func userSortPriority(role string) int {
	switch roleLevel(role) {
	case 5:
		return 0
	case 4:
		return 1
	case 3:
		return 2
	case 2:
		return 3
	case 1:
		return 4
	default:
		return 9
	}
}

func normalizeRole(role string) string {
	switch strings.TrimSpace(role) {
	case "manager", "support_manager":
		return RoleSupportHead
	case "senior_specialist", "support_senior_specialist":
		return RoleSupportSenior
	case "employee", "support_employee":
		return RoleSupportEmployee
	case "tech_manager":
		return RoleTechnicalHead
	case "senior_technician":
		return RoleTechnicalSenior
	case "technician":
		return RoleTechnicalEmployee
	case "mrk_manager":
		return RoleCommercialSubscriberHead
	case "senior_mrk":
		return RoleCommercialSeniorMRK
	case "mrk_employee":
		return RoleCommercialEmployeeMRK
	case RoleAdmin, RoleGlobalDirector, RoleExecutiveDirector, RoleTechnicalDirector,
		RoleSupportHead, RoleSupportSenior, RoleSupportSysadmin,
		RoleTechnicalHead, RoleTechnicalSenior, RoleTechnicalEmployee,
		RoleTelecomDirector, RoleTelecomHead, RoleTelecomSeniorVOLS, RoleTelecomSeniorLVS, RoleTelecomEmployeeVOLS, RoleTelecomEmployeeLVS,
		RoleSKUDHead, RoleSKUDProjectManager, RoleSKUDSeniorService, RoleSKUDSeniorInstaller, RoleSKUDServiceEngineer, RoleSKUDInstaller,
		RoleApprovalHead, RoleApprovalSenior, RoleApprovalEmployee,
		RoleMarketingHead, RoleMarketingCourier,
		RoleCommercialDirector, RoleCommercialSubscriberHead, RoleCommercialActiveSalesHead, RoleCommercialSeniorMRK, RoleCommercialSeniorMRYU, RoleCommercialEmployeeMRK, RoleCommercialEmployeeMRYU,
		RoleFinanceHead, RoleFinanceEmployee,
		RoleLegalEmployee,
		RoleDevelopmentHead, RoleDevelopmentSenior, RoleDevelopmentEmployee:
		return strings.TrimSpace(role)
	default:
		return RoleSupportEmployee
	}
}

func normalizeAssignableRole(role string) string {
	role = normalizeRole(role)
	if role == RoleAdmin {
		return RoleSupportEmployee
	}
	if _, ok := roleCatalog[role]; ok {
		return role
	}
	return RoleSupportEmployee
}

func roleViewDepartments(role string) []string {
	switch normalizeRole(role) {
	case RoleAdmin, RoleGlobalDirector, RoleExecutiveDirector:
		return append([]string{}, allBusinessDepartments...)
	case RoleTechnicalDirector:
		return []string{DepartmentTechnical, DepartmentTelecom}
	default:
		dept := roleDepartment(role)
		if dept == DepartmentGlobal {
			return nil
		}
		return []string{dept}
	}
}

func roleCanManageUsers(role string) bool {
	switch normalizeRole(role) {
	case RoleAdmin:
		return true
	case RoleSupportHead, RoleSupportSysadmin, RoleTechnicalHead, RoleTelecomDirector, RoleTelecomHead, RoleSKUDHead, RoleApprovalHead, RoleMarketingHead, RoleCommercialDirector, RoleCommercialSubscriberHead, RoleCommercialActiveSalesHead, RoleFinanceHead, RoleDevelopmentHead:
		return true
	default:
		return false
	}
}

func roleCanCreateUsers(role string) bool {
	// Heads/sysadmins (via roleCanManageUsers), senior specialists (level 2) and
	// directors (level 4) may create users. The exact roles they are allowed to
	// assign — and in which departments — are constrained by canCreateRole.
	return roleCanManageUsers(role) || roleLevel(role) == 2 || roleLevel(role) == 4
}

func departmentInScope(actorRole string, targetDepartment string) bool {
	for _, department := range roleViewDepartments(actorRole) {
		if department == targetDepartment {
			return true
		}
	}
	return false
}

func canCreateUsers(role string) bool {
	return roleCanCreateUsers(role)
}

func canCreateRole(actorRole string, targetRole string) bool {
	actorRole = normalizeRole(actorRole)
	targetRole = normalizeAssignableRole(targetRole)
	if actorRole == RoleAdmin {
		return targetRole != RoleAdmin
	}
	if !roleCanCreateUsers(actorRole) {
		return false
	}
	// The target department must be within the actor's visible scope. For heads and
	// seniors that is only their own department; for directors it spans every
	// department they oversee (e.g. a technical director covers technical+telecom).
	if !departmentInScope(actorRole, roleDepartment(targetRole)) {
		return false
	}
	// No one below admin may create another global-department account (i.e. a
	// director or admin), regardless of level — directors cannot create directors.
	if roleDepartment(targetRole) == DepartmentGlobal {
		return false
	}
	return roleLevel(actorRole) > roleLevel(targetRole)
}

func canViewAllArchives(role string) bool {
	return normalizeRole(role) == RoleAdmin || normalizeRole(role) == RoleGlobalDirector || normalizeRole(role) == RoleExecutiveDirector
}

func canViewManagedUsers(role string) bool {
	return len(roleViewDepartments(role)) > 0 && (roleCanCreateUsers(role) || roleLevel(role) >= 4)
}

func canModerateArchives(role string) bool {
	return len(roleViewDepartments(role)) > 0 && roleLevel(role) >= 2
}

func canSeeUser(actorRole string, targetRole string) bool {
	actorRole = normalizeRole(actorRole)
	targetRole = normalizeRole(targetRole)
	if actorRole == RoleAdmin {
		return true
	}
	if !departmentInScope(actorRole, roleDepartment(targetRole)) {
		return false
	}
	if roleDepartment(actorRole) == DepartmentGlobal {
		return targetRole != RoleAdmin && roleDepartment(targetRole) != DepartmentGlobal
	}
	if !canViewManagedUsers(actorRole) {
		return false
	}
	return roleLevel(actorRole) > roleLevel(targetRole)
}

func canResetUserPassword(actor User, target User) bool {
	actor.Role = normalizeRole(actor.Role)
	target.Role = normalizeRole(target.Role)
	if target.Username == "admin" || target.Username == actor.Username {
		return false
	}
	if actor.Role == RoleAdmin {
		return target.Role != RoleAdmin
	}
	return roleCanCreateUsers(actor.Role) && canSeeUser(actor.Role, target.Role)
}

func canViewArchiveRole(actor *User, authorRole string, authorUsername string) bool {
	viewerRole := normalizeRole(actor.Role)
	authorRole = normalizeRole(authorRole)
	if viewerRole == RoleAdmin {
		return true
	}
	if authorUsername == actor.Username {
		return true
	}
	if !departmentInScope(viewerRole, roleDepartment(authorRole)) {
		return false
	}
	if roleDepartment(viewerRole) == DepartmentGlobal {
		return roleDepartment(authorRole) != DepartmentGlobal
	}
	switch roleLevel(viewerRole) {
	case 3:
		return roleLevel(authorRole) <= 2
	case 2:
		return roleLevel(authorRole) == 1
	default:
		return false
	}
}

func normalizeCategory(category string) string {
	switch category {
	case CategoryPrimary, CategorySecondary, CategoryClosing:
		return category
	default:
		return CategoryPrimary
	}
}

// EN: Function `generateServiceCode`.
//
// EN: What it does: generateServiceCode builds a stable service code from the service name for weight and mapping logic.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func generateServiceCode(name string) string {
	base := strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "(", "", ")", "", ",", "", ".", "")
	base = replacer.Replace(base)
	builder := strings.Builder{}
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	code := strings.Trim(builder.String(), "-")
	hashSource := sha256.Sum256([]byte(name))
	suffix := hex.EncodeToString(hashSource[:])[:8]
	if code == "" {
		return fmt.Sprintf("service-%s", suffix)
	}
	return fmt.Sprintf("%s-%s", code, suffix)
}

// EN: Function `normalizeUnit`.
//
// EN: What it does: normalizeUnit restricts service units to the two UI-supported values: hours and pieces.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func normalizeUnit(unit string) string {
	switch strings.TrimSpace(strings.ToLower(unit)) {
	case "\u0447", "\u0447.", "\u0447\u0430\u0441", "\u0447\u0430\u0441\u044b":
		return "\u0447."
	case "\u0448\u0442", "\u0448\u0442.", "\u0448\u0442\u0443\u043a\u0430", "\u0448\u0442\u0443\u043a\u0438":
		return "\u0448\u0442."
	default:
		return ""
	}
}

// EN: What it does: sessionStateLocked builds a safe guest-like session snapshot used when no active session is available.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
