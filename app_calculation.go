package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"
)

func (a *App) CalculateAmount(req CalculationRequest) (CalculationResult, error) {
	if _, err := a.requireAuth(); err != nil {
		return CalculationResult{}, err
	}
	if req.TargetAmount <= 0 {
		return CalculationResult{}, errors.New("Введите сумму больше 0.")
	}
	if req.TargetAmount > 1_500_000 {
		return CalculationResult{}, fmt.Errorf("Сумма %d р превышает допустимый предел 1500000 р.", req.TargetAmount)
	}

	services, err := a.GetServices()
	if err != nil {
		return CalculationResult{}, err
	}
	if len(services) == 0 {
		return CalculationResult{}, errors.New("Сначала создайте хотя бы одну услугу.")
	}

	activeServices, activeIndexes := filterCalculableServices(services)
	if len(activeServices) == 0 {
		return CalculationResult{}, errors.New("Нет услуг, участвующих в расчёте. Уберите нулевой процент у нужных позиций или добавьте активные услуги.")
	}

	weights := normalizeWeights(req.Weights, services)
	quantities, ok := solveStructuredAllocation(req.TargetAmount, activeServices, weights)
	if !ok {
		quantities, ok = solveExact(req.TargetAmount, activeServices, weights)
	}
	if !ok {
		return CalculationResult{}, fmt.Errorf("Не удалось подобрать точный расчёт на сумму %s . Измените сумму, проценты или набор активных услуг.", displayMoney(req.TargetAmount))
	}

	fullQuantities := make([]int, len(services))
	for idx, quantity := range quantities {
		fullQuantities[activeIndexes[idx]] = quantity
	}

	items := make([]CalculationItem, 0, len(services))
	total := 0
	active := 0
	for idx, service := range services {
		quantity := fullQuantities[idx]
		lineTotal := service.Rate * quantity
		if quantity > 0 {
			active++
		}
		total += lineTotal
		items = append(items, CalculationItem{
			ServiceID:         service.ID,
			ServiceCode:       service.Code,
			Name:              service.Name,
			Unit:              service.Unit,
			Rate:              service.Rate,
			Quantity:          quantity,
			LineTotal:         lineTotal,
			Description:       service.Description,
			Weight:            weights[service.Code],
			Category:          service.Category,
			AllocationPercent: service.AllocationPercent,
		})
	}
	if total != req.TargetAmount {
		return CalculationResult{}, fmt.Errorf("Не удалось подобрать точный расчёт на сумму %s . Измените сумму, проценты или набор активных услуг.", displayMoney(req.TargetAmount))
	}

	return CalculationResult{
		TargetAmount:   req.TargetAmount,
		TotalAmount:    total,
		Items:          items,
		FoundExact:     true,
		GeneratedAt:    time.Now().Format(time.RFC3339),
		ActiveServices: active,
		Weights:        weights,
	}, nil
}

// RU: Функция `normalizeWeights`.
// EN: Function `normalizeWeights`.
//
// RU: Что делает: нормализует и ограничивает веса, пришедшие с фронтенда, по текущему списку услуг.
// EN: What it does: normalizes and clamps frontend-provided weights against the current service list.
//
// RU: Ключевые моменты: не создаёт лишние услуги; для отсутствующих весов оставляет значение 0.
// EN: Key points: does not create extra services; keeps missing weights at zero.
func normalizeWeights(input map[string]int, services []Service) map[string]int {
	result := make(map[string]int, len(services))
	for _, service := range services {
		result[service.Code] = 0
	}
	for code, weight := range input {
		result[code] = clampWeight(weight)
	}
	return result
}

func filterCalculableServices(services []Service) ([]Service, []int) {
	filtered := make([]Service, 0, len(services))
	indexes := make([]int, 0, len(services))
	for idx, service := range services {
		if serviceExcludedFromCalculation(service) {
			continue
		}
		filtered = append(filtered, service)
		indexes = append(indexes, idx)
	}
	return filtered, indexes
}

func serviceExcludedFromCalculation(service Service) bool {
	if service.AllocationPercent == nil {
		return false
	}
	return *service.AllocationPercent <= 0
}

