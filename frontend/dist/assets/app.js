const appState = {
  session: { authenticated: false, canManage: false, canAdmin: false, canModerate: false, user: null },
  services: [],
  users: [],
  savedCalculations: [],
  currentCalculation: null,
  activeHistoryId: null,
  archiveOwnerFilter: "all",
  activeTab: "calculator",
  weights: {},
  editingServiceId: 0,
  defaultGroupPercent: { primary: 0.79, secondary: 0.2, closing: 0.01 },
};

const api = window.go?.main?.App;

const loginScreen = document.getElementById("login-screen");
const appShell = document.getElementById("app-shell");
const loginUsername = document.getElementById("login-username");
const loginPassword = document.getElementById("login-password");
const loginButton = document.getElementById("login-button");
const loginMessage = document.getElementById("login-message");
const sessionUsername = document.getElementById("session-username");
const sessionRole = document.getElementById("session-role");
const logoutButton = document.getElementById("logout-button");
const settingsTabButton = document.getElementById("settings-tab-button");
const amountInput = document.getElementById("amount-input");
const titleInput = document.getElementById("title-input");
const calculateButton = document.getElementById("calculate-button");
const saveButton = document.getElementById("save-button");
const resetWeightsButton = document.getElementById("reset-weights-button");
const deleteArchiveButton = document.getElementById("delete-archive-button");
const messageNode = document.getElementById("message");
const resultBody = document.getElementById("result-body");
const resultTotal = document.getElementById("result-total");
const summaryTotal = document.getElementById("summary-total");
const summaryItems = document.getElementById("summary-items");
const summaryHistory = document.getElementById("summary-history");
const exactStatus = document.getElementById("exact-status");
const defaultPercentages = document.getElementById("default-percentages");
const servicesList = document.getElementById("services-list");
const historyList = document.getElementById("history-list");
const archiveOwnerField = document.getElementById("archive-owner-field");
const archiveOwnerFilter = document.getElementById("archive-owner-filter");
const archiveBody = document.getElementById("archive-body");
const archiveTitle = document.getElementById("archive-title");
const archiveTotal = document.getElementById("archive-total");
const archiveMeta = document.getElementById("archive-meta");
const servicesAdminList = document.getElementById("services-admin-list");
const serviceNameInput = document.getElementById("service-name");
const serviceUnitInput = document.getElementById("service-unit");
const serviceRateInput = document.getElementById("service-rate");
const serviceCategoryInput = document.getElementById("service-category");
const servicePercentInput = document.getElementById("service-percent");
const servicePercentHint = document.getElementById("service-percent-hint");
const saveServiceButton = document.getElementById("save-service-button");
const resetServiceFormButton = document.getElementById("reset-service-form-button");
const serviceFormMessage = document.getElementById("service-form-message");
const usersPanel = document.getElementById("users-panel");
const usersList = document.getElementById("users-list");
const userUsernameInput = document.getElementById("user-username");
const userPasswordInput = document.getElementById("user-password");
const userRoleInput = document.getElementById("user-role");
const createUserButton = document.getElementById("create-user-button");
const userFormMessage = document.getElementById("user-form-message");
const tabButtons = document.querySelectorAll("[data-tab]");
const tabPanels = {
  calculator: document.getElementById("calculator-tab"),
  archive: document.getElementById("archive-tab"),
  settings: document.getElementById("settings-tab"),
};
const confirmModal = document.getElementById("confirm-modal");
const confirmBackdrop = document.getElementById("confirm-backdrop");
const confirmTitle = document.getElementById("confirm-title");
const confirmText = document.getElementById("confirm-text");
const confirmCancel = document.getElementById("confirm-cancel");
const confirmSubmit = document.getElementById("confirm-submit");

let confirmResolver = null;

function formatMoney(value) {
  return `${new Intl.NumberFormat("ru-RU").format(value)} р`;
}

function closeConfirmModal(confirmed = false) {
  confirmModal.classList.add("hidden");
  confirmModal.setAttribute("aria-hidden", "true");
  const resolver = confirmResolver;
  confirmResolver = null;
  if (resolver) {
    resolver(confirmed);
  }
}

function openConfirmModal({ title, text, confirmLabel = "Удалить" }) {
  confirmTitle.textContent = title;
  confirmText.textContent = text;
  confirmSubmit.textContent = confirmLabel;
  confirmModal.classList.remove("hidden");
  confirmModal.setAttribute("aria-hidden", "false");
  return new Promise((resolve) => {
    confirmResolver = resolve;
  });
}

async function requestDeleteConfirmation(entityLabel, entityName) {
  return openConfirmModal({
    title: `Удалить ${entityLabel}?`,
    text: `Подтвердите удаление: ${entityName}. Это действие нельзя отменить.`,
    confirmLabel: "Удалить",
  });
}

