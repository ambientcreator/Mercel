package appcore_test

import (
	. "statistic/appcore"
	"strings"
	"testing"
	"time"
)

func sampleExportItems() []CalculationItem {
	return []CalculationItem{{
		Name:      "\u0423\u0441\u043b\u0443\u0433\u0430",
		Unit:      "\u0447.",
		Rate:      100,
		Quantity:  1,
		LineTotal: 100,
	}}
}

func TestBuildActPDFDataRequiresActNumber(t *testing.T) {
	_, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:        0,
		EmployeeFullName: "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447",
		ContractCode:     "2",
		ContractNumber:   "15",
		ContractDate:     "2026-01-30",
		TargetAmount:     100,
		Items:            sampleExportItems(),
	}, &User{})
	if err == nil || !strings.Contains(err.Error(), "\u043d\u043e\u043c\u0435\u0440 \u0430\u043a\u0442\u0430") {
		t.Fatalf("expected act number validation error, got %v", err)
	}
}

func TestBuildActPDFDataRequiresFullName(t *testing.T) {
	_, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:      1,
		ContractCode:   "2",
		ContractNumber: "15",
		ContractDate:   "2026-01-30",
		TargetAmount:   100,
		Items:          sampleExportItems(),
	}, &User{})
	if err == nil || !strings.Contains(err.Error(), "\u0424\u0418\u041e \u0441\u043e\u0442\u0440\u0443\u0434\u043d\u0438\u043a\u0430") {
		t.Fatalf("expected full name validation error, got %v", err)
	}
}

func TestBuildActPDFDataRequiresContractNumber(t *testing.T) {
	_, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:        1,
		EmployeeFullName: "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447",
		ContractCode:     "2",
		ContractDate:     "2026-01-30",
		TargetAmount:     100,
		Items:            sampleExportItems(),
	}, &User{})
	if err == nil || !strings.Contains(err.Error(), "\u0434\u043e\u0433\u043e\u0432\u043e\u0440") {
		t.Fatalf("expected contract validation error, got %v", err)
	}
}

func TestBuildActPDFDataRejectsUnknownContract(t *testing.T) {
	_, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:        1,
		EmployeeFullName: "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447",
		ContractCode:     "999",
		ContractNumber:   "15",
		ContractDate:     "2026-01-30",
		TargetAmount:     100,
		Items:            sampleExportItems(),
	}, &User{})
	if err == nil || !strings.Contains(err.Error(), "\u0412\u044b\u0431\u0435\u0440\u0438\u0442\u0435 \u0434\u043e\u0433\u043e\u0432\u043e\u0440") {
		t.Fatalf("expected unknown contract validation error, got %v", err)
	}
}

func TestBuildActPDFDataRequiresContractDate(t *testing.T) {
	_, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:        1,
		EmployeeFullName: "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447",
		ContractCode:     "2",
		ContractNumber:   "15",
		TargetAmount:     100,
		Items:            sampleExportItems(),
	}, &User{})
	if err == nil || !strings.Contains(err.Error(), "\u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430") {
		t.Fatalf("expected contract date validation error, got %v", err)
	}
}

func TestBuildActPDFDataRequiresExactTotal(t *testing.T) {
	_, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:        1,
		EmployeeFullName: "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447",
		ContractCode:     "2",
		ContractNumber:   "15",
		ContractDate:     "2026-01-30",
		TargetAmount:     150,
		Items:            sampleExportItems(),
	}, &User{})
	if err == nil || !strings.Contains(err.Error(), "\u0442\u043e\u043b\u044c\u043a\u043e \u0434\u043b\u044f \u0442\u043e\u0447\u043d\u043e\u0433\u043e \u0440\u0430\u0441\u0447\u0451\u0442\u0430") {
		t.Fatalf("expected exact-total validation error, got %v", err)
	}
}