// RU: Р¤СѓРЅРєС†РёСЏ `solveStructuredAllocation`.
// EN: Function `solveStructuredAllocation`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: solveStructuredAllocation tries the preferred group-based solver before falling back to a generic exact search.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func solveStructuredAllocation(target int, services []Service, weights map[string]int) ([]int, bool) {
	baseQuantities, reducedTarget, ok := reserveMinimumServices(target, services)
	if !ok {
		return nil, false
	}
	byCategory := map[string][]Service{
		CategoryPrimary:   {},
		CategorySecondary: {},
		CategoryClosing:   {},
	}
	indexMap := map[string][]int{
		CategoryPrimary:   {},
		CategorySecondary: {},
		CategoryClosing:   {},
	}
	for idx, service := range services {
		category := normalizeCategory(service.Category)
		byCategory[category] = append(byCategory[category], service)
		indexMap[category] = append(indexMap[category], idx)
	}

	groupTargets, ok := chooseGroupTargets(reducedTarget, byCategory)
	if !ok {
		return nil, false
	}
	targets := buildServiceTargets(byCategory, groupTargets, weights)
	result := append([]int(nil), baseQuantities...)
	for _, category := range []string{CategoryPrimary, CategorySecondary, CategoryClosing} {
		group := byCategory[category]
		if len(group) == 0 {
			continue
		}
		serviceTargets := make([]int, len(group))
		for i, service := range group {
			serviceTargets[i] = targets[service.Code]
		}
		allocation, ok := bestGroupAllocation(groupTargets[category], group, serviceTargets, weights)
		if !ok {
			return nil, false
		}
		mergeGroupQuantities(result, allocation.quantities, indexMap[category])
	}

	total := 0
	for idx, quantity := range result {
		total += quantity * services[idx].Rate
	}
	if total != target {
		return nil, false
	}
	return result, true
}

// RU: Р¤СѓРЅРєС†РёСЏ `chooseGroupTargets`.
// EN: Function `chooseGroupTargets`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: chooseGroupTargets picks category subtotals close to configured percentages while staying exactly reachable.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func chooseGroupTargets(target int, byCategory map[string][]Service) (map[string]int, bool) {
	defaults := defaultGroupPercent()
	categoryTargets := allocateByRatios(target, []float64{defaults[CategoryPrimary], defaults[CategorySecondary], defaults[CategoryClosing]})
	desired := map[string]int{
		CategoryPrimary:   categoryTargets[0],
		CategorySecondary: categoryTargets[1],
		CategoryClosing:   categoryTargets[2],
	}
	reachable := map[string][]bool{}
	for _, category := range []string{CategoryPrimary, CategorySecondary, CategoryClosing} {
		reachable[category] = reachableAmounts(target, byCategory[category])
	}

	best := map[string]int{}
	bestScore := math.MaxInt
	windows := []int{500, 1500, 5000, 15000, target}
	for _, window := range windows {
		primaryCandidates := candidateAmounts(reachable[CategoryPrimary], desired[CategoryPrimary], window)
		secondaryCandidates := candidateAmounts(reachable[CategorySecondary], desired[CategorySecondary], window)
		for _, primary := range primaryCandidates {
			for _, secondary := range secondaryCandidates {
				closing := target - primary - secondary
				if closing < 0 || closing >= len(reachable[CategoryClosing]) || !reachable[CategoryClosing][closing] {
					continue
				}
				score := absInt(primary-desired[CategoryPrimary]) + absInt(secondary-desired[CategorySecondary]) + absInt(closing-desired[CategoryClosing])
				if score < bestScore {
					bestScore = score
					best[CategoryPrimary] = primary
					best[CategorySecondary] = secondary
					best[CategoryClosing] = closing
				}
			}
		}
		if bestScore != math.MaxInt {
			return best, true
		}
	}
	return nil, false
}

