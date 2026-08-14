package appcore_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	. "statistic/appcore"
)

func actTemplateExportRequest(items int) ExportCalculationRequest {
	exportItems := make([]CalculationItem, 0, items)
	total := 0
	for index := 0; index < items; index++ {
		lineTotal := 12345
		exportItems = append(exportItems, CalculationItem{
			Name:      fmt.Sprintf("%d. Диагностика, настройка и восстановление работоспособности абонентского оборудования связи", index+1),
			Unit:      "ч.",
			Rate:      4115,
			Quantity:  3,
			LineTotal: lineTotal,
		})
		total += lineTotal
	}
	return ExportCalculationRequest{
		ActNumber:        17,
		EmployeeFullName: "Иванов Иван Иванович",
		ContractCode:     "1",
		ContractNumber:   "10/32-фе",
		ContractDate:     "2026-01-30",
		GeneratedAt:      "2026-03-30T12:00:00Z",
		TargetAmount:     total,
		Items:            exportItems,
	}
}

func pdfPageCount(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	matches := regexp.MustCompile(`/Type /Pages[^>]*?/Count (\d+)`).FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		matches = regexp.MustCompile(`/Count (\d+)`).FindAllStringSubmatch(string(raw), -1)
	}
	if len(matches) == 0 {
		t.Fatalf("no page count found in %s", path)
	}
	count, err := strconv.Atoi(matches[len(matches)-1][1])
	if err != nil {
		t.Fatalf("parse page count: %v", err)
	}
	return count
}