function formatDate(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function stripSpaces(value) {
  return String(value ?? "").replace(/\s+/g, "");
}

function categoryLabel(category) {
  switch (category) {
    case "primary":
      return "Основная";
    case "secondary":
      return "Вторичная";
    case "closing":
      return "Закрывающая";
    default:
      return "Без группы";
  }
}

function categoryClass(category) {
  switch (category) {
    case "primary":
      return "category-primary";
    case "secondary":
      return "category-secondary";
    default:
      return "category-closing";
  }
}

function roleLabel(role) {
  switch (role) {
    case "admin":
      return "Администратор";
    case "manager":
      return "Руководитель";
    case "senior_specialist":
      return "Старший специалист ТП";
    default:
      return "Специалист Тех. Поддержки";
  }
}

function normalizeErrorText(text, fallback = "Произошла ошибка.") {
  const value = String(text ?? "").replace(/^Error:\s*/, "").trim();
  if (!value) {
    return fallback;
  }
  if (value.includes("Р ") || value.includes("РЎ") || value.includes("СЃ") || value.includes("РґРѕСЃС‚")) {
    return fallback;
  }
  return value;
}

function setMessage(text, type = "muted") {
  messageNode.textContent = normalizeErrorText(text, "Произошла ошибка.");
  messageNode.className = `message ${type}`;
}

function setLoginMessage(text, type = "muted") {
  loginMessage.textContent = normalizeErrorText(text, "Ошибка входа. Проверьте логин и пароль.");
  loginMessage.className = `message ${type}`;
}

function setServiceFormMessage(text, type = "muted") {
  serviceFormMessage.textContent = normalizeErrorText(text, "Не удалось сохранить услугу.");
  serviceFormMessage.className = `message ${type}`;
}

function setUserFormMessage(text, type = "muted") {
  userFormMessage.textContent = normalizeErrorText(text, "Не удалось выполнить действие с пользователем.");
  userFormMessage.className = `message ${type}`;
}

function setActiveTab(tab) {
  appState.activeTab = tab;
  tabButtons.forEach((button) => {
    button.classList.toggle("active", button.dataset.tab === tab);
  });
  Object.entries(tabPanels).forEach(([name, panel]) => {
    panel.classList.toggle("active", name === tab);
  });
}

function applyBootstrap(data) {
  appState.session = data.session ?? { authenticated: false, canManage: false, canAdmin: false, canModerate: false, user: null };
  appState.services = Array.isArray(data.services) ? data.services : [];
  appState.users = Array.isArray(data.users) ? data.users : [];
  appState.savedCalculations = Array.isArray(data.savedCalculations) ? data.savedCalculations : [];
  appState.defaultGroupPercent = data.defaultGroupPercent ?? appState.defaultGroupPercent;

  const nextWeights = {};
  appState.services.forEach((service) => {
    nextWeights[service.code] = appState.weights[service.code] ?? 0;
  });
  appState.weights = nextWeights;

  const authors = [...new Set(appState.savedCalculations.map((item) => item.createdBy).filter(Boolean))];
  if (!authors.includes(appState.archiveOwnerFilter)) {
    appState.archiveOwnerFilter = appState.session?.canModerate ? (authors[0] ?? "all") : "all";
  }

  if (!appState.savedCalculations.some((item) => item.id === appState.activeHistoryId)) {
    appState.activeHistoryId = appState.savedCalculations[0]?.id ?? null;
  }
}

function renderShellState() {
  const authenticated = Boolean(appState.session?.authenticated);
  loginScreen.classList.toggle("hidden", authenticated);
  appShell.classList.toggle("hidden", !authenticated);

  if (!authenticated) {
    return;
  }

  sessionUsername.textContent = appState.session.user?.username ?? "Гость";
  sessionRole.textContent = roleLabel(appState.session.user?.role);
  settingsTabButton.classList.remove("hidden");
  usersPanel.classList.toggle("hidden", !appState.session.canModerate);
}

function applyWeightAdjustmentsToPercents(group, result, weightMap) {
  if (!weightMap || group.length < 2) {
    return;
  }
  group.forEach((item) => {
    const code = item.serviceCode ?? item.code;
    let weight = Number(weightMap?.[code] ?? 0);
    if (!Number.isFinite(weight)) {
      weight = 0;
    }
    weight = Math.max(0, Math.min(10, Math.trunc(weight)));
    for (let step = 0; step < weight; step += 1) {
      const donors = group
        .map((entry) => entry.serviceCode ?? entry.code)
        .filter((donorCode) => donorCode !== code && (result[donorCode] ?? 0) > 0);
      let remaining = 2;
      let activeDonors = donors;
      while (remaining > 0.0001 && activeDonors.length) {
        const slice = remaining / activeDonors.length;
        const nextDonors = [];
        let moved = 0;
        activeDonors.forEach((donorCode) => {
          const donorValue = result[donorCode] ?? 0;
          const take = Math.min(slice, donorValue);
          if (take <= 0) {
            return;
          }
          result[donorCode] = donorValue - take;
          moved += take;
          if (result[donorCode] > 0.0001) {
            nextDonors.push(donorCode);
          }
        });
        if (moved <= 0) {
          break;
        }
        result[code] = (result[code] ?? 0) + moved;
        remaining -= moved;
        activeDonors = nextDonors;
      }
    }
  });
}

function buildEffectivePercentMap(items, weightMap = null) {
  const grouped = { primary: [], secondary: [], closing: [] };
  items.forEach((item) => {
    const category = item.category ?? "primary";
    if (!grouped[category]) {
      grouped[category] = [];
    }
    grouped[category].push(item);
  });

  const result = {};
  Object.entries(grouped).forEach(([category, group]) => {
    if (!group.length) {
      return;
    }

    const categoryPercent = (appState.defaultGroupPercent?.[category] ?? 0) * 100;
    let explicitTotal = 0;
    let unassigned = 0;
    group.forEach((item) => {
      if (typeof item.allocationPercent === "number") {
        explicitTotal += item.allocationPercent;
      } else {
        unassigned += 1;
      }
    });

    const fallback = unassigned > 0 ? Math.max(0, categoryPercent - explicitTotal) / unassigned : 0;
    group.forEach((item) => {
      const key = item.serviceCode ?? item.code;
      result[key] = typeof item.allocationPercent === "number" ? item.allocationPercent : fallback;
    });
    applyWeightAdjustmentsToPercents(group, result, weightMap);
  });

  return result;
}

function getEditableServicesSnapshot() {
  const services = appState.services.map((service) => ({ ...service }));
  const targetIndex = services.findIndex((service) => service.id === appState.editingServiceId);
  if (targetIndex >= 0) {
    const nextPercent = servicePercentInput.value.trim();
    services[targetIndex] = {
      ...services[targetIndex],
      category: serviceCategoryInput.value,
      allocationPercent: nextPercent ? Number(nextPercent.replace(",", ".")) : null,
    };
  }
  return services;
}

function updateServicePercentHint() {
  const previewServices = getEditableServicesSnapshot();
  const percentMap = buildEffectivePercentMap(previewServices);
  const nextPercentRaw = servicePercentInput.value.trim();
  const nextPercent = nextPercentRaw ? Number(nextPercentRaw.replace(",", ".")) : null;
  const editingService = previewServices.find((service) => service.id === appState.editingServiceId);
  const fallbackGroupServices = previewServices.filter((service) => service.category === serviceCategoryInput.value);

  let currentValue = 0;
  if (editingService) {
    currentValue = percentMap[editingService.code] ?? 0;
  } else if (nextPercent !== null && Number.isFinite(nextPercent)) {
    currentValue = nextPercent;
  } else if (fallbackGroupServices.length) {
    const categoryPercent = (appState.defaultGroupPercent?.[serviceCategoryInput.value] ?? 0) * 100;
    currentValue = categoryPercent / fallbackGroupServices.length;
  }

  const placeholderValue = formatPercent(currentValue || 0);
  servicePercentInput.placeholder = `Пусто = ${placeholderValue} по умолчанию`;
  servicePercentHint.textContent = nextPercent !== null && Number.isFinite(nextPercent)
    ? `Сейчас для услуги будет установлен процент ${formatPercent(nextPercent)}.`
    : `Сейчас для услуги действует процент по группе: ${placeholderValue}. Он рассчитан с учётом количества услуг в этой группе.`;
}

function formatPercent(value) {
  return `${value.toFixed(value % 1 === 0 ? 0 : 1)}%`;
}

function percentText(service, percentMap) {
  return formatPercent(percentMap[service.code] ?? 0);
}

function renderDefaultPercentages() {
  const groups = appState.defaultGroupPercent;
  defaultPercentages.textContent = `По умолчанию: основные ${Math.round((groups.primary ?? 0) * 100)}%, вторичные ${Math.round((groups.secondary ?? 0) * 100)}%, закрывающие ${Math.round((groups.closing ?? 0) * 100)}%. Внутри группы этот процент делится равномерно, если для услуги не задан свой процент.`;
}

function renderServices() {
  if (!appState.services.length) {
    servicesList.innerHTML = '<div class="service-card">Услуги ещё не созданы. Добавьте их во вкладке настроек.</div>';
    return;
  }

  const effectivePercentMap = buildEffectivePercentMap(appState.services, appState.weights);
  servicesList.innerHTML = appState.services.map((service, index) => `
    <article class="service-card">
      <div class="service-card-top">
        <div>
          <strong class="truncate-text" title="${escapeHtml(service.name)}">${index + 1}. ${escapeHtml(service.name)}</strong>
          <div class="service-meta">1 ${escapeHtml(service.unit)} = ${formatMoney(service.rate)}</div>
          <div class="service-meta">Процент: ${escapeHtml(percentText(service, effectivePercentMap))}</div>
          <span class="category-badge ${categoryClass(service.category)}">${categoryLabel(service.category)}</span>
        </div>
        <div class="weight-box">
          <label for="weight-${service.code}">Вес</label>
          <input id="weight-${service.code}" class="weight-input" type="number" min="0" max="10" step="1" value="${appState.weights[service.code] ?? 0}" data-weight-code="${escapeHtml(service.code)}" />
        </div>
      </div>
    </article>
  `).join("");

  servicesList.querySelectorAll("[data-weight-code]").forEach((input) => {
    input.addEventListener("input", () => {
      const nextValue = Number(input.value);
      appState.weights[input.dataset.weightCode] = Number.isFinite(nextValue) ? Math.max(0, Math.min(10, Math.trunc(nextValue))) : 0;
    });
  });
}

function renderResult(calculation) {
  if (!calculation || !Array.isArray(calculation.items) || !calculation.items.length) {
    resultBody.innerHTML = '<tr><td colspan="7" class="placeholder">Результаты появятся здесь после расчёта.</td></tr>';
    resultTotal.textContent = "0 р";
    summaryTotal.textContent = "0 р";
    summaryItems.textContent = `0/${appState.services.length}`;
    exactStatus.textContent = "Ожидание расчёта";
    saveButton.disabled = true;
    return;
  }

  const effectivePercentMap = buildEffectivePercentMap(appState.services, appState.weights);
  resultBody.innerHTML = calculation.items.map((item, index) => `
    <tr class="${item.quantity === 0 ? "muted-row" : ""}">
      <td>${index + 1}</td>
      <td>
        <strong class="truncate-text" title="${escapeHtml(item.name)}">${escapeHtml(item.name)}</strong>
        <div class="service-meta">${formatMoney(item.rate)} / ${escapeHtml(item.unit)}</div>
      </td>
      <td class="group-cell">${categoryLabel(item.category)}</td>
      <td class="weight-value">${formatPercent(effectivePercentMap[item.serviceCode] ?? 0)}</td>
      <td class="weight-value">${item.weight ?? 0}</td>
      <td class="qty">${item.quantity} ${escapeHtml(item.unit)}</td>
      <td class="money">${formatMoney(item.lineTotal)}</td>
    </tr>
  `).join("");

  resultTotal.textContent = formatMoney(calculation.totalAmount);
  summaryTotal.textContent = formatMoney(calculation.totalAmount);
  summaryItems.textContent = `${calculation.activeServices}/${calculation.items.length}`;
  exactStatus.textContent = calculation.foundExact ? "Точное совпадение найдено" : "Точное совпадение не найдено";
  saveButton.disabled = !calculation.foundExact;
}

function renderArchiveDetails(saved) {
  deleteArchiveButton.disabled = !saved;
  if (!saved) {
    archiveTitle.textContent = appState.session?.canModerate ? "Выберите расчёт сотрудника" : "Выберите расчёт из архива";
    archiveTotal.textContent = "0 р";
    archiveMeta.textContent = "Здесь появятся дата, автор и состав выбранного расчёта.";
    archiveBody.innerHTML = '<tr><td colspan="7" class="placeholder">Архив пока пуст или расчёт ещё не выбран.</td></tr>';
    return;
  }

  const weightMap = Object.fromEntries(saved.items.map((item) => [item.serviceCode, item.weight ?? 0]));
  const effectivePercentMap = buildEffectivePercentMap(saved.items, weightMap);
  archiveTitle.textContent = saved.title;
  archiveTotal.textContent = formatMoney(saved.totalAmount);
  archiveMeta.innerHTML = `<span>Целевая сумма: <strong>${formatMoney(saved.targetAmount)}</strong></span><span>Автор: <strong>${escapeHtml(saved.createdBy || "не указан")}</strong></span><span class="history-date">${formatDate(saved.createdAt)}</span>`;
  archiveBody.innerHTML = saved.items.map((item, index) => `
    <tr class="${item.quantity === 0 ? "muted-row" : ""}">
      <td>${index + 1}</td>
      <td>
        <strong class="truncate-text" title="${escapeHtml(item.name)}">${escapeHtml(item.name)}</strong>
        <div class="service-meta">${formatMoney(item.rate)} / ${escapeHtml(item.unit)}</div>
      </td>
      <td class="group-cell">${categoryLabel(item.category)}</td>
      <td class="weight-value">${formatPercent(effectivePercentMap[item.serviceCode] ?? 0)}</td>
      <td class="weight-value">${item.weight ?? 0}</td>
      <td class="qty">${item.quantity} ${escapeHtml(item.unit)}</td>
      <td class="money">${formatMoney(item.lineTotal)}</td>
    </tr>
  `).join("");
}

function getFilteredCalculations() {
  if (!appState.session?.canModerate || appState.archiveOwnerFilter === "all") {
    return appState.savedCalculations;
  }
  return appState.savedCalculations.filter((item) => item.createdBy === appState.archiveOwnerFilter);
}

function renderArchiveOwnerFilter() {
  const canSelectOwner = Boolean(appState.session?.canModerate);
  archiveOwnerField.classList.toggle("hidden", !canSelectOwner);
  if (!canSelectOwner) {
    return;
  }

  const authors = [...new Set(appState.savedCalculations.map((item) => item.createdBy).filter(Boolean))];
  const options = [];
  if (appState.session?.canAdmin) {
    options.push('<option value="all">Все доступные архивы</option>');
  }
  authors.forEach((author) => {
    options.push(`<option value="${escapeHtml(author)}" ${appState.archiveOwnerFilter === author ? "selected" : ""}>${escapeHtml(author)}</option>`);
  });
  archiveOwnerFilter.innerHTML = options.join("");
  if (!appState.session?.canAdmin && authors.length && !authors.includes(appState.archiveOwnerFilter)) {
    appState.archiveOwnerFilter = authors[0];
  }
}

function renderHistory() {
  renderArchiveOwnerFilter();
  summaryHistory.textContent = String(appState.savedCalculations.length);
  const visibleCalculations = getFilteredCalculations();
  if (!visibleCalculations.length) {
    historyList.innerHTML = '<div class="history-empty">Пока ничего не сохранено.</div>';
    appState.activeHistoryId = null;
    renderArchiveDetails(null);
    return;
  }

  if (!visibleCalculations.some((item) => item.id === appState.activeHistoryId)) {
    appState.activeHistoryId = visibleCalculations[0]?.id ?? null;
  }

  historyList.innerHTML = visibleCalculations.map((item) => `
    <article class="history-card ${appState.activeHistoryId === item.id ? "active" : ""}">
      <div class="history-card-top">
        <button type="button" class="history-open" data-history-id="${item.id}">
          <strong class="truncate-text" title="${escapeHtml(item.title)}">${escapeHtml(item.title)}</strong>
          <div class="history-meta">
            <span>Цель: ${formatMoney(item.targetAmount)}</span>
            <span>Итог: ${formatMoney(item.totalAmount)}</span>
            <span>Автор: ${escapeHtml(item.createdBy || "-")}</span>
            <span class="history-date">${formatDate(item.createdAt)}</span>
          </div>
        </button>
        <button type="button" class="history-delete" data-delete-id="${item.id}">Удалить</button>
      </div>
    </article>
  `).join("");

  historyList.querySelectorAll("[data-history-id]").forEach((node) => {
    node.addEventListener("click", () => {
      const id = Number(node.dataset.historyId);
      appState.activeHistoryId = id;
      const current = visibleCalculations.find((item) => item.id === id) ?? null;
      renderHistory();
      renderArchiveDetails(current);
      setActiveTab("archive");
      setMessage("Архивный расчёт открыт.", "success");
    });
  });

  historyList.querySelectorAll("[data-delete-id]").forEach((node) => {
    node.addEventListener("click", async () => {
      await deleteCalculation(Number(node.dataset.deleteId));
    });
  });

  const current = visibleCalculations.find((item) => item.id === appState.activeHistoryId) ?? visibleCalculations[0];
  if (current) {
    appState.activeHistoryId = current.id;
    renderArchiveDetails(current);
  }
}

function resetServiceForm() {
  appState.editingServiceId = 0;
  serviceNameInput.value = "";
  serviceUnitInput.value = "ч.";
  serviceRateInput.value = "";
  serviceCategoryInput.value = "primary";
  servicePercentInput.value = "";
  saveServiceButton.textContent = "Сохранить услугу";
  updateServicePercentHint();
}

function startEditService(id) {
  const service = appState.services.find((item) => item.id === id);
  if (!service || !appState.session.canManage) {
    return;
  }
  appState.editingServiceId = id;
  serviceNameInput.value = service.name;
  serviceUnitInput.value = service.unit;
  serviceRateInput.value = String(service.rate);
  serviceCategoryInput.value = service.category;
  servicePercentInput.value = typeof service.allocationPercent === "number" ? String(service.allocationPercent) : "";
  saveServiceButton.textContent = "Обновить услугу";
  updateServicePercentHint();
  setActiveTab("settings");
  setServiceFormMessage(`Редактируется услуга «${service.name}».`, "muted");
}

function renderServicesAdmin() {
  if (!appState.services.length) {
    servicesAdminList.innerHTML = '<div class="settings-empty">Список услуг пуст.</div>';
    return;
  }

  const effectivePercentMap = buildEffectivePercentMap(appState.services, appState.weights);
  if (!appState.session.canManage) {
    servicesAdminList.innerHTML = appState.services.map((service) => `
      <article class="settings-item glass">
        <div>
          <strong class="truncate-text" title="${escapeHtml(service.name)}">${escapeHtml(service.name)}</strong>
          <div class="service-meta">${formatMoney(service.rate)} / ${escapeHtml(service.unit)}</div>
          <div class="service-meta">${categoryLabel(service.category)}, текущий процент ${formatPercent(effectivePercentMap[service.code] ?? 0)}</div>
        </div>
      </article>
    `).join("");
    saveServiceButton.disabled = true;
    resetServiceFormButton.disabled = true;
    [serviceNameInput, serviceUnitInput, serviceRateInput, serviceCategoryInput, servicePercentInput].forEach((node) => {
      node.disabled = true;
    });
    setServiceFormMessage("Здесь видно, какой процент сейчас у каждой вашей услуги. Каждый пользователь управляет только своим списком услуг.", "muted");
    updateServicePercentHint();
    return;
  }

  saveServiceButton.disabled = false;
  resetServiceFormButton.disabled = false;
  [serviceNameInput, serviceUnitInput, serviceRateInput, serviceCategoryInput, servicePercentInput].forEach((node) => {
    node.disabled = false;
  });

  servicesAdminList.innerHTML = appState.services.map((service) => `
    <article class="settings-item glass">
      <div>
        <strong class="truncate-text" title="${escapeHtml(service.name)}">${escapeHtml(service.name)}</strong>
        <div class="service-meta">${formatMoney(service.rate)} / ${escapeHtml(service.unit)}</div>
        <div class="service-meta">${categoryLabel(service.category)}, текущий процент ${formatPercent(effectivePercentMap[service.code] ?? 0)}</div>
      </div>
      <div class="settings-item-actions">
        <button type="button" class="ghost-button" data-edit-service="${service.id}">Изменить</button>
        <button type="button" class="danger-button" data-delete-service="${service.id}">Удалить</button>
      </div>
    </article>
  `).join("");

  servicesAdminList.querySelectorAll("[data-edit-service]").forEach((button) => {
    button.addEventListener("click", () => startEditService(Number(button.dataset.editService)));
  });

  servicesAdminList.querySelectorAll("[data-delete-service]").forEach((button) => {
    button.addEventListener("click", async () => {
      await deleteService(Number(button.dataset.deleteService));
    });
  });

  updateServicePercentHint();
}

function renderUsers() {
  const sortedUsers = [...appState.users].sort((left, right) => {
    const leftPriority = userSortPriority(left.role);
    const rightPriority = userSortPriority(right.role);
    if (leftPriority != rightPriority) {
      return leftPriority - rightPriority;
    }
    if ((left.createdAt || "") !== (right.createdAt || "")) {
      return String(left.createdAt || "").localeCompare(String(right.createdAt || ""));
    }
    return String(left.username || "").localeCompare(String(right.username || ""), "ru");
  });
  const roleChoices = appState.session.canAdmin
    ? [["employee", "Специалист Тех. Поддержки"], ["senior_specialist", "Старший специалист ТП"], ["manager", "Руководитель"]]
    : appState.session.user?.role === "manager"
      ? [["employee", "Специалист Тех. Поддержки"], ["senior_specialist", "Старший специалист ТП"]]
      : [["employee", "Специалист Тех. Поддержки"]];

  userRoleInput.innerHTML = roleChoices
    .map(([value, label]) => `<option value="${value}">${label}</option>`)
    .join("");

  if (!roleChoices.some(([value]) => value === userRoleInput.value)) {
    userRoleInput.value = roleChoices[0]?.[0] ?? "employee";
  }

  createUserButton.disabled = !appState.session.canModerate;
  userUsernameInput.disabled = !appState.session.canModerate;
  userPasswordInput.disabled = !appState.session.canModerate;
  userRoleInput.disabled = !appState.session.canModerate;

  if (!appState.session.canModerate) {
    usersList.innerHTML = '<div class="settings-empty">Управление пользователями доступно старшему специалисту, руководителю и администратору.</div>';
    setUserFormMessage("Создавать пользователей могут старший специалист ТП, руководитель и администратор.", "muted");
    return;
  }
  if (!sortedUsers.length) {
    usersList.innerHTML = '<div class="settings-empty">Пользователи пока не созданы.</div>';
    setUserFormMessage("Можно создать первого пользователя для вашей роли.", "muted");
    return;
  }

  usersList.innerHTML = sortedUsers.map((user) => {
    const canChangeRole = appState.session.canManage;
    const canDelete = appState.session.canAdmin ? true : (user.role === "employee" || (appState.session.user?.role === "manager" && user.role === "senior_specialist"));
    const roleOptions = roleChoices.map(([value, label]) => `<option value="${value}" ${user.role === value ? "selected" : ""}>${label}</option>`).join("");

    return `
      <article class="settings-item glass user-item">
        <div>
          <strong>${escapeHtml(user.username)}</strong>
          <div class="service-meta">${roleLabel(user.role)}</div>
          <div class="service-meta">Создан: ${formatDate(user.createdAt)}</div>
        </div>
        <div class="settings-item-actions stacked-actions">
          ${canChangeRole ? `<select class="input-select compact-select" data-user-role-id="${user.id}">${roleOptions}</select><button type="button" class="ghost-button" data-apply-role-id="${user.id}">Сменить роль</button>` : `<div class="service-meta">Смена ролей доступна руководителю и администратору.</div>`}
          <button type="button" class="danger-button" ${canDelete ? `data-delete-user-id="${user.id}"` : "disabled"}>Удалить</button>
        </div>
      </article>
    `;
  }).join("");

  usersList.querySelectorAll("[data-apply-role-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const userId = Number(button.dataset.applyRoleId);
      const roleInput = usersList.querySelector(`[data-user-role-id="${userId}"]`);
      await updateUserRole(userId, roleInput?.value ?? "employee");
    });
  });

  usersList.querySelectorAll("[data-delete-user-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      await deleteUser(Number(button.dataset.deleteUserId));
    });
  });

  setUserFormMessage("Доступные роли в создании пользователя зависят от вашего уровня доступа.", "muted");
}
function userSortPriority(role) {
  switch (role) {
    case "manager":
      return 1;
    case "senior_specialist":
      return 2;
    case "employee":
      return 3;
    default:
      return 9;
  }
}

