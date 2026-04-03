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
const archiveDateInput = document.getElementById("archive-date-input");
const calculateButton = document.getElementById("calculate-button");
const resetWeightsButton = document.getElementById("reset-weights-button");
const deleteArchiveButton = document.getElementById("delete-archive-button");
const copyArchiveServicesButton = document.getElementById("copy-archive-services-button");
const messageNode = document.getElementById("message");
const resultBody = document.getElementById("result-body");
const resultTotal = document.getElementById("result-total");
const randomizeButton = document.getElementById("randomize-button");
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

function formatArchiveTitle(date = new Date()) {
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(date);
}

function syncArchiveDateInput() {
  if (archiveDateInput) {
    archiveDateInput.value = formatArchiveTitle(new Date());
  }
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

function normalizeRole(role) {
  switch (String(role ?? "").trim()) {
    case "manager":
    case "support_manager":
    case "support_head":
      return "support_head";
    case "senior_specialist":
    case "support_senior_specialist":
    case "support_senior":
      return "support_senior";
    case "employee":
    case "support_employee":
      return "support_employee";
    case "tech_manager":
    case "technical_head":
      return "technical_head";
    case "senior_technician":
    case "technical_senior":
      return "technical_senior";
    case "technician":
    case "technical_employee":
      return "technical_employee";
    case "mrk_manager":
    case "commercial_subscriber_head":
      return "commercial_subscriber_head";
    case "senior_mrk":
    case "commercial_senior_mrk":
      return "commercial_senior_mrk";
    case "mrk_employee":
    case "commercial_employee_mrk":
      return "commercial_employee_mrk";
    default:
      return String(role ?? "").trim() || "support_employee";
  }
}

function roleMeta(role) {
  const catalog = {
    admin: { level: 5, department: "global", label: "Администратор" },
    global_director: { level: 4, department: "global", label: "Генеральный директор" },
    executive_director: { level: 4, department: "global", label: "Исполнительный директор" },
    technical_director: { level: 4, department: "global", label: "Технический директор" },
    support_head: { level: 3, department: "support", label: "Руководитель Тех. Поддержки" },
    support_sysadmin: { level: 3, department: "support", label: "Системный администратор" },
    support_senior: { level: 2, department: "support", label: "Старший специалист техподдержки" },
    support_employee: { level: 1, department: "support", label: "Специалист техподдержки" },
    technical_head: { level: 3, department: "technical", label: "Руководитель технического отдела" },
    technical_senior: { level: 2, department: "technical", label: "Старший техник" },
    technical_employee: { level: 1, department: "technical", label: "Техник" },
    telecom_construction_director: { level: 3, department: "telecom", label: "Директор по строительству" },
    telecom_construction_head: { level: 3, department: "telecom", label: "Руководитель строительного отдела" },
    telecom_senior_vols: { level: 2, department: "telecom", label: "Старший монтажник ВОЛС" },
    telecom_senior_lvs: { level: 2, department: "telecom", label: "Старший монтажник ЛВС" },
    telecom_employee_vols: { level: 1, department: "telecom", label: "Монтажник ВОЛС" },
    telecom_employee_lvs: { level: 1, department: "telecom", label: "Монтажник ЛВС" },
    skud_head: { level: 3, department: "skud", label: "Руководитель отдела технического обслуживания СКУД" },
    skud_project_manager: { level: 2, department: "skud", label: "Менеджер проектов СКУД" },
    skud_senior_service_engineer: { level: 2, department: "skud", label: "Старший сервисный инженер СКУД" },
    skud_senior_installer: { level: 2, department: "skud", label: "Старший монтажник СКУД" },
    skud_service_engineer: { level: 1, department: "skud", label: "Сервисный инженер СКУД" },
    skud_installer: { level: 1, department: "skud", label: "Монтажник СКУД" },
    approval_head: { level: 3, department: "approval", label: "Руководитель согласования" },
    approval_senior: { level: 2, department: "approval", label: "Старший менеджер согласования" },
    approval_employee: { level: 1, department: "approval", label: "Менеджер по согласованию" },
    marketing_head: { level: 3, department: "marketing", label: "Руководитель отдела рекламы и маркетинга" },
    marketing_courier: { level: 1, department: "marketing", label: "Курьер" },
    commercial_director: { level: 4, department: "commercial", label: "Директор коммерческого блока" },
    commercial_subscriber_head: { level: 3, department: "commercial", label: "Руководитель абонентского отдела" },
    commercial_active_sales_head: { level: 3, department: "commercial", label: "Менеджер активных продаж" },
    commercial_senior_mrk: { level: 2, department: "commercial", label: "Старший МРК" },
    commercial_senior_mryu: { level: 2, department: "commercial", label: "Старший МРЮ" },
    commercial_employee_mrk: { level: 1, department: "commercial", label: "МРК" },
    commercial_employee_mryu: { level: 1, department: "commercial", label: "МРЮ" },
    finance_head: { level: 3, department: "finance", label: "Гл. бухгалтер" },
    finance_employee: { level: 1, department: "finance", label: "Помощник бухгалтера" },
    legal_employee: { level: 1, department: "legal", label: "Юрист" },
    development_head: { level: 3, department: "development", label: "Руководитель группы разработки" },
    development_senior: { level: 2, department: "development", label: "Старший разработчик" },
    development_employee: { level: 1, department: "development", label: "Разработчик" },
  };
  return catalog[normalizeRole(role)] ?? { level: 0, department: "support", label: "Пользователь" };
}

function roleDepartment(role) {
  return roleMeta(role).department;
}

function roleLevel(role) {
  return roleMeta(role).level;
}

function roleViewDepartments(role) {
  switch (normalizeRole(role)) {
    case "admin":
    case "global_director":
    case "executive_director":
      return ["support", "technical", "telecom", "skud", "approval", "marketing", "commercial", "finance", "legal", "development"];
    case "technical_director":
      return ["technical", "telecom"];
    default:
      return [roleDepartment(role)];
  }
}

function roleCanCreateUsers(role) {
  switch (normalizeRole(role)) {
    case "admin":
    case "support_head":
    case "support_sysadmin":
    case "support_senior":
    case "technical_head":
    case "technical_senior":
    case "telecom_construction_director":
    case "telecom_construction_head":
    case "telecom_senior_vols":
    case "telecom_senior_lvs":
    case "skud_head":
    case "skud_project_manager":
    case "skud_senior_service_engineer":
    case "skud_senior_installer":
    case "approval_head":
    case "approval_senior":
    case "marketing_head":
    case "commercial_director":
    case "commercial_subscriber_head":
    case "commercial_active_sales_head":
    case "commercial_senior_mrk":
    case "commercial_senior_mryu":
    case "finance_head":
    case "development_head":
    case "development_senior":
      return true;
    default:
      return false;
  }
}

function departmentLabel(department) {
  switch (department) {
    case "global":
      return "Глобальные роли";
    case "support":
      return "Техподдержка";
    case "technical":
      return "Технический отдел";
    case "telecom":
      return "Телеком / ВОЛС / ЛВС";
    case "skud":
      return "СКУД";
    case "approval":
      return "Согласование";
    case "marketing":
      return "Маркетинг";
    case "commercial":
      return "Коммерческий блок";
    case "finance":
      return "Финансы";
    case "legal":
      return "Юридический отдел";
    case "development":
      return "Разработка";
    default:
      return "Отдел";
  }
}

function departmentSortPriority(department) {
  switch (department) {
    case "global":
      return 0;
    case "support":
      return 1;
    case "technical":
      return 2;
    case "telecom":
      return 3;
    case "skud":
      return 4;
    case "approval":
      return 5;
    case "marketing":
      return 6;
    case "commercial":
      return 7;
    case "finance":
      return 8;
    case "legal":
      return 9;
    case "development":
      return 10;
    default:
      return 99;
  }
}

function roleChoiceComparator(left, right) {
  const leftMeta = roleMeta(left);
  const rightMeta = roleMeta(right);
  const leftDepartmentPriority = departmentSortPriority(leftMeta.department);
  const rightDepartmentPriority = departmentSortPriority(rightMeta.department);
  if (leftDepartmentPriority !== rightDepartmentPriority) {
    return leftDepartmentPriority - rightDepartmentPriority;
  }

  const leftLevel = roleLevel(left);
  const rightLevel = roleLevel(right);
  if (leftLevel !== rightLevel) {
    return rightLevel - leftLevel;
  }

  return roleLabel(left).localeCompare(roleLabel(right), "ru");
}

function roleChoicesForUser(role) {
  role = normalizeRole(role);
  const allRoles = [
    "global_director", "executive_director", "technical_director",
    "support_head", "support_sysadmin", "support_senior", "support_employee",
    "technical_head", "technical_senior", "technical_employee",
    "telecom_construction_director", "telecom_construction_head", "telecom_senior_vols", "telecom_senior_lvs", "telecom_employee_vols", "telecom_employee_lvs",
    "skud_head", "skud_project_manager", "skud_senior_service_engineer", "skud_senior_installer", "skud_service_engineer", "skud_installer",
    "approval_head", "approval_senior", "approval_employee",
    "marketing_head", "marketing_courier",
    "commercial_director", "commercial_subscriber_head", "commercial_active_sales_head", "commercial_senior_mrk", "commercial_senior_mryu", "commercial_employee_mrk", "commercial_employee_mryu",
    "finance_head", "finance_employee",
    "legal_employee",
    "development_head", "development_senior", "development_employee"
  ];

  let visibleRoles = [];
  if (role === "admin") {
    visibleRoles = [...allRoles];
  } else if (roleCanCreateUsers(role)) {
    const actorDept = roleDepartment(role);
    const actorLevel = roleLevel(role);
    visibleRoles = allRoles.filter((candidate) => roleDepartment(candidate) === actorDept && roleLevel(candidate) < actorLevel);
  }

  return visibleRoles
    .sort(roleChoiceComparator)
    .map((value) => ({
      value,
      label: roleLabel(value),
      level: roleLevel(value),
      department: roleDepartment(value),
    }));
}

function renderRoleOptions(roleChoices, selectedValue = "") {
  if (!roleChoices.length) {
    return "";
  }

  const groups = new Map();
  roleChoices.forEach((choice) => {
    const key = choice.department;
    if (!groups.has(key)) {
      groups.set(key, {
        department: choice.department,
        items: [],
      });
    }
    groups.get(key).items.push(choice);
  });

  return Array.from(groups.values()).map((group) => {
    const options = group.items
      .sort((left, right) => {
        if (left.level !== right.level) {
          return right.level - left.level;
        }
        return left.label.localeCompare(right.label, "ru");
      })
      .map((choice) => `<option value="${choice.value}" ${selectedValue === choice.value ? "selected" : ""}>${choice.label}</option>`)
      .join("");
    return `<optgroup label="${departmentLabel(group.department)}">${options}</optgroup>`;
  }).join("");
}

function canDeleteManagedUser(actorRole, targetRole) {
  actorRole = normalizeRole(actorRole);
  targetRole = normalizeRole(targetRole);
  if (actorRole === "admin") {
    return targetRole !== "admin";
  }
  if (!roleCanCreateUsers(actorRole)) {
    return false;
  }
  return roleDepartment(actorRole) === roleDepartment(targetRole) && roleLevel(actorRole) > roleLevel(targetRole);
}

function roleLabel(role) {
  return roleMeta(role).label;
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

  syncArchiveDateInput();
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
    const weight = clampWeightValue(Number(weightMap?.[code] ?? 0));
    if (weight > 0) {
      for (let step = 0; step < weight; step += 1) {
        transferPercentToTarget(group, result, code, 2);
      }
      return;
    }
    for (let step = 0; step < Math.abs(weight); step += 1) {
      transferPercentFromTarget(group, result, code, 2);
    }
  });
}