// RU: Р¤СѓРЅРєС†РёСЏ `reserveMinimumServices`.
// EN: Function `reserveMinimumServices`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: reserveMinimumServices reserves one unit per service when the target allows all services to stay active.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func reserveMinimumServices(target int, services []Service) ([]int, int, bool) {
	result := make([]int, len(services))
	if len(services) == 0 {
		return result, target, true
	}
	minimumTotal := 0
	for _, service := range services {
		minimumTotal += service.Rate
	}
	if minimumTotal > target {
		return result, target, true
	}
	for idx := range services {
		result[idx] = 1
	}
	return result, target - minimumTotal, true
}

// RU: Р¤СѓРЅРєС†РёСЏ `buildServiceTargets`.
// EN: Function `buildServiceTargets`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: buildServiceTargets distributes each category budget down to concrete services before exact solving.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func buildServiceTargets(byCategory map[string][]Service, categoryTargetMap map[string]int, weights map[string]int) map[string]int {
	categoryOrder := []string{CategoryPrimary, CategorySecondary, CategoryClosing}

	totalServices := 0
	for _, group := range byCategory {
		totalServices += len(group)
	}
	result := make(map[string]int, totalServices)
	for _, category := range categoryOrder {
		group := byCategory[category]
		if len(group) == 0 {
			continue
		}
		ratios := buildCategoryRatios(category, group, weights)
		allocated := allocateByRatios(categoryTargetMap[category], ratios)
		for idx, service := range group {
			result[service.Code] = allocated[idx]
		}
	}
	return result
}

// RU: Р¤СѓРЅРєС†РёСЏ `buildCategoryRatios`.
// EN: Function `buildCategoryRatios`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: buildCategoryRatios calculates initial per-service ratios inside one category.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func buildCategoryRatios(category string, group []Service, weights map[string]int) []float64 {
	groupPercent := defaultGroupPercent()[category] * 100.0
	ratios := make([]float64, len(group))
	explicitTotal := 0.0
	unassigned := 0
	for idx, service := range group {
		if service.AllocationPercent != nil {
			ratios[idx] = *service.AllocationPercent
			explicitTotal += *service.AllocationPercent
		} else {
			unassigned++
		}
	}

	remainingPercent := groupPercent - explicitTotal
	if remainingPercent < 0 {
		remainingPercent = 0
	}
	fallback := 0.0
	if unassigned > 0 {
		fallback = remainingPercent / float64(unassigned)
	}
	for idx, service := range group {
		if service.AllocationPercent == nil {
			ratios[idx] = fallback
		}
	}
	applyWeightAdjustments(ratios, group, weights)
	return ratios
}

// RU: Р¤СѓРЅРєС†РёСЏ `applyWeightAdjustments`.
// EN: Function `applyWeightAdjustments`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: applyWeightAdjustments shifts ratio share toward weighted services while preserving total category mass.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func applyWeightAdjustments(ratios []float64, group []Service, weights map[string]int) {
	if len(group) < 2 {
		return
	}
	for idx, service := range group {
		weight := clampWeight(weights[service.Code])
		if weight > 0 {
			for step := 0; step < weight; step++ {
				transferShareToTarget(ratios, idx, 2.0)
			}
			continue
		}
		for step := 0; step < -weight; step++ {
			transferShareFromTarget(ratios, idx, 2.0)
		}
	}
}

// RU: Р¤СѓРЅРєС†РёСЏ `transferShare`.
// EN: Function `transferShare`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: transferShare moves percentage share from peer services to one preferred target service.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func transferShareToTarget(ratios []float64, target int, share float64) {
	if share <= 0 {
		return
	}
	donors := make([]int, 0, len(ratios)-1)
	for idx, value := range ratios {
		if idx == target || value <= 0 {
			continue
		}
		donors = append(donors, idx)
	}
	remaining := share
	for remaining > 0.0001 && len(donors) > 0 {
		slice := remaining / float64(len(donors))
		nextDonors := make([]int, 0, len(donors))
		moved := 0.0
		for _, donor := range donors {
			take := math.Min(slice, ratios[donor])
			if take <= 0 {
				continue
			}
			ratios[donor] -= take
			moved += take
			if ratios[donor] > 0.0001 {
				nextDonors = append(nextDonors, donor)
			}
		}
		if moved <= 0 {
			break
		}
		ratios[target] += moved
		remaining -= moved
		donors = nextDonors
	}
}