func TestCreateUserAssignsRandomActTemplate(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	seen := make(map[string]int)
	for index := 0; index < 24; index++ {
		user, err := app.CreateUser(UserWithPassword{
			Username: fmt.Sprintf("blankuser%d", index),
			Password: "Str0ng!Passw0rd",
			Role:     RoleTechnicalEmployee,
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		if NormalizeActTemplateForTest(user.ActTemplate) == "" {
			t.Fatalf("expected a known act template, got %q", user.ActTemplate)
		}
		seen[user.ActTemplate]++
	}
	if len(seen) < 2 {
		t.Fatalf("expected employees to be spread across blanks, got %+v", seen)
	}
}

func TestActTemplateAssignmentStaysBalanced(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	counts := make(map[string]int)
	// The admin account already holds a blank, so count it in as well.
	admin := app.GetSession()
	if admin.User == nil {
		t.Fatal("expected an authenticated admin session")
	}
	counts[admin.User.ActTemplate]++

	for index := 0; index < 12; index++ {
		user, err := app.CreateUser(UserWithPassword{
			Username: fmt.Sprintf("balanced%d", index),
			Password: "Str0ng!Passw0rd",
			Role:     RoleTechnicalEmployee,
		})
		if err != nil {
			t.Fatalf("CreateUser() error = %v", err)
		}
		counts[user.ActTemplate]++
	}

	least := -1
	most := 0
	for _, id := range ActTemplateIDsForTest() {
		if least < 0 || counts[id] < least {
			least = counts[id]
		}
		if counts[id] > most {
			most = counts[id]
		}
	}
	if most-least > 1 {
		t.Fatalf("expected an even split across blanks, got %+v", counts)
	}
}

func TestUpdateUserActTemplateChangesOwnBlank(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	session := app.GetSession()
	if session.User == nil {
		t.Fatal("expected an authenticated admin session")
	}

	target := ""
	for _, id := range ActTemplateIDsForTest() {
		if id != session.User.ActTemplate {
			target = id
			break
		}
	}
	if target == "" {
		t.Fatal("expected at least two blanks to switch between")
	}

	updated, err := app.UpdateUserActTemplate(UpdateUserActTemplateRequest{UserID: session.User.ID, ActTemplate: target})
	if err != nil {
		t.Fatalf("UpdateUserActTemplate() error = %v", err)
	}
	if updated.ActTemplate != target {
		t.Fatalf("expected blank %q, got %q", target, updated.ActTemplate)
	}
	if current := app.GetSession(); current.User == nil || current.User.ActTemplate != target {
		t.Fatalf("expected the session to carry blank %q, got %+v", target, current.User)
	}

	app.Logout()
	loginAsAdmin(t, app)
	if reloaded := app.GetSession(); reloaded.User == nil || reloaded.User.ActTemplate != target {
		t.Fatalf("expected blank %q after relogin, got %+v", target, reloaded.User)
	}
}

func TestUpdateUserActTemplateRejectsUnknownBlank(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	session := app.GetSession()
	if session.User == nil {
		t.Fatal("expected an authenticated admin session")
	}
	if _, err := app.UpdateUserActTemplate(UpdateUserActTemplateRequest{UserID: session.User.ID, ActTemplate: "42"}); err == nil {
		t.Fatal("expected an unknown blank to be rejected")
	}
}

func TestUpdateUserActTemplateRejectsForeignEmployee(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	first, err := app.CreateUser(UserWithPassword{Username: "blankowner", Password: "Str0ng!Passw0rd", Role: RoleTechnicalEmployee})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if _, err := app.CreateUser(UserWithPassword{Username: "blankstranger", Password: "Str0ng!Passw0rd", Role: RoleTechnicalEmployee}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	loginAsUser(t, app, "blankstranger", "Str0ng!Passw0rd")
	if _, err := app.UpdateUserActTemplate(UpdateUserActTemplateRequest{UserID: first.ID, ActTemplate: ActTemplateTypographic}); err == nil {
		t.Fatal("expected an employee to be unable to change somebody else's blank")
	}
}

func TestListActTemplatesReturnsEveryBlank(t *testing.T) {
	app := withTempDB(t)
	options := app.ListActTemplates()
	if len(options) != len(ActTemplateIDsForTest()) {
		t.Fatalf("expected %d blanks, got %d", len(ActTemplateIDsForTest()), len(options))
	}
	for index, id := range ActTemplateIDsForTest() {
		if options[index].ID != id {
			t.Fatalf("expected blank %q at position %d, got %q", id, index, options[index].ID)
		}
		if strings.TrimSpace(options[index].Label) == "" {
			t.Fatalf("expected a label for blank %q", id)
		}
	}
}

func TestActTemplateStaysWithTheEmployee(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	created, err := app.CreateUser(UserWithPassword{
		Username: "blankstable",
		Password: "Str0ng!Passw0rd",
		Role:     RoleTechnicalEmployee,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	loginAsUser(t, app, "blankstable", "Str0ng!Passw0rd")
	session := app.GetSession()
	if session.User == nil || session.User.ActTemplate != created.ActTemplate {
		t.Fatalf("expected session blank %q, got %+v", created.ActTemplate, session.User)
	}

	app.Logout()
	loginAsUser(t, app, "blankstable", "Str0ng!Passw0rd")
	reloaded := app.GetSession()
	if reloaded.User == nil || reloaded.User.ActTemplate != created.ActTemplate {
		t.Fatalf("expected blank %q after relogin, got %+v", created.ActTemplate, reloaded.User)
	}
}

func TestEnsureUserActTemplateAssignsLegacyAccounts(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	created, err := app.CreateUser(UserWithPassword{
		Username: "blanklegacy",
		Password: "Str0ng!Passw0rd",
		Role:     RoleTechnicalEmployee,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if _, err := app.DBForTest().Exec(`UPDATE users SET act_template = '' WHERE id = ?`, created.ID); err != nil {
		t.Fatalf("clear act template: %v", err)
	}

	assigned, err := app.EnsureUserActTemplateForTest(created.ID, "")
	if err != nil {
		t.Fatalf("EnsureUserActTemplateForTest() error = %v", err)
	}
	if NormalizeActTemplateForTest(assigned) == "" {
		t.Fatalf("expected a known act template, got %q", assigned)
	}

	stored, err := app.GetUserByIDForTest(created.ID)
	if err != nil {
		t.Fatalf("GetUserByIDForTest() error = %v", err)
	}
	if stored.ActTemplate != assigned {
		t.Fatalf("expected persisted blank %q, got %q", assigned, stored.ActTemplate)
	}

	again, err := app.EnsureUserActTemplateForTest(stored.ID, stored.ActTemplate)
	if err != nil {
		t.Fatalf("EnsureUserActTemplateForTest() second call error = %v", err)
	}
	if again != assigned {
		t.Fatalf("expected the blank to stay %q, got %q", assigned, again)
	}
}

func TestMigrationRewritesLegacyActTemplateIDs(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "legacy-blanks.sqlite")
	usePathResolvers(t, dbPath, filepath.Join(tempDir, "missing-legacy.sqlite"), "")

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if _, err := app.Login(LoginRequest{Username: "admin", Password: "siuW*R%wWkQS"}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	created, err := app.CreateUser(UserWithPassword{Username: "legacyblank", Password: "Str0ng!Passw0rd", Role: RoleTechnicalEmployee})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	// Values written by the very first version of the blank feature.
	if _, err := app.DBForTest().Exec(`UPDATE users SET act_template = '3' WHERE id = ?`, created.ID); err != nil {
		t.Fatalf("write legacy act template: %v", err)
	}
	if _, err := app.DBForTest().Exec(`UPDATE users SET act_template = '1' WHERE username = 'admin'`); err != nil {
		t.Fatalf("write legacy admin act template: %v", err)
	}
	if err := app.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	reopened, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() after migration error = %v", err)
	}
	t.Cleanup(func() {
		_ = reopened.Close()
	})

	migrated, err := reopened.GetUserByIDForTest(created.ID)
	if err != nil {
		t.Fatalf("GetUserByIDForTest() error = %v", err)
	}
	if migrated.ActTemplate != ActTemplateTypographic {
		t.Fatalf("expected legacy \"3\" to become %q, got %q", ActTemplateTypographic, migrated.ActTemplate)
	}

	if _, err := reopened.Login(LoginRequest{Username: "admin", Password: "siuW*R%wWkQS"}); err != nil {
		t.Fatalf("Login() after migration error = %v", err)
	}
	session := reopened.GetSession()
	if session.User == nil || session.User.ActTemplate != ActTemplateStandard {
		t.Fatalf("expected legacy \"1\" to become %q, got %+v", ActTemplateStandard, session.User)
	}
}

func TestNormalizeActTemplateRejectsUnknownValues(t *testing.T) {
	for _, id := range ActTemplateIDsForTest() {
		if NormalizeActTemplateForTest(id) != id {
			t.Fatalf("expected %q to be a known blank", id)
		}
		if strings.TrimSpace(ActTemplateLabelForTest(id)) == "" {
			t.Fatalf("expected a label for blank %q", id)
		}
	}
	for _, value := range []string{"", " ", "0", "2", "typographic", "blank2"} {
		if NormalizeActTemplateForTest(value) != "" {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestEveryActTemplateRendersOnASinglePage(t *testing.T) {
	for _, template := range ActTemplateIDsForTest() {
		for _, items := range []int{1, 5, 10} {
			path := filepath.Join(t.TempDir(), fmt.Sprintf("act_%s_%d.pdf", template, items))
			err := RenderActPDFForTest(path, actTemplateExportRequest(items), &User{
				FullName:    "Иванов Иван Иванович",
				ActTemplate: template,
			})
			if err != nil {
				t.Fatalf("RenderActPDFForTest(blank %s, %d items) error = %v", template, items, err)
			}
			if pages := pdfPageCount(t, path); pages != 1 {
				t.Fatalf("blank %s with %d items rendered %d pages, want 1", template, items, pages)
			}
		}
	}
}

func TestActTemplatesProduceDifferentDocuments(t *testing.T) {
	dir := t.TempDir()
	seen := make(map[string]string)
	for _, template := range ActTemplateIDsForTest() {
		path := filepath.Join(dir, fmt.Sprintf("act_%s.pdf", template))
		if err := RenderActPDFForTest(path, actTemplateExportRequest(4), &User{
			FullName:    "Иванов Иван Иванович",
			ActTemplate: template,
		}); err != nil {
			t.Fatalf("RenderActPDFForTest(blank %s) error = %v", template, err)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read pdf: %v", err)
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(raw))
		if other, exists := seen[digest]; exists {
			t.Fatalf("blanks %s and %s render byte-identical documents", other, template)
		}
		seen[digest] = template
	}
	if len(seen) != len(ActTemplateIDsForTest()) {
		t.Fatalf("expected %d distinct documents, got %d", len(ActTemplateIDsForTest()), len(seen))
	}
}

func TestActTemplateFontsAreDistinctPerBlank(t *testing.T) {
	fonts := map[string]string{}
	for _, template := range ActTemplateIDsForTest() {
		fonts[template] = ActTemplateFontForTest(template)
	}
	if fonts[ActTemplateStandard] == fonts[ActTemplateFormal] {
		t.Fatalf("the standard blank and blank №1 must not share a font family, both use %q", fonts[ActTemplateStandard])
	}
	if fonts[ActTemplateContract] == fonts[ActTemplateFormal] {
		t.Fatalf("blank №4 and blank №1 must not share a font family, both use %q", fonts[ActTemplateContract])
	}
	for template, family := range fonts {
		if strings.TrimSpace(family) == "" {
			t.Fatalf("blank %s has no font family", template)
		}
	}
}