func TestBuildActPDFDataUsesCurrentUserFullNameFallback(t *testing.T) {
	generatedAt := time.Date(2026, time.March, 30, 12, 0, 0, 0, time.UTC)
	data, err := BuildActPDFDataForTest(ExportCalculationRequest{
		ActNumber:      7,
		ContractCode:   "2",
		ContractNumber: "15",
		ContractDate:   "2026-01-30",
		GeneratedAt:    generatedAt.Format(time.RFC3339),
		TargetAmount:   69960,
		Items: []CalculationItem{{
			Name:      "\u0423\u0441\u043b\u0443\u0433\u0430",
			Unit:      "\u0447.",
			Rate:      69960,
			Quantity:  1,
			LineTotal: 69960,
		}},
	}, &User{FullName: "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447"})
	if err != nil {
		t.Fatalf("buildActPDFData error = %v", err)
	}
	if data.EmployeeFullName != "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447" {
		t.Fatalf("expected fallback full name, got %q", data.EmployeeFullName)
	}
	if data.ContractNumber != "15" {
		t.Fatalf("expected contract number, got %q", data.ContractNumber)
	}
	if data.ContractTitle != "\u00ab\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c\u00bb \u21162" {
		t.Fatalf("expected contract title for 2, got %q", data.ContractTitle)
	}
	if data.CustomerName != "\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c" {
		t.Fatalf("expected customer name for 2, got %q", data.CustomerName)
	}
	if data.ContractDate.Format("2006-01-02") != "2026-01-30" {
		t.Fatalf("expected contract date 2026-01-30, got %s", data.ContractDate.Format("2006-01-02"))
	}
	if data.GeneratedAt.Format(time.RFC3339) != generatedAt.Format(time.RFC3339) {
		t.Fatalf("expected generatedAt %s, got %s", generatedAt.Format(time.RFC3339), data.GeneratedAt.Format(time.RFC3339))
	}
}

func TestResolveContractInfo(t *testing.T) {
	info, err := ResolveContractInfoForTest("1")
	if err != nil {
		t.Fatalf("resolveContractInfo error = %v", err)
	}
	if info.Title != "\u00ab\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438\u00bb \u21161" {
		t.Fatalf("unexpected title: %q", info.Title)
	}
}

func TestNumberToRussianWordsSimple(t *testing.T) {
	actual, err := NumberToRussianWordsForTest(69960)
	if err != nil {
		t.Fatalf("numberToRussianWords error = %v", err)
	}
	if actual != "\u0448\u0435\u0441\u0442\u044c\u0434\u0435\u0441\u044f\u0442 \u0434\u0435\u0432\u044f\u0442\u044c \u0442\u044b\u0441\u044f\u0447 \u0434\u0435\u0432\u044f\u0442\u044c\u0441\u043e\u0442 \u0448\u0435\u0441\u0442\u044c\u0434\u0435\u0441\u044f\u0442" {
		t.Fatalf("unexpected words: %q", actual)
	}
}

func TestUpdateUserContractDetailsPersistsPreferredContract(t *testing.T) {
	app := withTempDB(t)
	loginAsAdmin1(t, app)

	session := app.GetSession()
	if session.User == nil {
		t.Fatal("expected authenticated user session")
	}

	updated, err := app.UpdateUserContractDetails(UpdateUserContractDetailsRequest{
		UserID:                session.User.ID,
		FullName:              "\u0418\u0432\u0430\u043d\u043e\u0432 \u0418\u0432\u0430\u043d \u0418\u0432\u0430\u043d\u043e\u0432\u0438\u0447",
		PreferredContractCode: "2",
		ContractSPBKSNumber:   "11",
		ContractGrizablNumber: "22",
		ContractSignedAt:      "2026-04-05",
	})
	if err != nil {
		t.Fatalf("UpdateUserContractDetails() error = %v", err)
	}
	if updated.PreferredContractCode != "2" {
		t.Fatalf("expected preferred contract code 2, got %q", updated.PreferredContractCode)
	}

	current := app.GetSession()
	if current.User == nil || current.User.PreferredContractCode != "2" {
		t.Fatalf("expected in-memory session preferred contract code 2, got %+v", current.User)
	}

	app.Logout()
	loginAsAdmin1(t, app)
	reloaded := app.GetSession()
	if reloaded.User == nil || reloaded.User.PreferredContractCode != "2" {
		t.Fatalf("expected persisted preferred contract code 2 after relogin, got %+v", reloaded.User)
	}
}