async function refreshBootstrap(options = {}) {
  const data = await api.GetBootstrap();
  applyBootstrap(data);
  renderShellState();
  renderDefaultPercentages();
  renderServices();
  renderServicesAdmin();
  renderUsers();
  renderHistory();
  renderResult(appState.currentCalculation);
  if (options.keepArchiveSelection !== true && !appState.activeHistoryId) {
    renderArchiveDetails(null);
  }
}

function buildWeightPayload() {
  const payload = {};
  appState.services.forEach((service) => {
    payload[service.code] = appState.weights[service.code] ?? 0;
  });
  return payload;
}

async function calculate() {
  const targetAmount = Number(amountInput.value);
  if (!Number.isInteger(targetAmount) || targetAmount <= 0) {
    setMessage("Введите корректную сумму больше 0.", "error");
    return;
  }

  calculateButton.disabled = true;
  setMessage("Собираю расчёт по текущему списку услуг и настройкам процентов...", "muted");

  try {
    const result = await api.CalculateAmount({
      targetAmount,
      weights: buildWeightPayload(),
    });
    appState.currentCalculation = result;
    renderResult(result);
    setMessage(`Расчёт готов. Активно ${result.activeServices} из ${result.items.length} услуг на сумму ${formatMoney(result.totalAmount)}.`, "success");
  } catch (error) {
    appState.currentCalculation = null;
    renderResult(null);
    setMessage(error, "error");
  } finally {
    calculateButton.disabled = false;
  }
}

