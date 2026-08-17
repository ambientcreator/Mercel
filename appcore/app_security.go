package appcore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
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

// EN: Variable `legacyRoleAliases`.
//
// EN: What it does: legacyRoleAliases maps role identifiers written by older builds onto their current equivalents.
//
// EN: Key points: it is the single source of truth for what an old stored role means, shared by normalizeRole and by
// EN: the archive SQL filter; migrations rewrite these values in place, but rows written before a migration ran may
// EN: still carry them, so both readers must agree on the mapping.
var legacyRoleAliases = map[string]string{
	"manager":                   RoleSupportHead,
	"support_manager":           RoleSupportHead,
	"senior_specialist":         RoleSupportSenior,
	"support_senior_specialist": RoleSupportSenior,
	"employee":                  RoleSupportEmployee,
	"tech_manager":              RoleTechnicalHead,
	"senior_technician":         RoleTechnicalSenior,
	"technician":                RoleTechnicalEmployee,
	"mrk_manager":               RoleCommercialSubscriberHead,
	"senior_mrk":                RoleCommercialSeniorMRK,
	"mrk_employee":              RoleCommercialEmployeeMRK,
}

func normalizeRole(role string) string {
	role = strings.TrimSpace(role)
	if canonical, ok := legacyRoleAliases[role]; ok {
		return canonical
	}
	if _, ok := roleCatalog[role]; ok {
		return role
	}
	return RoleSupportEmployee
}

// EN: Function `recognizedStoredRoles`.
//
// EN: What it does: recognizedStoredRoles lists every raw string normalizeRole recognises — the canonical identifiers
// EN: plus the legacy aliases — in a stable order.
//
// EN: Key points: anything outside this set normalizes to RoleSupportEmployee, which is why the archive filter needs
// EN: the list: it turns "unrecognized" into an expressible SQL condition instead of a silent bucket.
func recognizedStoredRoles() []string {
	values := make([]string, 0, len(roleCatalog)+len(legacyRoleAliases))
	for role := range roleCatalog {
		values = append(values, role)
	}
	for alias := range legacyRoleAliases {
		if _, canonical := roleCatalog[alias]; !canonical {
			values = append(values, alias)
		}
	}
	sort.Strings(values)
	return values
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

// EN: Constant `unrecognizedRoleProbe`.
//
// EN: What it does: unrecognizedRoleProbe is a role string that can never be a real identifier, used to ask
// EN: canViewArchiveRole how it treats roles it does not recognise.
//
// EN: Key points: it contains a NUL byte, so no stored role and no validated input can collide with it.
const unrecognizedRoleProbe = "\x00unrecognized-role"

// EN: Function `sqlPlaceholders`.
//
// EN: What it does: sqlPlaceholders renders `?, ?, ?` for a bound list of the given length.
//
// EN: Key points: returns an empty string for non-positive counts so callers can skip the clause entirely.
func sqlPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?, ", count), ", ")
}

// EN: Function `archiveVisibilityFilter`.
//
// EN: What it does: archiveVisibilityFilter renders canViewArchiveRole as a SQL predicate over
// EN: calculations.created_by and calculations.created_role, so the database returns only the rows the actor may
// EN: read instead of the caller reading every archive row and discarding most of them in Go.
//
// EN: Key points: canViewArchiveRole stays the single source of truth — every recognised role identifier is asked,
// EN: and the answers become an IN list, so the rules are never restated in SQL by hand. Rows whose stored role is
// EN: unrecognised normalize to RoleSupportEmployee, so they are admitted by a NOT IN clause exactly when that role
// EN: is visible. An empty predicate means "no restriction" and is only returned for the admin.
func archiveVisibilityFilter(actor *User) (string, []any) {
	if actor == nil || actor.Username == "" {
		return "0 = 1", nil
	}
	if normalizeRole(actor.Role) == RoleAdmin {
		return "", nil
	}

	recognized := recognizedStoredRoles()
	visible := make([]string, 0, len(recognized))
	for _, role := range recognized {
		if canViewArchiveRole(actor, role, "") {
			visible = append(visible, role)
		}
	}

	// Authors always reach their own archive, whatever role was recorded on the row.
	clauses := []string{"c.created_by = ?"}
	args := []any{actor.Username}

	if len(visible) > 0 {
		clauses = append(clauses, "c.created_role IN ("+sqlPlaceholders(len(visible))+")")
		for _, role := range visible {
			args = append(args, role)
		}
	}
	if canViewArchiveRole(actor, unrecognizedRoleProbe, "") {
		clauses = append(clauses, "c.created_role NOT IN ("+sqlPlaceholders(len(recognized))+")")
		for _, role := range recognized {
			args = append(args, role)
		}
	}

	return "(" + strings.Join(clauses, " OR ") + ")", args
}

// EN: Function `canCopyArchiveServices`.
//
// EN: What it does: canCopyArchiveServices decides who may import the services of an archived calculation.
//
// EN: Key points: importing rewrites the actor's own service list, so it stays a top-level tool: the admin and the
// EN: directors may use it, each one limited to the archives they are allowed to see anyway.
func canCopyArchiveServices(actor *User, authorRole string, authorUsername string) bool {
	if actor == nil {
		return false
	}
	if !roleCanCopyArchiveServices(actor.Role) {
		return false
	}
	return canViewArchiveRole(actor, authorRole, authorUsername)
}

// EN: Function `roleCanCopyArchiveServices`.
//
// EN: What it does: roleCanCopyArchiveServices tells whether a role may import services from the archive at all.
//
// EN: Key points: the admin and every director (level 4) qualify; which archives they actually reach is still
// EN: decided per calculation by canViewArchiveRole.
func roleCanCopyArchiveServices(role string) bool {
	role = normalizeRole(role)
	return role == RoleAdmin || roleLevel(role) >= 4
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
