package appcore

import (
	"database/sql"
	"time"
)

type PathResolvers struct {
	DatabasePath         func() (string, error)
	LegacyDatabasePath   func() (string, error)
	PreviousMercelDBPath func() (string, error)
}

func OverridePathResolversForTest(resolvers PathResolvers) func() {
	originalDatabase := resolveDatabasePath
	originalLegacy := resolveLegacyDatabasePath
	originalPrevious := resolvePreviousMercelDatabasePath

	if resolvers.DatabasePath != nil {
		resolveDatabasePath = resolvers.DatabasePath
	}
	if resolvers.LegacyDatabasePath != nil {
		resolveLegacyDatabasePath = resolvers.LegacyDatabasePath
	}
	if resolvers.PreviousMercelDBPath != nil {
		resolvePreviousMercelDatabasePath = resolvers.PreviousMercelDBPath
	}

	return func() {
		resolveDatabasePath = originalDatabase
		resolveLegacyDatabasePath = originalLegacy
		resolvePreviousMercelDatabasePath = originalPrevious
	}
}

func (a *App) DBForTest() *sql.DB {
	return a.db
}

func (a *App) RequireAuthForTest() (*User, error) {
	return a.requireAuth()
}

func (a *App) GetUserByIDForTest(id int64) (User, error) {
	return a.getUserByID(id)
}

func HashPasswordForTest(password string) string {
	return hashPassword(password)
}

func LegacyHashPasswordForTest(password string) string {
	return legacyHashPassword(password)
}

func NumberToRussianWordsForTest(amount int) (string, error) {
	return numberToRussianWords(amount)
}

func RubleNounForTest(amount int) string {
	return rubleNoun(amount)
}

func BuildServiceTargetsForTest(byCategory map[string][]Service, categoryTargets map[string]int, weights map[string]int) map[string]int {
	return buildServiceTargets(byCategory, categoryTargets, weights)
}

func RoleDepartmentForTest(role string) string {
	return roleDepartment(role)
}

func AbsIntForTest(value int) int {
	return absInt(value)
}

func BuildActPDFDataForTest(req ExportCalculationRequest, currentUser *User) (actPDFData, error) {
	return buildActPDFData(req, currentUser)
}

func ResolveContractInfoForTest(code string) (contractInfo, error) {
	return resolveContractInfo(code)
}

func ActTemplateIDsForTest() []string {
	return append([]string(nil), actTemplateIDs...)
}

func NormalizeActTemplateForTest(value string) string {
	return normalizeActTemplate(value)
}

func ActTemplateLabelForTest(value string) string {
	return actTemplateLabel(value)
}

func (a *App) EnsureUserActTemplateForTest(userID int64, current string) (string, error) {
	return a.ensureUserActTemplate(userID, current)
}

func ActTemplateFontForTest(template string) string {
	return actTemplateFont(template)
}

func RoleCanCopyArchiveServicesForTest(role string) bool {
	return roleCanCopyArchiveServices(role)
}

func RenderActPDFForTest(path string, req ExportCalculationRequest, currentUser *User) error {
	data, err := buildActPDFData(req, currentUser)
	if err != nil {
		return err
	}
	return renderActPDF(path, data)
}

func ArchiveDateTitleForTest(now time.Time) string {
	return archiveDateTitle(now)
}