async function saveCalculation() {
  if (!appState.currentCalculation?.items?.length) {
    setMessage("Сначала выполните расчёт, затем его можно сохранить.", "error");
    return;
  }

  saveButton.disabled = true;
  try {
    await api.SaveCalculation({
      title: titleInput.value.trim(),
      targetAmount: appState.currentCalculation.targetAmount,
      items: appState.currentCalculation.items,
    });
    await refreshBootstrap({ keepArchiveSelection: true });
    setMessage("Расчёт сохранён в архив SQLite.", "success");
  } catch (error) {
    setMessage(error, "error");
  } finally {
    saveButton.disabled = false;
  }
}

async function deleteCalculation(id) {
  const current = appState.savedCalculations.find((item) => item.id === id);
  if (!current) {
    return;
  }
  if (!await requestDeleteConfirmation("\u0440\u0430\u0441\u0447\u0451\u0442", current.title)) {
    return;
  }
  try {
    await api.DeleteCalculation(id);
    if (appState.activeHistoryId === id) {
      appState.activeHistoryId = null;
    }
    await refreshBootstrap();
    setMessage(`\u0420\u0430\u0441\u0447\u0451\u0442 ${current.title} \u0443\u0434\u0430\u043b\u0451\u043d \u0438\u0437 \u0430\u0440\u0445\u0438\u0432\u0430.`, "success");
  } catch (error) {
    setMessage(error, "error");
  }
}