function transferPercentToTarget(group, result, code, amount) {
  const donors = group
    .map((entry) => entry.serviceCode ?? entry.code)
    .filter((donorCode) => donorCode !== code && (result[donorCode] ?? 0) > 0);
  let remaining = amount;
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

function transferPercentFromTarget(group, result, code, amount) {
  const receivers = group
    .map((entry) => entry.serviceCode ?? entry.code)
    .filter((otherCode) => otherCode !== code);
  const current = result[code] ?? 0;
  if (!receivers.length || current <= 0) {
    return;
  }
  const moved = Math.min(amount, current);
  if (moved <= 0) {
    return;
  }
  result[code] = current - moved;
  const slice = moved / receivers.length;
  receivers.forEach((receiverCode) => {
    result[receiverCode] = (result[receiverCode] ?? 0) + slice;
  });
}

function clampWeightValue(value) {
  if (!Number.isFinite(value)) {
    return 0;
  }
  return Math.max(-10, Math.min(10, Math.trunc(value)));
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

function resetServiceForm() {
  appState.editingServiceId = 0;
  serviceNameInput.value = "";
  serviceUnitInput.value = "Часы";
  serviceRateInput.value = "";
  serviceCategoryInput.value = "primary";
  servicePercentInput.value = "";
  saveServiceButton.textContent = "Сохранить услугу";
  updateServicePercentHint();
}

function renderResult(result) {
  if (!result) {
    resultTotal.textContent = "0 р";
    summaryTotal.textContent = "0 р";
    summaryItems.textContent = `0/${appState.services.length}`;
    exactStatus.textContent = "Ожидание расчёта";
    resultBody.innerHTML = '<tr><td colspan="7" class="placeholder">Результаты появятся здесь после расчёта.</td></tr>';
    return;
  }

  const weightMap = Object.fromEntries(result.items.map((item) => [item.serviceCode, item.weight ?? appState.weights[item.serviceCode] ?? 0]));
  const effectivePercentMap = buildEffectivePercentMap(result.items, weightMap);
  resultTotal.textContent = formatMoney(result.totalAmount);
  summaryTotal.textContent = formatMoney(result.totalAmount);
  summaryItems.textContent = `${result.activeServices}/${result.items.length}`;
  exactStatus.textContent = result.exactMatch ? "Точное совпадение найдено" : "Есть отклонение от целевой суммы";
  resultBody.innerHTML = result.items.map((item, index) => `
    <tr class="${item.quantity === 0 ? "muted-row" : ""}">
      <td>${index + 1}</td>
      <td>
        <strong class="truncate-text" title="${escapeHtml(item.name)}">${escapeHtml(item.name)}</strong>
        <div class="service-meta">${formatMoney(item.rate)} / ${escapeHtml(item.unit)}</div>
      </td>
      <td class="group-cell">${categoryLabel(item.category)}</td>
      <td class="weight-value">${formatPercent(effectivePercentMap[item.serviceCode] ?? 0)}</td>
      <td class="weight-value">${item.weight ?? appState.weights[item.serviceCode] ?? 0}</td>
      <td class="qty">${item.quantity} ${escapeHtml(item.unit)}</td>
      <td class="money">${formatMoney(item.lineTotal)}</td>
    </tr>
  `).join("");
}

function renderDefaultPercentages() {
  const groups = appState.defaultGroupPercent;
  defaultPercentages.textContent = `По умолчанию: основные ${Math.round((groups.primary ?? 0) * 100)}%, вторичные ${Math.round((groups.secondary ?? 0) * 100)}%, закрывающие ${Math.round((groups.closing ?? 0) * 100)}%. Внутри группы этот процент делится равномерно, если для услуги не задан свой процент. Вес можно задавать от -10 до 10: плюс усиливает услугу, минус ослабляет.`;
}

function renderServices() {
  if (!appState.services.length) {
    servicesList.innerHTML = '<div class="service-card">Услуги ещё не созданы. Добавьте их во вкладке настроек.</div>';
    return;
  }

  const effectivePercentMap = buildEffectivePercentMap(appState.services, appState.weights);
  servicesList.innerHTML = appState.services.map((service, index) => {
    const weight = appState.weights[service.code] ?? 0;
    const weightClass = weight > 0 ? "weight-positive" : (weight < 0 ? "weight-negative" : "weight-neutral");
    return `
      <article class="service-card glass">
        <div>
          <strong class="truncate-text" title="${escapeHtml(`${index + 1}. ${service.name}`)}">${escapeHtml(`${index + 1}. ${service.name}`)}</strong>
          <div class="service-meta">${formatMoney(service.rate)} / ${escapeHtml(service.unit)}</div>
          <div class="service-meta">Процент: ${percentText(service, effectivePercentMap)}</div>
          <span class="category-pill ${categoryClass(service.category)}">${categoryLabel(service.category)}</span>
        </div>
        <label class="weight-field">
          <span>Вес</span>
          <input type="number" min="-10" max="10" step="1" value="${weight}" class="${weightClass}" data-weight-code="${service.code}" />
        </label>
      </article>
    `;
  }).join("");

  servicesList.querySelectorAll("[data-weight-code]").forEach((node) => {
    node.addEventListener("input", (event) => {
      const nextValue = Number(event.target.value);
      appState.weights[node.dataset.weightCode] = clampWeightValue(nextValue);
      renderServices();
    });
  });
}

function startEditService(id) {
  const service = appState.services.find((item) => item.id === id);
  if (!service || !appState.session?.authenticated) {
    return;
  }
  appState.editingServiceId = id;
  serviceNameInput.value = service.name;
  serviceUnitInput.value = service.unit;
  serviceRateInput.value = String(service.rate);
  serviceCategoryInput.value = service.category;
  servicePercentInput.value = typeof service.allocationPercent === "number" ? String(service.allocationPercent) : "";
  saveServiceButton.textContent = "\u041e\u0431\u043d\u043e\u0432\u0438\u0442\u044c \u0443\u0441\u043b\u0443\u0433\u0443";
  updateServicePercentHint();
  setActiveTab("settings");
  setServiceFormMessage(`Редактируется услуга «${service.name}».`, "muted");
}

function renderServicesAdmin() {
  saveServiceButton.disabled = false;
  resetServiceFormButton.disabled = false;
  [serviceNameInput, serviceUnitInput, serviceRateInput, serviceCategoryInput, servicePercentInput].forEach((node) => {
    node.disabled = false;
  });

  if (!appState.services.length) {
    servicesAdminList.innerHTML = '<div class="settings-empty">\u0421\u043f\u0438\u0441\u043e\u043a \u0443\u0441\u043b\u0443\u0433 \u043f\u0443\u0441\u0442.</div>';
    setServiceFormMessage("\u0423 \u0432\u0430\u0441 \u043f\u043e\u043a\u0430 \u043d\u0435\u0442 \u0443\u0441\u043b\u0443\u0433. \u041c\u043e\u0436\u043d\u043e \u0441\u043e\u0437\u0434\u0430\u0442\u044c \u043f\u0435\u0440\u0432\u0443\u044e \u043f\u0440\u044f\u043c\u043e \u0441\u0435\u0439\u0447\u0430\u0441.", "muted");
    updateServicePercentHint();
    return;
  }

  const effectivePercentMap = buildEffectivePercentMap(appState.services, appState.weights);
  servicesAdminList.innerHTML = appState.services.map((service) => `
    <article class="settings-item glass">
      <div>
        <strong class="truncate-text" title="${escapeHtml(service.name)}">${escapeHtml(service.name)}</strong>
        <div class="service-meta">${formatMoney(service.rate)} / ${escapeHtml(service.unit)}</div>
        <div class="service-meta">${categoryLabel(service.category)}, \u0442\u0435\u043a\u0443\u0449\u0438\u0439 \u043f\u0440\u043e\u0446\u0435\u043d\u0442 ${formatPercent(effectivePercentMap[service.code] ?? 0)}</div>
      </div>
      <div class="settings-item-actions">
        <button type="button" class="ghost-button" data-edit-service="${service.id}">\u0418\u0437\u043c\u0435\u043d\u0438\u0442\u044c</button>
        <button type="button" class="danger-button" data-delete-service="${service.id}">\u0423\u0434\u0430\u043b\u0438\u0442\u044c</button>
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

  setServiceFormMessage("\u041a\u0430\u0436\u0434\u044b\u0439 \u043f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u044c \u0443\u043f\u0440\u0430\u0432\u043b\u044f\u0435\u0442 \u0442\u043e\u043b\u044c\u043a\u043e \u0441\u0432\u043e\u0438\u043c \u0441\u043f\u0438\u0441\u043a\u043e\u043c \u0443\u0441\u043b\u0443\u0433 \u0438 \u043c\u043e\u0436\u0435\u0442 \u0440\u0435\u0434\u0430\u043a\u0442\u0438\u0440\u043e\u0432\u0430\u0442\u044c \u0441\u0432\u043e\u0438 \u0437\u0430\u043f\u0438\u0441\u0438.", "muted");
  updateServicePercentHint();
}

function renderUsers() {
  const sortedUsers = [...appState.users].sort((left, right) => {
    const leftPriority = userSortPriority(left.role);
    const rightPriority = userSortPriority(right.role);
    if (leftPriority !== rightPriority) {
      return leftPriority - rightPriority;
    }
    if ((left.createdAt || "") !== (right.createdAt || "")) {
      return String(left.createdAt || "").localeCompare(String(right.createdAt || ""));
    }
    return String(left.username || "").localeCompare(String(right.username || ""), "ru");
  });
  const viewerRole = normalizeRole(appState.session.user?.role);
  const isAdminSession = viewerRole === "admin";
  const canModerateUsers = Boolean(isAdminSession || appState.session.canModerate || appState.session.canManage);
  const roleChoices = roleChoicesForUser(viewerRole);

  userRoleInput.innerHTML = renderRoleOptions(roleChoices, userRoleInput.value);

  if (!roleChoices.some((choice) => choice.value === userRoleInput.value)) {
    userRoleInput.value = roleChoices[0]?.value ?? "";
  }

  createUserButton.disabled = !canModerateUsers || roleChoices.length === 0;
  userUsernameInput.disabled = !canModerateUsers || roleChoices.length === 0;
  userPasswordInput.disabled = !canModerateUsers || roleChoices.length === 0;
  userRoleInput.disabled = !canModerateUsers || roleChoices.length === 0;

  if (!canModerateUsers || roleChoices.length === 0) {
    usersList.innerHTML = '<div class="settings-empty">Управление пользователями доступно старшим и руководителям своего отдела, а также администратору.</div>';
    setUserFormMessage("Создавать пользователей можно только внутри своего отдела и своей зоны ответственности.", "muted");
    return;
  }
  if (!sortedUsers.length) {
    usersList.innerHTML = '<div class="settings-empty">Пользователи пока не созданы.</div>';
    setUserFormMessage("Можно создать первого пользователя в рамках доступных вам ролей.", "muted");
    return;
  }

  usersList.innerHTML = sortedUsers.map((user) => {
    const canChangeRole = Boolean(appState.session?.canAdmin)
      || (appState.session.canManage
        && roleDepartment(viewerRole) === roleDepartment(user.role));
    const canDelete = canDeleteManagedUser(appState.session.user?.role, user.role);
    const canResetPassword = roleCanCreateUsers(appState.session.user?.role) && canDeleteManagedUser(appState.session.user?.role, user.role);
    const roleOptions = renderRoleOptions(roleChoices, user.role);

    return `
      <article class="settings-item glass user-item">
        <div>
          <strong>${escapeHtml(user.username)}</strong>
          <div class="service-meta">${roleLabel(user.role)}</div>
          <div class="service-meta">Создан: ${formatDate(user.createdAt)}</div>
          ${canResetPassword ? `<div class="user-password-row"><input type="password" class="input-select compact-select inline-password-input" placeholder="Новый пароль" data-user-password-id="${user.id}" /><button type="button" class="ghost-button compact-action-button" data-apply-password-id="${user.id}">Сменить пароль</button></div>` : ""}
        </div>
        <div class="settings-item-actions stacked-actions">
          ${canChangeRole ? `<select class="input-select compact-select" data-user-role-id="${user.id}">${roleOptions}</select><button type="button" class="ghost-button" data-apply-role-id="${user.id}">Сменить роль</button>` : `<div class="service-meta">Смена ролей доступна только руководителю отдела и администратору.</div>`}
          <button type="button" class="danger-button" ${canDelete ? `data-delete-user-id="${user.id}"` : "disabled"}>Удалить</button>
        </div>
      </article>
    `;
  }).join("");

  usersList.querySelectorAll("[data-apply-role-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const userId = Number(button.dataset.applyRoleId);
      const roleInput = usersList.querySelector(`[data-user-role-id="${userId}"]`);
      await updateUserRole(userId, roleInput?.value ?? "support_employee");
    });
  });

  usersList.querySelectorAll("[data-apply-password-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const userId = Number(button.dataset.applyPasswordId);
      const passwordInput = usersList.querySelector(`[data-user-password-id="${userId}"]`);
      const nextPassword = stripSpaces(passwordInput?.value ?? "");
      if (!nextPassword) {
        setUserFormMessage("\u0412\u0432\u0435\u0434\u0438\u0442\u0435 \u043d\u043e\u0432\u044b\u0439 \u043f\u0430\u0440\u043e\u043b\u044c.", "error");
        passwordInput?.focus();
        return;
      }
      await resetUserPassword(userId, nextPassword);
      if (passwordInput) {
        passwordInput.value = "";
      }
    });
  });

  usersList.querySelectorAll("[data-delete-user-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      await deleteUser(Number(button.dataset.deleteUserId));
    });
  });

  setUserFormMessage("Пользователи отображаются и создаются только в пределах вашего отдела и уровня доступа.", "muted");
}

function userSortPriority(role) {
  switch (roleLevel(role)) {
    case 5:
      return 0;
    case 4:
      return 1;
    case 3:
      return 2;
    case 2:
      return 3;
    case 1:
      return 4;
    default:
      return 9;
  }
}

function renderArchiveDetails(saved) {
  const canCopyArchiveServices = Boolean(appState.session?.canAdmin && saved);
  const canDeleteArchive = Boolean((appState.session?.canManage || appState.session?.canAdmin) && saved);
  deleteArchiveButton.classList.toggle("hidden", !canDeleteArchive);
  deleteArchiveButton.disabled = !canDeleteArchive;
  copyArchiveServicesButton.classList.toggle("hidden", !canCopyArchiveServices);
  copyArchiveServicesButton.disabled = !canCopyArchiveServices;
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
  archiveMeta.innerHTML = `<span>Целевая сумма: <strong>${formatMoney(saved.targetAmount)}</strong></span><span>Автор: <strong>${escapeHtml(saved.createdBy || "Не указан")}</strong></span><span class="history-date">${formatDate(saved.createdAt)}</span>`;
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
  if (!appState.session?.canAdmin && authors.length) {
    if (!authors.includes(appState.archiveOwnerFilter)) {
      appState.archiveOwnerFilter = authors[0];
      archiveOwnerFilter.value = authors[0];
    }
  } else if (appState.session?.canAdmin) {
    archiveOwnerFilter.value = appState.archiveOwnerFilter || "all";
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

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function buildRandomWeightPayload() {
  const next = {};
  const groups = new Map();
  appState.services.forEach((service) => {
    const category = service.category || "closing";
    if (!groups.has(category)) {
      groups.set(category, []);
    }
    groups.get(category).push(service);
  });

  groups.forEach((services) => {
    if (!services.length) {
      return;
    }
    if (services.length === 1) {
      next[services[0].code] = randomInt(-4, 4);
      return;
    }

    let hasAccent = false;
    services.forEach((service) => {
      const value = randomInt(-4, 4);
      next[service.code] = value;
      if (value !== 0) {
        hasAccent = true;
      }
    });

    if (!hasAccent) {
      const target = services[randomInt(0, services.length - 1)];
      next[target.code] = randomInt(1, 4);
    }
  });

  return next;
}

async function randomizeCalculation() {
  const targetAmount = Number(amountInput.value);
  if (!Number.isInteger(targetAmount) || targetAmount <= 0) {
    setMessage("Сначала введите корректную сумму, а потом запускайте рандомизацию.", "error");
    amountInput.focus();
    return;
  }
  if (!appState.services.length) {
    setMessage("Сначала создайте хотя бы одну услугу для расчёта.", "error");
    return;
  }

  appState.weights = buildRandomWeightPayload();
  renderServices();
  setMessage("Рандомизация сместила проценты между услугами. Выполняю новый расчёт...", "muted");
  await calculate();
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

    const saved = await api.SaveCalculation({
      targetAmount: result.targetAmount,
      items: result.items,
    });
    appState.activeHistoryId = saved.id ?? null;
    await refreshBootstrap({ keepArchiveSelection: true });
    setMessage(`Расчёт сохранён в архив за ${saved.title}. Активно ${result.activeServices} из ${result.items.length} услуг на сумму ${formatMoney(result.totalAmount)}.`, "success");
  } catch (error) {
    appState.currentCalculation = null;
    renderResult(null);
    setMessage(error, "error");
  } finally {
    calculateButton.disabled = false;
  }
}

async function copyArchiveServicesToAdmin(id) {
  const current = appState.savedCalculations.find((item) => item.id === id);
  if (!current || !appState.session?.canAdmin) {
    return;
  }
  try {
    const result = await api.CopyArchiveServicesToAdmin(id);
    await refreshBootstrap({ keepArchiveSelection: true });
    const created = Number(result?.created ?? 0);
    const updated = Number(result?.updated ?? 0);
    setMessage(`Услуги из архива ${current.title} скопированы к администратору: создано ${created}, обновлено ${updated}.`, "success");
    setActiveTab("settings");
  } catch (error) {
    setMessage(error, "error");
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
    userRoleInput.value = roleChoicesForUser(appState.session.user?.role)[0]?.value ?? "";
    await refreshBootstrap({ keepArchiveSelection: true });
    setUserFormMessage("Пользователь создан.", "success");
  } catch (error) {
    setUserFormMessage(error, "error");
  } finally {
    createUserButton.disabled = false;
  }
}

async function resetUserPassword(userId, password) {
  try {
    await api.ResetUserPassword({ userID: userId, password });
    setUserFormMessage("Пароль пользователя обновлён.", "success");
  } catch (error) {
    setUserFormMessage(error, "error");
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
    setUserFormMessage(`Пользователь ${user.username} удалён.`, "success");
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
randomizeButton?.addEventListener("click", randomizeCalculation);
resetWeightsButton.addEventListener("click", resetWeights);
archiveOwnerFilter?.addEventListener("change", () => {
  appState.archiveOwnerFilter = archiveOwnerFilter.value || "all";
  appState.activeHistoryId = null;
  renderHistory();
});

deleteArchiveButton.addEventListener("click", async () => {
  if (appState.activeHistoryId !== null) {
    await deleteCalculation(appState.activeHistoryId);
  }
});
copyArchiveServicesButton?.addEventListener("click", async () => {
  if (appState.activeHistoryId !== null) {
    await copyArchiveServicesToAdmin(appState.activeHistoryId);
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

