func transferShareFromTarget(ratios []float64, target int, share float64) {
	if share <= 0 || target < 0 || target >= len(ratios) {
		return
	}
	receivers := make([]int, 0, len(ratios)-1)
	for idx := range ratios {
		if idx == target {
			continue
		}
		receivers = append(receivers, idx)
	}
	if len(receivers) == 0 || ratios[target] <= 0 {
		return
	}
	moved := math.Min(share, ratios[target])
	if moved <= 0 {
		return
	}
	ratios[target] -= moved
	slice := moved / float64(len(receivers))
	for _, receiver := range receivers {
		ratios[receiver] += slice
	}
}

func clampWeight(weight int) int {
	if weight < -10 {
		return -10
	}
	if weight > 10 {
		return 10
	}
	return weight
}

// RU: Р¤СѓРЅРєС†РёСЏ `reachableAmounts`.
// EN: Function `reachableAmounts`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: reachableAmounts marks which totals can be composed from the service rates of one category.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func reachableAmounts(limit int, services []Service) []bool {
	reachable := make([]bool, limit+1)
	reachable[0] = true
	for amount := 0; amount <= limit; amount++ {
		if !reachable[amount] {
			continue
		}
		for _, service := range services {
			next := amount + service.Rate
			if next <= limit {
				reachable[next] = true
			}
		}
	}
	return reachable
}

// RU: Р¤СѓРЅРєС†РёСЏ `candidateAmounts`.
// EN: Function `candidateAmounts`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: candidateAmounts collects reachable totals nearest to a desired subtotal for later scoring.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func candidateAmounts(reachable []bool, desired int, window int) []int {
	minAmount := desired - window
	if minAmount < 0 {
		minAmount = 0
	}
	maxAmount := desired + window
	if maxAmount >= len(reachable) {
		maxAmount = len(reachable) - 1
	}
	candidates := make([]int, 0)
	for amount := minAmount; amount <= maxAmount; amount++ {
		if reachable[amount] {
			candidates = append(candidates, amount)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return absInt(candidates[i]-desired) < absInt(candidates[j]-desired)
	})
	return candidates
}

// RU: Р¤СѓРЅРєС†РёСЏ `allocateByRatios`.
// EN: Function `allocateByRatios`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: allocateByRatios converts floating-point target shares into integer money targets that still sum exactly.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func allocateByRatios(total int, ratios []float64) []int {
	result := make([]int, len(ratios))
	if total <= 0 || len(ratios) == 0 {
		return result
	}

	sum := 0.0
	for _, ratio := range ratios {
		if ratio > 0 {
			sum += ratio
		}
	}
	if sum <= 0 {
		base := total / len(ratios)
		rest := total % len(ratios)
		for idx := range ratios {
			result[idx] = base
			if idx < rest {
				result[idx]++
			}
		}
		return result
	}

	type remainderPart struct {
		idx       int
		remainder float64
	}
	remainders := make([]remainderPart, 0, len(ratios))
	allocated := 0
	for idx, ratio := range ratios {
		normalized := 0.0
		if ratio > 0 {
			normalized = ratio / sum
		}
		exact := normalized * float64(total)
		base := int(exact)
		result[idx] = base
		allocated += base
		remainders = append(remainders, remainderPart{idx: idx, remainder: exact - float64(base)})
	}

	sort.SliceStable(remainders, func(i, j int) bool {
		return remainders[i].remainder > remainders[j].remainder
	})
	for idx := 0; idx < total-allocated; idx++ {
		result[remainders[idx%len(remainders)].idx]++
	}
	return result
}

// RU: Р¤СѓРЅРєС†РёСЏ `sumTargets`.
// EN: Function `sumTargets`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: sumTargets totals a slice of integer allocations.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func sumTargets(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