function resetWeights() {
  appState.services.forEach((service) => {
    appState.weights[service.code] = 0;
  });
  renderServices();
  setMessage("Все веса сброшены в нейтральное значение.", "muted");
}

async function saveService() {
  const rate = Number(serviceRateInput.value);
  const percentRaw = servicePercentInput.value.trim();
  let allocationPercent = null;
  if (percentRaw) {
    const percentValue = Number(percentRaw.replace(",", "."));
    if (!Number.isFinite(percentValue)) {
      setServiceFormMessage("Процент услуги должен быть числом.", "error");
      return;
    }
    allocationPercent = percentValue;
  }

  saveServiceButton.disabled = true;
  try {
    await api.UpsertService({
      id: appState.editingServiceId,
      name: serviceNameInput.value.trim(),
      unit: serviceUnitInput.value,
      rate,
      category: serviceCategoryInput.value,
      allocationPercent,
    });
    resetServiceForm();
    await refreshBootstrap();
    setServiceFormMessage("Услуга сохранена и уже участвует в расчёте.", "success");
    setMessage("Список услуг обновлён.", "success");
  } catch (error) {
    setServiceFormMessage(error, "error");
  } finally {
    saveServiceButton.disabled = false;
  }
}

async function deleteService(id) {
  const service = appState.services.find((item) => item.id === id);
  if (!service) {
    return;
  }
  if (!await requestDeleteConfirmation("\u0443\u0441\u043b\u0443\u0433\u0443", service.name)) {
    return;
  }
  try {
    await api.DeleteService(id);
    if (appState.editingServiceId === id) {
      resetServiceForm();
    }
    delete appState.weights[service.code];
    await refreshBootstrap();
    setServiceFormMessage(`\u0423\u0441\u043b\u0443\u0433\u0430 ${service.name} \u0443\u0434\u0430\u043b\u0435\u043d\u0430.`, "success");
  } catch (error) {
    setServiceFormMessage(error, "error");
  }
}

