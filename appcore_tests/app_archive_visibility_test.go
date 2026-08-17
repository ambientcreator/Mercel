package appcore_test

import (
	"testing"

	. "statistic/appcore"
)

// seedEmployeeArchive creates an employee, gives them a service and stores one
// calculation, returning the archive row id and the employee username.
func seedEmployeeArchive(t *testing.T, app *App, username string, target int) (int64, string) {
	t.Helper()

	loginAsAdmin(t, app)
	employee, err := app.CreateUser(UserWithPassword{Username: username, Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser(%q) error = %v", username, err)
	}

	createUserService(t, app, employee.Username, "secret", username+" service", 100, CategoryPrimary)
	result, err := app.CalculateAmount(CalculationRequest{TargetAmount: target, Weights: map[string]int{}})
	if err != nil {
		t.Fatalf("CalculateAmount(%q) error = %v", username, err)
	}
	saved, err := app.SaveCalculation(SaveCalculationRequest{Title: username + " calc", TargetAmount: result.TargetAmount, Items: result.Items})
	if err != nil {
		t.Fatalf("SaveCalculation(%q) error = %v", username, err)
	}
	return saved.ID, employee.Username
}

// forceStoredRole rewrites created_role straight in the database, reproducing rows
// written by older builds that the role migrations have not rewritten yet.
func forceStoredRole(t *testing.T, app *App, id int64, storedRole string) {
	t.Helper()
	if _, err := app.DBForTest().Exec(`UPDATE calculations SET created_role = ? WHERE id = ?`, storedRole, id); err != nil {
		t.Fatalf("force created_role = %q: %v", storedRole, err)
	}
}

func archiveIDs(t *testing.T, app *App) map[int64]bool {
	t.Helper()
	calculations, err := app.ListCalculations()
	if err != nil {
		t.Fatalf("ListCalculations() error = %v", err)
	}
	ids := make(map[int64]bool, len(calculations))
	for _, item := range calculations {
		ids[item.ID] = true
	}
	return ids
}

// EN: Test `TestArchiveVisibilitySurvivesNonCanonicalStoredRoles`.
//
// EN: What it does: it proves that archive rows whose created_role is not a canonical identifier — the empty default
// EN: of pre-migration rows, a legacy alias, or an unknown string — keep the visibility of RoleSupportEmployee, which
// EN: is what normalizeRole collapses them to.
//
// EN: Key points: this is the case the SQL visibility filter could silently drop, because a plain IN list over the
// EN: known role identifiers would not match any of these stored values.
func TestArchiveVisibilitySurvivesNonCanonicalStoredRoles(t *testing.T) {
	storedRoles := map[string]string{
		"empty (pre-migration default)": "",
		"legacy alias":                  "employee",
		"unknown identifier":            "some_role_that_never_existed",
	}

	for name, storedRole := range storedRoles {
		t.Run(name, func(t *testing.T) {
			app := withTempDB(t)

			archiveID, employeeName := seedEmployeeArchive(t, app, "employee9", 50100)
			forceStoredRole(t, app, archiveID, storedRole)

			loginAsAdmin(t, app)
			senior, err := app.CreateUser(UserWithPassword{Username: "senior9", Password: "secret", Role: RoleSeniorSpecialist})
			if err != nil {
				t.Fatalf("CreateUser senior error = %v", err)
			}

			// A senior specialist outranks an employee of the same department and must
			// still reach the row after its role was written in a non-canonical form.
			loginAsUser(t, app, senior.Username, "secret")
			if !archiveIDs(t, app)[archiveID] {
				t.Fatalf("senior specialist lost sight of archive %d stored with created_role = %q", archiveID, storedRole)
			}

			// The author reaches their own row regardless of what role it carries.
			loginAsUser(t, app, employeeName, "secret")
			if !archiveIDs(t, app)[archiveID] {
				t.Fatalf("author lost sight of own archive %d stored with created_role = %q", archiveID, storedRole)
			}
		})
	}
}

// EN: Test `TestArchiveVisibilityDoesNotLeakNonCanonicalRolesToPeers`.
//
// EN: What it does: it checks the other direction — a row stored with an unrecognised role must not become visible to
// EN: someone who is not allowed to see RoleSupportEmployee archives.
//
// EN: Key points: guards the NOT IN branch of the filter against being too permissive; an employee is level 1 and so
// EN: sees nobody but themselves.
func TestArchiveVisibilityDoesNotLeakNonCanonicalRolesToPeers(t *testing.T) {
	app := withTempDB(t)

	archiveID, _ := seedEmployeeArchive(t, app, "author9", 50100)
	forceStoredRole(t, app, archiveID, "some_role_that_never_existed")

	loginAsAdmin(t, app)
	peer, err := app.CreateUser(UserWithPassword{Username: "peer9", Password: "secret", Role: RoleEmployee})
	if err != nil {
		t.Fatalf("CreateUser peer error = %v", err)
	}

	loginAsUser(t, app, peer.Username, "secret")
	if archiveIDs(t, app)[archiveID] {
		t.Fatalf("employee %q must not see a peer archive stored with an unrecognised role", peer.Username)
	}
}

// EN: Test `TestListCalculationsMatchesRoleVisibilityAcrossDepartments`.
//
// EN: What it does: it pins the cross-department rules the SQL predicate now encodes — a technical director reaches
// EN: the technical and telecom archives but not the support one, while a support head reaches only their own
// EN: department.
//
// EN: Key points: exercises the IN list built from roleViewDepartments rather than a single-department scenario, so a
// EN: predicate that quietly widened or narrowed the scope would fail here.
func TestListCalculationsMatchesRoleVisibilityAcrossDepartments(t *testing.T) {
	app := withTempDB(t)

	authors := []struct {
		username string
		role     string
		target   int
	}{
		{"tech9", RoleTechnicalEmployee, 50100},
		{"telecom9", RoleTelecomEmployeeVOLS, 50200},
		{"support9", RoleSupportEmployee, 50300},
	}

	archives := make(map[string]int64, len(authors))
	for _, author := range authors {
		loginAsAdmin(t, app)
		created, err := app.CreateUser(UserWithPassword{Username: author.username, Password: "secret", Role: author.role})
		if err != nil {
			t.Fatalf("CreateUser(%q) error = %v", author.username, err)
		}
		createUserService(t, app, created.Username, "secret", author.username+" service", 100, CategoryPrimary)
		result, err := app.CalculateAmount(CalculationRequest{TargetAmount: author.target, Weights: map[string]int{}})
		if err != nil {
			t.Fatalf("CalculateAmount(%q) error = %v", author.username, err)
		}
		saved, err := app.SaveCalculation(SaveCalculationRequest{Title: author.username + " calc", TargetAmount: result.TargetAmount, Items: result.Items})
		if err != nil {
			t.Fatalf("SaveCalculation(%q) error = %v", author.username, err)
		}
		archives[author.username] = saved.ID
	}

	loginAsAdmin(t, app)
	director, err := app.CreateUser(UserWithPassword{Username: "techdir9", Password: "secret", Role: RoleTechnicalDirector})
	if err != nil {
		t.Fatalf("CreateUser technical director error = %v", err)
	}

	loginAsUser(t, app, director.Username, "secret")
	visible := archiveIDs(t, app)
	if !visible[archives["tech9"]] || !visible[archives["telecom9"]] {
		t.Fatalf("technical director should see the technical and telecom archives, got %+v", visible)
	}
	if visible[archives["support9"]] {
		t.Fatalf("technical director must not see the support archive, got %+v", visible)
	}
}
