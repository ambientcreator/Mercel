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
	if updated.TargetAmount != 520 {
		t.Fatalf("updated.TargetAmount = %d, want 520", updated.TargetAmount)
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