async function createUser() {
  createUserButton.disabled = true;
  try {
    await api.CreateUser({
      username: stripSpaces(userUsernameInput.value),
      password: stripSpaces(userPasswordInput.value),
      role: userRoleInput.value,
    });
    userUsernameInput.value = "";
    userPasswordInput.value = "";
    userRoleInput.value = "employee";
    await refreshBootstrap({ keepArchiveSelection: true });
    setUserFormMessage("Пользователь создан.", "success");
  } catch (error) {
    setUserFormMessage(error, "error");
  } finally {
    createUserButton.disabled = false;
  }
}

async function updateUserRole(userId, role) {
  try {
    await api.UpdateUserRole(userId, role);
    await refreshBootstrap({ keepArchiveSelection: true });
    setUserFormMessage("Роль пользователя обновлена.", "success");
  } catch (error) {
    setUserFormMessage(error, "error");
  }
}

async function deleteUser(userId) {
  const user = appState.users.find((item) => item.id === userId);
  if (!user) {
    return;
  }
  if (!await requestDeleteConfirmation("\u043f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u044f", user.username)) {
    return;
  }
  try {
    await api.DeleteUser(userId);
    await refreshBootstrap({ keepArchiveSelection: true });
    setUserFormMessage(`ÐÐ¾Ð»ÑÐ·Ð¾Ð²Ð°ÑÐµÐ»Ñ ${user.username} ÑÐ´Ð°Ð»ÑÐ½.`, "success");
  } catch (error) {
    setUserFormMessage(error, "error");
  }
}