// RU: Р¤СѓРЅРєС†РёСЏ `bestGroupAllocation`.
// EN: Function `bestGroupAllocation`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: bestGroupAllocation finds the best exact quantities for one category under a target-share preference.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func bestGroupAllocation(target int, services []Service, targets []int, weights map[string]int) (groupAllocation, bool) {
	if len(services) == 0 {
		return groupAllocation{ok: target == 0, amount: 0, quantities: []int{}}, target == 0
	}
	type groupStage struct {
		score  int
		count  int
		prev   int
		qty    int
		exists bool
	}

	for _, margin := range []int{6, 14, 32, 80, 180, 400} {
		current := map[int]groupStage{
			0: {exists: true},
		}
		parents := make([]map[int]groupStage, len(services))

		for idx, service := range services {
			next := map[int]groupStage{}
			parents[idx] = map[int]groupStage{}
			targetQty := float64(targets[idx]) / float64(service.Rate)
			qtyMargin := margin + int(math.Ceil(math.Sqrt(targetQty+1)))
			minQty := 0
			maxQty := int(math.Ceil(targetQty)) + qtyMargin
			limitQty := target / service.Rate
			if maxQty > limitQty {
				maxQty = limitQty
			}
			for amount, state := range current {
				for qty := minQty; qty <= maxQty; qty++ {
					nextAmount := amount + qty*service.Rate
					if nextAmount > target {
						break
					}
					score := state.score + serviceAllocationScore(service, qty, targets[idx], weights[service.Code])
					count := state.count + qty
					existing, ok := next[nextAmount]
					if !ok || score > existing.score || (score == existing.score && count < existing.count) {
						stage := groupStage{score: score, count: count, prev: amount, qty: qty, exists: true}
						next[nextAmount] = stage
						parents[idx][nextAmount] = stage
					}
				}
			}
			current = next
		}

		finalState, ok := current[target]
		if !ok || !finalState.exists {
			continue
		}
		quantities := make([]int, len(services))
		currentAmount := target
		for idx := len(services) - 1; idx >= 0; idx-- {
			stage := parents[idx][currentAmount]
			quantities[idx] = stage.qty
			currentAmount = stage.prev
		}
		return groupAllocation{amount: target, score: finalState.score, quantities: quantities, ok: true}, true
	}
	return groupAllocation{}, false
}

// RU: Р¤СѓРЅРєС†РёСЏ `serviceAllocationScore`.
// EN: Function `serviceAllocationScore`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: serviceAllocationScore ranks candidate quantities by closeness to target share and weight preference.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func serviceAllocationScore(service Service, qty int, targetShare int, weight int) int {
	allocated := qty * service.Rate
	score := -absInt(allocated-targetShare) * 100
	if targetShare > 0 && qty == 0 {
		score -= 50_000
	}
	score += weight * qty * 25
	if qty > 0 {
		score += 500
	}
	return score
}

// RU: Р¤СѓРЅРєС†РёСЏ `absInt`.
// EN: Function `absInt`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РІСЃРїРѕРјРѕРіР°С‚РµР»СЊРЅРѕРµ РїСЂРµРѕР±СЂР°Р·РѕРІР°РЅРёРµ, РїСЂРѕРІРµСЂРєСѓ РёР»Рё РїРѕРґРіРѕС‚РѕРІРєСѓ РґР°РЅРЅС‹С….
// EN: What it does: absInt is a tiny helper used by scoring and distance calculations.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РІР°Р¶РµРЅ РґР»СЏ СѓСЃС‚РѕР№С‡РёРІРѕСЃС‚Рё Р»РѕРіРёРєРё; РјРѕР¶РµС‚ РёСЃРїРѕР»СЊР·РѕРІР°С‚СЊСЃСЏ СЃСЂР°Р·Сѓ РІ РЅРµСЃРєРѕР»СЊРєРёС… РјРµСЃС‚Р°С…; РёР·РјРµРЅРµРЅРёСЏ СЃС‚РѕРёС‚ РґРµР»Р°С‚СЊ РѕСЃРѕР·РЅР°РЅРЅРѕ.
// EN: Key points: supports consistency and readability of the project; may be reused by several code paths; changes should be made deliberately.
func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// RU: Р¤СѓРЅРєС†РёСЏ `solveExact`.
// EN: Function `solveExact`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: solveExact is the fallback exact solver used when the structured strategy cannot satisfy the target.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func solveExact(target int, services []Service, weights map[string]int) ([]int, bool) {
	baseQuantities, reducedTarget, _ := reserveMinimumServices(target, services)
	dp := make([]allocationState, reducedTarget+1)
	dp[0] = allocationState{ok: true, prev: -1, idx: -1}
	for amount := 0; amount <= reducedTarget; amount++ {
		if !dp[amount].ok {
			continue
		}
		for idx, service := range services {
			next := amount + service.Rate
			if next > reducedTarget {
				continue
			}
			candidate := allocationState{score: dp[amount].score + serviceAllocationScore(service, 1, service.Rate, weights[service.Code]), count: dp[amount].count + 1, prev: amount, idx: idx, ok: true}
			if !dp[next].ok || candidate.score > dp[next].score || (candidate.score == dp[next].score && candidate.count < dp[next].count) {
				dp[next] = candidate
			}
		}
	}
	if !dp[reducedTarget].ok {
		return nil, false
	}
	quantities := append([]int(nil), baseQuantities...)
	for current := reducedTarget; current > 0; {
		step := dp[current]
		quantities[step.idx]++
		current = step.prev
	}
	return quantities, true
}

