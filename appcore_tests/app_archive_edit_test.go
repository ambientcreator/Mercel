package appcore_test

import (
	"path/filepath"
	"testing"

	. "statistic/appcore"
)

func TestAdminCanUpdateArchivedCalculation(t *testing.T) {
	tempDir := t.TempDir()
	usePathResolvers(t, filepath.Join(tempDir, "test.sqlite"), filepath.Join(tempDir, "legacy.sqlite"), filepath.Join(tempDir, "previous.sqlite"))

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	t.Cleanup(func() {
		_ = app.Close()
	})

	loginAsAdmin(t, app)

	saved, err := app.SaveCalculation(SaveCalculationRequest{
		TargetAmount: 550,
		Items: []CalculationItem{
			{ServiceCode: "svc-1", Name: "Старая услуга", Unit: "ч.", Rate: 100, Quantity: 5, LineTotal: 500, Category: CategoryPrimary, Weight: 0},
			{ServiceCode: "svc-2", Name: "Добивка", Unit: "шт.", Rate: 50, Quantity: 1, LineTotal: 50, Category: CategoryClosing, Weight: 0},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation() error = %v", err)
	}

	percent := 12.5
	updated, err := app.UpdateCalculationAsAdmin(UpdateSavedCalculationRequest{
		ID: saved.ID,
		Items: []CalculationItem{
			{ServiceCode: "svc-1", Name: "Новая услуга", Unit: "ч.", Rate: 120, Quantity: 4, Category: CategoryPrimary, Weight: 3, AllocationPercent: &percent},
			{ServiceCode: "svc-2", Name: "Добивка", Unit: "шт.", Rate: 20, Quantity: 2, Category: CategoryClosing, Weight: -1},
		},
	})
	if err != nil {
		t.Fatalf("UpdateCalculationAsAdmin() error = %v", err)
	}

	if updated.TotalAmount != 520 {
		t.Fatalf("updated.TotalAmount = %d, want 520", updated.TotalAmount)
	}
	// The edit changes what was delivered, not what was originally asked for, so the
	// target the author saved with has to survive it.
	if updated.TargetAmount != 550 {
		t.Fatalf("updated.TargetAmount = %d, want the original 550 preserved", updated.TargetAmount)
	}
	if len(updated.Items) != 2 {
		t.Fatalf("len(updated.Items) = %d, want 2", len(updated.Items))
	}
	if updated.Items[0].Name != "Новая услуга" {
		t.Fatalf("updated.Items[0].Name = %q, want %q", updated.Items[0].Name, "Новая услуга")
	}
	if updated.Items[1].Weight != -1 {
		t.Fatalf("updated.Items[1].Weight = %d, want -1", updated.Items[1].Weight)
	}

	list, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	var found *SavedCalculation
	for i := range list {
		if list[i].ID == saved.ID {
			found = &list[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("updated calculation with id %d not found in archive list", saved.ID)
	}
	if found.Items[0].Name != "Новая услуга" {
		t.Fatalf("found.Items[0].Name = %q, want %q", found.Items[0].Name, "Новая услуга")
	}
	if found.TotalAmount != 520 {
		t.Fatalf("found.TotalAmount = %d, want 520", found.TotalAmount)
	}
	if found.TargetAmount != 550 {
		t.Fatalf("found.TargetAmount = %d, want the original 550 persisted", found.TargetAmount)
	}
}

// EN: Test `TestSaveCalculationKeepsOneRowPerAuthorPerDay`.
//
// EN: What it does: it saves twice on the same day under different titles and requires the second save to replace the
// EN: first instead of adding a row.
//
// EN: Key points: the day is keyed off created_at, not off the caller-supplied title, so a request sending its own
// EN: title can no longer store a second archive row for a day that is meant to hold one.
func TestSaveCalculationKeepsOneRowPerAuthorPerDay(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	first, err := app.SaveCalculation(SaveCalculationRequest{
		Title:        "утренний расчёт",
		TargetAmount: 500,
		Items: []CalculationItem{
			{ServiceCode: "svc-1", Name: "Услуга", Unit: "ч.", Rate: 100, Quantity: 5, LineTotal: 500, Category: CategoryPrimary},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation() first error = %v", err)
	}

	second, err := app.SaveCalculation(SaveCalculationRequest{
		Title:        "вечерний расчёт",
		TargetAmount: 700,
		Items: []CalculationItem{
			{ServiceCode: "svc-1", Name: "Услуга", Unit: "ч.", Rate: 100, Quantity: 7, LineTotal: 700, Category: CategoryPrimary},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation() second error = %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("second save created row %d instead of replacing row %d", second.ID, first.ID)
	}

	list, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	own := 0
	for _, item := range list {
		if item.CreatedBy == "admin" {
			own++
		}
	}
	if own != 1 {
		t.Fatalf("admin archive holds %d rows for today, want 1", own)
	}
	if list[0].TotalAmount != 700 {
		t.Fatalf("stored TotalAmount = %d, want the later 700", list[0].TotalAmount)
	}
	if list[0].Title != "вечерний расчёт" {
		t.Fatalf("stored Title = %q, want the later title", list[0].Title)
	}
}

// EN: Test `TestCopyArchiveServicesLeavesOwnListIntactWhenImportFails`.
//
// EN: What it does: it copies an archive holding a service the importer must reject and requires the caller to keep
// EN: every service they already had.
//
// EN: Key points: the import replaces the caller's whole list, so it deletes before it inserts; without validating
// EN: the batch up front and writing under one transaction, a rejected row left the caller with no services at all.
func TestCopyArchiveServicesLeavesOwnListIntactWhenImportFails(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin(t, app)

	before, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() before error = %v", err)
	}
	if len(before) == 0 {
		t.Fatal("expected the admin to start with seeded services")
	}

	// "километр" is not a unit the app supports, so the import has to reject this row.
	saved, err := app.SaveCalculation(SaveCalculationRequest{
		Title:        "архив с плохой единицей",
		TargetAmount: 300,
		Items: []CalculationItem{
			{ServiceCode: "svc-ok", Name: "Нормальная услуга", Unit: "ч.", Rate: 100, Quantity: 2, LineTotal: 200, Category: CategoryPrimary},
			{ServiceCode: "svc-bad", Name: "Кривая услуга", Unit: "километр", Rate: 100, Quantity: 1, LineTotal: 100, Category: CategoryPrimary},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation() error = %v", err)
	}

	if _, err := app.CopyArchiveServicesToAdmin(saved.ID); err == nil {
		t.Fatal("CopyArchiveServicesToAdmin() accepted an archive with an unsupported unit")
	}

	after, err := app.GetServices()
	if err != nil {
		t.Fatalf("GetServices() after error = %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("failed import changed the service list: had %d services, now %d", len(before), len(after))
	}
}

func TestArchiveDeleteHonorsOwnershipAndRoles(t *testing.T) {
	app := withTempDB(t)

	loginAsAdmin(t, app)
	if _, err := app.CreateUser(UserWithPassword{Username: "archive_employee_delete", Password: "secret", Role: RoleEmployee}); err != nil {
		t.Fatalf("CreateUser employee error = %v", err)
	}
	if _, err := app.CreateUser(UserWithPassword{Username: "archive_senior_delete", Password: "secret", Role: RoleSeniorSpecialist}); err != nil {
		t.Fatalf("CreateUser senior error = %v", err)
	}

	loginAsUser(t, app, "archive_employee_delete", "secret")
	employeeSaved, err := app.SaveCalculation(SaveCalculationRequest{
		TargetAmount: 100,
		Items: []CalculationItem{
			{Name: "Удаляемая услуга", Unit: "ч.", Rate: 100, Quantity: 1, LineTotal: 100, Category: CategoryPrimary},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation employee error = %v", err)
	}

	loginAsUser(t, app, "archive_senior_delete", "secret")
	if err := app.DeleteCalculation(employeeSaved.ID); err == nil {
		t.Fatalf("expected senior specialist to be denied when deleting another user's archive")
	}

	loginAsUser(t, app, "archive_employee_delete", "secret")
	if err := app.DeleteCalculation(employeeSaved.ID); err != nil {
		t.Fatalf("DeleteCalculation owner error = %v", err)
	}

	adminSaved, err := app.SaveCalculation(SaveCalculationRequest{
		TargetAmount: 200,
		Items: []CalculationItem{
			{Name: "Админская проверка", Unit: "ч.", Rate: 200, Quantity: 1, LineTotal: 200, Category: CategoryPrimary},
		},
	})
	if err != nil {
		t.Fatalf("SaveCalculation second employee error = %v", err)
	}

	loginAsAdmin(t, app)
	if err := app.DeleteCalculation(adminSaved.ID); err != nil {
		t.Fatalf("DeleteCalculation admin error = %v", err)
	}
}