async function login() {
  const username = stripSpaces(loginUsername.value);
  const password = stripSpaces(loginPassword.value);
  loginButton.disabled = true;
  setLoginMessage("Проверяю учётные данные...", "muted");
  try {
    await api.Login({ username, password });
    loginPassword.value = "";
    await refreshBootstrap();
    setActiveTab("calculator");
    setMessage(`Вы вошли как ${username}.`, "success");
    setLoginMessage("Вход выполнен.", "success");
  } catch (error) {
    setLoginMessage(error, "error");
  } finally {
    loginButton.disabled = false;
  }
}

async function logout() {
  try {
    await api.Logout();
    appState.currentCalculation = null;
    appState.activeHistoryId = null;
    renderResult(null);
    renderArchiveDetails(null);
    await refreshBootstrap();
    loginUsername.value = "";
    loginPassword.value = "";
    setLoginMessage("Вы вышли из системы.", "muted");
  } catch (error) {
    setMessage(error, "error");
  }
}

let autoCalculateTimer = null;
amountInput.addEventListener("input", () => {
  clearTimeout(autoCalculateTimer);
  if (!amountInput.value.trim()) {
    appState.currentCalculation = null;
    renderResult(null);
    setMessage("Введите сумму и при необходимости сместите акцент весами по услугам.", "muted");
    return;
  }
  autoCalculateTimer = setTimeout(() => {
    calculate();
  }, 350);
});