// RU: Р¤СѓРЅРєС†РёСЏ `mergeGroupQuantities`.
// EN: Function `mergeGroupQuantities`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: СѓС‡Р°СЃС‚РІСѓРµС‚ РІРѕ РІРЅСѓС‚СЂРµРЅРЅРµР№ РјР°С‚РµРјР°С‚РёРєРµ СЂР°СЃС‡С‘С‚Р° Рё С‚РѕС‡РЅРѕРј СЂР°СЃРїСЂРµРґРµР»РµРЅРёРё СЃСѓРјРјС‹.
// EN: What it does: mergeGroupQuantities writes group-local quantities back into the full result slice.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: СЏРІР»СЏРµС‚СЃСЏ С‡Р°СЃС‚СЊСЋ СЂР°СЃС‡С‘С‚РЅРѕРіРѕ РїР°Р№РїР»Р°Р№РЅР°; С‡СѓРІСЃС‚РІРёС‚РµР»РµРЅ Рє РіСЂР°РЅРёС‡РЅС‹Рј СЃР»СѓС‡Р°СЏРј; С‚СЂРµР±СѓРµС‚ С‚РµСЃС‚РѕРІРѕР№ РїСЂРѕРІРµСЂРєРё РїРѕСЃР»Рµ РїСЂР°РІРѕРє.
// EN: Key points: belongs to the calculation pipeline; is sensitive to edge cases and exact arithmetic; should be changed together with tests.
func mergeGroupQuantities(result []int, quantities []int, indexMap []int) {
	for idx, quantity := range quantities {
		result[indexMap[idx]] += quantity
	}
}

// RU: РњРµС‚РѕРґ `SaveCalculation`.
// EN: Method `SaveCalculation`.
//
// RU: Р§С‚Рѕ РґРµР»Р°РµС‚: РІС‹РїРѕР»РЅСЏРµС‚ РёР·РјРµРЅРµРЅРёРµ РґР°РЅРЅС‹С… РІ РїСЂРёР»РѕР¶РµРЅРёРё Рё РїСЂРѕРІРѕРґРёС‚ Р±РёР·РЅРµСЃ-РѕРїРµСЂР°С†РёСЋ.
// EN: What it does: SaveCalculation stores the current calculation in SQLite together with its author and line items.
//
// RU: РљР»СЋС‡РµРІС‹Рµ РјРѕРјРµРЅС‚С‹: РјРµРЅСЏРµС‚ СЃРѕСЃС‚РѕСЏРЅРёРµ SQLite РёР»Рё СЃРµСЃСЃРёРё; РѕРїРёСЂР°РµС‚СЃСЏ РЅР° РїСЂРѕРІРµСЂРєРё СЂРѕР»РµР№ Рё РІР»Р°РґРµРЅРёСЏ; РѕС€РёР±РєРё Р·РґРµСЃСЊ Р·Р°РјРµС‚РЅС‹ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ СЃСЂР°Р·Сѓ.
// EN: Key points: mutates SQLite and/or session state; depends on role and ownership checks; failures here are visible to the user immediately.
