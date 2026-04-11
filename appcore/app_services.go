package appcore

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func formatMoney(value int) string {
	return displayMoney(value)
}

// EN: Function `displayMoney`.
//
// EN: What it does: displayMoney is a UI-facing helper that mirrors money formatting where a separate name reads clearer.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func displayMoney(value int) string {
	return fmt.Sprintf("%d \u0440", value)
}

func archiveDateTitle(now time.Time) string {
	return now.Format("02.01.2006")
}

// EN: Method `saveServiceForOwner`.
//
// EN: What it does: saveServiceForOwner performs the actual insert/update of a service for a concrete user owner.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) saveServiceForOwner(req UpsertServiceRequest, owner string) (Service, error) {
	name := strings.TrimSpace(req.Name)
	unit := normalizeUnit(req.Unit)
	if err := validateServiceName(name); err != nil {
		return Service{}, err
	}
	if unit == "" || req.Rate <= 0 || req.Rate > 100000000 {
		return Service{}, errors.New("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043a\u043e\u0440\u0440\u0435\u043a\u0442\u043d\u044b\u0435 \u0435\u0434\u0438\u043d\u0438\u0446\u0443 \u0438 \u0441\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c \u0443\u0441\u043b\u0443\u0433\u0438.")
	}
	category := normalizeCategory(req.Category)
	var allocation interface{}
	if req.AllocationPercent != nil {
		if *req.AllocationPercent < 0 || *req.AllocationPercent > 100 {
			return Service{}, errors.New("\u041f\u0440\u043e\u0446\u0435\u043d\u0442 \u0443\u0441\u043b\u0443\u0433\u0438 \u0434\u043e\u043b\u0436\u0435\u043d \u0431\u044b\u0442\u044c \u0432 \u0434\u0438\u0430\u043f\u0430\u0437\u043e\u043d\u0435 \u043e\u0442 0 \u0434\u043e 100.")
		}
		allocation = *req.AllocationPercent
	}
	code := generateServiceCode(owner + "-" + name)
	createdAt := time.Now().Format(time.RFC3339)

	if req.ID == 0 {
		result, err := a.db.Exec(`INSERT INTO services(code, name, unit, rate, category, allocation_percent, created_by, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, code, name, unit, req.Rate, category, allocation, owner, createdAt)
		if err != nil {
			return Service{}, fmt.Errorf("create service: %w", err)
		}
		id, _ := result.LastInsertId()
		return a.getServiceByIDForOwner(id, owner)
	}

	_, err := a.db.Exec(`UPDATE services SET name = ?, unit = ?, rate = ?, category = ?, allocation_percent = ? WHERE id = ? AND created_by = ?`, name, unit, req.Rate, category, allocation, req.ID, owner)
	if err != nil {
		return Service{}, fmt.Errorf("update service: %w", err)
	}
	return a.getServiceByIDForOwner(req.ID, owner)
}

func (a *App) saveService(req UpsertServiceRequest) (Service, error) {
	user, err := a.requireAuth()
	if err != nil {
		return Service{}, err
	}
	return a.saveServiceForOwner(req, user.Username)
}

// EN: Method `getServiceByIDForOwner`.
//
// EN: What it does: getServiceByIDForOwner fetches one service while enforcing ownership boundaries.
//
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func (a *App) getServiceByIDForOwner(id int64, owner string) (Service, error) {
	var item Service
	var allocation sql.NullFloat64
	err := a.db.QueryRow(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE id = ? AND created_by = ?`, id, owner).Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt)
	if err != nil {
		return Service{}, err
	}
	item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
	if allocation.Valid {
		value := allocation.Float64
		item.AllocationPercent = &value
	}
	return item, nil
}

func (a *App) getServiceByNameForOwner(name string, owner string) (Service, error) {
	var item Service
	var allocation sql.NullFloat64
	err := a.db.QueryRow(`SELECT id, code, name, unit, rate, category, allocation_percent, created_by, created_at FROM services WHERE created_by = ? AND name = ? ORDER BY id ASC LIMIT 1`, owner, strings.TrimSpace(name)).Scan(&item.ID, &item.Code, &item.Name, &item.Unit, &item.Rate, &item.Category, &allocation, &item.CreatedBy, &item.CreatedAt)
	if err != nil {
		return Service{}, err
	}
	item.Description = fmt.Sprintf("1 %s = %s", strings.TrimSuffix(item.Unit, "."), displayMoney(item.Rate))
	if allocation.Valid {
		value := allocation.Float64
		item.AllocationPercent = &value
	}
	return item, nil
}

// EN: Method `UpsertService`.
//
// EN: What it does: UpsertService is the public service create/update API used by the settings screen.
//
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) UpsertService(req UpsertServiceRequest) (Service, error) {
	if _, err := a.requireAuth(); err != nil {
		return Service{}, err
	}
	return a.saveService(req)
}

// EN: Method `DeleteService`.
//
// EN: What it does: DeleteService removes one owned service after authorization and ownership checks.
//
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
func (a *App) DeleteService(id int64) error {
	user, err := a.requireAuth()
	if err != nil {
		return err
	}
	if id <= 0 {
		return errors.New("Некорректный идентификатор услуги.")
	}
	_, err = a.db.Exec(`DELETE FROM services WHERE id = ? AND created_by = ?`, id, user.Username)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	return nil
}

// EN: Method `CreateUser`.
//
// EN: What it does: CreateUser creates a subordinate account according to the current actor permissions.
//
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