serviceCategoryInput.addEventListener("change", updateServicePercentHint);
servicePercentInput.addEventListener("input", updateServicePercentHint);
loginButton.addEventListener("click", login);
loginPassword.addEventListener("keydown", (event) => {
  if (event.key === "Enter") {
    login();
  }
});
calculateButton.addEventListener("click", calculate);
saveButton.addEventListener("click", saveCalculation);
resetWeightsButton.addEventListener("click", resetWeights);
archiveOwnerFilter?.addEventListener("change", () => {
  appState.archiveOwnerFilter = archiveOwnerFilter.value || "all";
  appState.activeHistoryId = null;
  renderHistory();
});deleteArchiveButton.addEventListener("click", async () => {
  if (appState.activeHistoryId !== null) {
    await deleteCalculation(appState.activeHistoryId);
  }
});
saveServiceButton.addEventListener("click", saveService);
resetServiceFormButton.addEventListener("click", () => {
  resetServiceForm();
  setServiceFormMessage("Форма очищена.", "muted");
});
confirmCancel.addEventListener("click", () => {
  closeConfirmModal(false);
});
confirmSubmit.addEventListener("click", () => {
  closeConfirmModal(true);
});
confirmBackdrop.addEventListener("click", () => {
  closeConfirmModal(false);
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && !confirmModal.classList.contains("hidden")) {
    closeConfirmModal(false);
    return;
  }

  const key = String(event.key || "").toLowerCase();
  const ctrlOrMeta = event.ctrlKey || event.metaKey;
  const blockedShortcuts =
    event.key === "F12" ||
    (ctrlOrMeta && event.shiftKey && ["i", "j", "c"].includes(key)) ||
    (ctrlOrMeta && ["u", "s", "p", "+", "=", "-", "0"].includes(key));

  if (blockedShortcuts) {
    event.preventDefault();
    event.stopPropagation();
  }
});

document.addEventListener("wheel", (event) => {
  if (event.ctrlKey || event.metaKey) {
    event.preventDefault();
  }
}, { passive: false });

document.addEventListener("contextmenu", (event) => {
  event.preventDefault();
});

createUserButton.addEventListener("click", createUser);
logoutButton.addEventListener("click", logout);

tabButtons.forEach((button) => {
  button.addEventListener("click", () => {
    setActiveTab(button.dataset.tab);
  });
});

window.addEventListener("DOMContentLoaded", async () => {
  if (!api?.GetBootstrap) {
    setLoginMessage("Wails API не обнаружен. Запустите проект через Wails.", "error");
    return;
  }

  loginUsername.value = "";
  loginPassword.value = "";
  resetServiceForm();
  renderResult(null);
  renderArchiveDetails(null);
  renderDefaultPercentages();

  try {
    await refreshBootstrap();
    if (appState.session.authenticated) {
      setActiveTab("calculator");
      setMessage("Сессия восстановлена.", "success");
    }
  } catch (error) {
    setLoginMessage(error, "error");
  }
});

















