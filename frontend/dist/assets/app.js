const appState = {
  session: { authenticated: false, canManage: false, canAdmin: false, canModerate: false, user: null },
  services: [],
  users: [],
  savedCalculations: [],
  currentCalculation: null,
  activeHistoryId: null,
  editingArchiveId: null,
  archiveOwnerFilter: "all",
  activeTab: "calculator",
  weights: {},
  editingServiceId: 0,
  defaultGroupPercent: { primary: 0.79, secondary: 0.2, closing: 0.01 },
  exportDraft: {
    actNumber: 1,
    employeeFullName: "",
    contractCode: "1",
    contractSpbksNumber: "",
    contractGrizablNumber: "",
    contractDate: "",
  },
};

const api = window.go?.appcore?.App || window.go?.main?.App;

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
const actNumberInput = document.getElementById("act-number-input");
const employeeFullNameInput = document.getElementById("employee-full-name-input");
const contractCodeInput = document.getElementById("contract-code-input");

const contractSPBKSNumberInput = document.getElementById("contract-spbks-number-input");

const contractGrizablNumberInput = document.getElementById("contract-grizabl-number-input");
const contractDateInput = document.getElementById("contract-date-input");
const contractDetailsButton = document.getElementById("contract-details-button");
const contractDetailsSummary = document.getElementById("contract-details-summary");
const saveOwnFullNameButton = document.getElementById("save-own-fullname-button");
const calculateButton = document.getElementById("calculate-button");
const resetWeightsButton = document.getElementById("reset-weights-button");
const deleteArchiveButton = document.getElementById("delete-archive-button");
const downloadPdfButton = document.getElementById("download-pdf-button");
const actNumberVisibleInput = document.getElementById("act-number-visible-input");
const previousActNumber = document.getElementById("previous-act-number");
const copyArchiveServicesButton = document.getElementById("copy-archive-services-button");
const editArchiveButton = document.getElementById("edit-archive-button");
const cancelArchiveEditButton = document.getElementById("cancel-archive-edit-button");
const saveArchiveEditButton = document.getElementById("save-archive-edit-button");
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
const contractModal = document.getElementById("contract-modal");
const contractBackdrop = document.getElementById("contract-backdrop");
const contractCancel = document.getElementById("contract-cancel");
const contractSave = document.getElementById("contract-save");
const contractModalSPBKSNumberInput = document.getElementById("contract-modal-spbks-number-input");
const contractModalGrizablNumberInput = document.getElementById("contract-modal-grizabl-number-input");
const contractModalCodeSPBKS = document.getElementById("contract-modal-code-spbks");
const contractModalCodeGrizabl = document.getElementById("contract-modal-code-grizabl");
const contractModalDateInput = document.getElementById("contract-modal-date-input");
const contractModalFullNameInput = document.getElementById("contract-modal-fullname-input");

let confirmResolver = null;

function renderActNumberControls() {
  const actNumber = clampPositiveInteger(appState.exportDraft.actNumber, (appState.session?.user?.lastActNumber || 0) + 1);
  const lastActNumber = Number.isInteger(Number(appState.session?.user?.lastActNumber)) ? Number(appState.session?.user?.lastActNumber) : 0;
  appState.exportDraft.actNumber = actNumber;
  if (actNumberVisibleInput) {
    actNumberVisibleInput.value = String(actNumber);
    actNumberVisibleInput.disabled = !appState.session?.authenticated;
  }
  if (previousActNumber) {
    previousActNumber.textContent = `\u041f\u043e\u0441\u043b\u0435\u0434\u043d\u0438\u0439 \u043d\u043e\u043c\u0435\u0440 \u0430\u043a\u0442\u0430: ${lastActNumber > 0 ? lastActNumber : "\u2014"}`;
  }
}

function clampPositiveInteger(value, fallback = 1) {
  const normalized = Number(value);
  if (Number.isInteger(normalized) && normalized > 0) {
    return normalized;
  }
  const safeFallback = Number(fallback);
  return Number.isInteger(safeFallback) && safeFallback > 0 ? safeFallback : 1;
}

function formatMoney(value) {
  return `${new Intl.NumberFormat("ru-RU").format(value)} \u0440`;
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

function openConfirmModal({ title, text, confirmLabel = "\u0423\u0434\u0430\u043b\u0438\u0442\u044c" }) {
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
    title: `\u0423\u0434\u0430\u043b\u0438\u0442\u044c ${entityLabel}?`,
    text: `\u041f\u043e\u0434\u0442\u0432\u0435\u0440\u0434\u0438\u0442\u0435 \u0443\u0434\u0430\u043b\u0435\u043d\u0438\u0435: ${entityName}. \u042d\u0442\u043e \u0434\u0435\u0439\u0441\u0442\u0432\u0438\u0435 \u043d\u0435\u043b\u044c\u0437\u044f \u043e\u0442\u043c\u0435\u043d\u0438\u0442\u044c.`,
    confirmLabel: "\u0423\u0434\u0430\u043b\u0438\u0442\u044c",
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

async function persistActNumber() {
  const currentUser = appState.session?.user;
  if (!currentUser?.id) {
    return;
  }

  const nextValue = clampPositiveInteger(actNumberVisibleInput?.value, appState.exportDraft.actNumber || ((currentUser.lastActNumber || 0) + 1));
  appState.exportDraft.actNumber = nextValue;
  updateHiddenExportInputs();
  renderActNumberControls();
}

function updateHiddenExportInputs() {

  if (actNumberInput) {

    actNumberInput.value = String(appState.exportDraft.actNumber || 1);

  }

  if (employeeFullNameInput) {

    employeeFullNameInput.value = appState.exportDraft.employeeFullName || "";

  }

  if (contractCodeInput) {

    contractCodeInput.value = appState.exportDraft.contractCode || "1";

  }

  if (contractSPBKSNumberInput) {

    contractSPBKSNumberInput.value = appState.exportDraft.contractSpbksNumber || "";

  }

  if (contractGrizablNumberInput) {

    contractGrizablNumberInput.value = appState.exportDraft.contractGrizablNumber || "";

  }

  if (contractDateInput) {

    contractDateInput.value = appState.exportDraft.contractDate || "";

  }

  renderActNumberControls();
}

function contractTemplateLabel(code) {
  return code === "2"
    ? "\u00ab\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c\u00bb"
    : "\u00ab\u0421\u0430\u043d\u043a\u0442-\u041f\u0435\u0442\u0435\u0440\u0431\u0443\u0440\u0433\u0441\u043a\u0438\u0435 \u043a\u043e\u043c\u043f\u044c\u044e\u0442\u0435\u0440\u043d\u044b\u0435 \u0441\u0435\u0442\u0438\u00bb";
}



function getSelectedContractNumber() {

  return appState.exportDraft.contractCode === "2"

    ? String(appState.exportDraft.contractGrizablNumber || "").trim()

    : String(appState.exportDraft.contractSpbksNumber || "").trim();

}



function renderContractDetailsSummary() {

  if (!contractDetailsSummary) {

    return;

  }

  const fullName = String(appState.exportDraft.employeeFullName || "").trim();

  const contractDate = String(appState.exportDraft.contractDate || "").trim();

  const selectedNumber = getSelectedContractNumber();

  if (!selectedNumber || !contractDate || !fullName) {

    contractDetailsSummary.textContent = "\u0414\u043b\u044f \u0432\u044b\u0433\u0440\u0443\u0437\u043a\u0438 \u0430\u043a\u0442\u0430 \u0443\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430, \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0438 \u0432\u0430\u0448\u0435 \u0424\u0418\u041e.";

    contractDetailsSummary.className = "message muted compact-note";

    return;

  }

  contractDetailsSummary.textContent = `\u0412\u044b\u0431\u0440\u0430\u043d \u0434\u043e\u0433\u043e\u0432\u043e\u0440: ${contractTemplateLabel(appState.exportDraft.contractCode)} \u2116${selectedNumber} \u043e\u0442 ${formatArchiveTitle(new Date(contractDate))}. \u0424\u0418\u041e: ${fullName}.`;

  contractDetailsSummary.className = "message success compact-note";

}

function syncExportFields() {

  const sessionUser = appState.session?.user;

  appState.exportDraft.actNumber = clampPositiveInteger(sessionUser?.lastActNumber, appState.exportDraft.actNumber || 1);

  const currentFullName = String(sessionUser?.fullName ?? "").trim();

  if (!String(appState.exportDraft.employeeFullName || "").trim() && currentFullName) {

    appState.exportDraft.employeeFullName = currentFullName;

  }

  if (!String(appState.exportDraft.contractSpbksNumber || "").trim()) {

    appState.exportDraft.contractSpbksNumber = String(sessionUser?.contractSPBKSNumber ?? "").trim();

  }

  if (!String(appState.exportDraft.contractGrizablNumber || "").trim()) {

    appState.exportDraft.contractGrizablNumber = String(sessionUser?.contractGrizablNumber ?? "").trim();

  }

  if (!String(appState.exportDraft.contractDate || "").trim()) {

    appState.exportDraft.contractDate = String(sessionUser?.contractSignedAt ?? "").trim() || new Date().toISOString().slice(0, 10);

  }

  if (!["1", "2"].includes(String(appState.exportDraft.contractCode || ""))) {

    appState.exportDraft.contractCode = "1";

  }

  updateHiddenExportInputs();

  renderContractDetailsSummary();

}

function openContractModal() {

  if (contractModalSPBKSNumberInput) {

    contractModalSPBKSNumberInput.value = appState.exportDraft.contractSpbksNumber || "";

  }

  if (contractModalGrizablNumberInput) {

    contractModalGrizablNumberInput.value = appState.exportDraft.contractGrizablNumber || "";

  }

  if (contractModalDateInput) {

    contractModalDateInput.value = appState.exportDraft.contractDate || new Date().toISOString().slice(0, 10);

  }

  if (contractModalFullNameInput) {

    contractModalFullNameInput.value = appState.exportDraft.employeeFullName || appState.session?.user?.fullName || "";

  }

  if (contractModalCodeGrizabl && contractModalCodeSPBKS) {

    contractModalCodeGrizabl.checked = appState.exportDraft.contractCode === "2";

    contractModalCodeSPBKS.checked = appState.exportDraft.contractCode !== "2";

  }

  contractModal?.classList.remove("hidden");

  contractModal?.setAttribute("aria-hidden", "false");

  contractModalSPBKSNumberInput?.focus();

}

function closeContractModal() {
  contractModal?.classList.add("hidden");
  contractModal?.setAttribute("aria-hidden", "true");
}

async function saveContractDetails() {

  const contractSpbksNumber = String(contractModalSPBKSNumberInput?.value ?? "").trim();

  const contractGrizablNumber = String(contractModalGrizablNumberInput?.value ?? "").trim();

  const contractDate = String(contractModalDateInput?.value ?? "").trim();

  const fullName = String(contractModalFullNameInput?.value ?? "").trim();

  const contractCode = contractModalCodeGrizabl?.checked ? "2" : "1";

  const selectedNumber = contractCode === "2" ? contractGrizablNumber : contractSpbksNumber;

  if (!selectedNumber) {

    setMessage("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430 \u0434\u043b\u044f \u0432\u044b\u0431\u0440\u0430\u043d\u043d\u043e\u0433\u043e \u0448\u0430\u0431\u043b\u043e\u043d\u0430.", "error");

    if (contractCode === "2") {
      contractModalGrizablNumberInput?.focus();
    } else {
      contractModalSPBKSNumberInput?.focus();
    }

    return;

  }

  if (!contractDate) {

    setMessage("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430.", "error");

    contractModalDateInput?.focus();

    return;

  }

  if (!fullName) {

    setMessage("\u0423\u043a\u0430\u0436\u0438\u0442\u0435 \u0432\u0430\u0448\u0435 \u0424\u0418\u041e \u0434\u043b\u044f \u0430\u043a\u0442\u0430.", "error");

    contractModalFullNameInput?.focus();

    return;

  }

  try {

    const currentUser = appState.session?.user;

    const updated = await api.UpdateUserContractDetails({

      userID: currentUser?.id ?? 0,

      fullName,

      preferredContractCode: contractCode,

      contractSPBKSNumber: contractSpbksNumber,

      contractGrizablNumber: contractGrizablNumber,

      contractSignedAt: contractDate,

    });

    if (updated) {

      if (appState.session?.user?.id === updated.id) {

        appState.session.user = { ...appState.session.user, ...updated };

      }

      appState.exportDraft.contractCode = contractCode;

      appState.exportDraft.contractSpbksNumber = contractSpbksNumber;

      appState.exportDraft.contractGrizablNumber = contractGrizablNumber;

      appState.exportDraft.contractDate = contractDate;

      appState.exportDraft.employeeFullName = fullName;

      updateHiddenExportInputs();

      renderContractDetailsSummary();

      closeContractModal();

      setMessage("\u0414\u0430\u043d\u043d\u044b\u0435 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430 \u0441\u043e\u0445\u0440\u0430\u043d\u0435\u043d\u044b. \u0422\u0435\u043f\u0435\u0440\u044c \u043c\u043e\u0436\u043d\u043e \u0441\u043a\u0430\u0447\u0430\u0442\u044c \u0430\u043a\u0442 PDF.", "success");

    }
  } catch (error) {

    setMessage(error, "error");

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

function escapeAttribute(value) {
  return escapeHtml(value).replaceAll("'", "&#39;");
}

function stripSpaces(value) {
  return String(value ?? "").replace(/\s+/g, "");
}

function categoryLabel(category) {
  switch (category) {
    case "primary":
      return "\u041e\u0441\u043d\u043e\u0432\u043d\u0430\u044f";
    case "secondary":
      return "\u0412\u0442\u043e\u0440\u0438\u0447\u043d\u0430\u044f";
    case "closing":
      return "\u0417\u0430\u043a\u0440\u044b\u0432\u0430\u044e\u0449\u0430\u044f";
    default:
      return "\u0411\u0435\u0437 \u0433\u0440\u0443\u043f\u043f\u044b";
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
    admin: { level: 5, department: "global", label: "\u0410\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440" },
    global_director: { level: 4, department: "global", label: "\u0413\u0435\u043d\u0435\u0440\u0430\u043b\u044c\u043d\u044b\u0439 \u0434\u0438\u0440\u0435\u043a\u0442\u043e\u0440" },
    executive_director: { level: 4, department: "global", label: "\u0418\u0441\u043f\u043e\u043b\u043d\u0438\u0442\u0435\u043b\u044c\u043d\u044b\u0439 \u0434\u0438\u0440\u0435\u043a\u0442\u043e\u0440" },
    technical_director: { level: 4, department: "global", label: "\u0422\u0435\u0445\u043d\u0438\u0447\u0435\u0441\u043a\u0438\u0439 \u0434\u0438\u0440\u0435\u043a\u0442\u043e\u0440" },
    support_head: { level: 3, department: "support", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0422\u0435\u0445. \u041f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0438" },
    support_sysadmin: { level: 3, department: "support", label: "\u0421\u0438\u0441\u0442\u0435\u043c\u043d\u044b\u0439 \u0430\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440" },
    support_senior: { level: 2, department: "support", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u0441\u043f\u0435\u0446\u0438\u0430\u043b\u0438\u0441\u0442 \u0442\u0435\u0445\u043f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0438" },
    support_employee: { level: 1, department: "support", label: "\u0421\u043f\u0435\u0446\u0438\u0430\u043b\u0438\u0441\u0442 \u0442\u0435\u0445\u043f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0438" },
    technical_head: { level: 3, department: "technical", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0442\u0435\u0445\u043d\u0438\u0447\u0435\u0441\u043a\u043e\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430" },
    technical_senior: { level: 2, department: "technical", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u0442\u0435\u0445\u043d\u0438\u043a" },
    technical_employee: { level: 1, department: "technical", label: "\u0422\u0435\u0445\u043d\u0438\u043a" },
    telecom_construction_director: { level: 3, department: "telecom", label: "\u0414\u0438\u0440\u0435\u043a\u0442\u043e\u0440 \u043f\u043e \u0441\u0442\u0440\u043e\u0438\u0442\u0435\u043b\u044c\u0441\u0442\u0432\u0443" },
    telecom_construction_head: { level: 3, department: "telecom", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0441\u0442\u0440\u043e\u0438\u0442\u0435\u043b\u044c\u043d\u043e\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430" },
    telecom_senior_vols: { level: 2, department: "telecom", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u043c\u043e\u043d\u0442\u0430\u0436\u043d\u0438\u043a \u0412\u041e\u041b\u0421" },
    telecom_senior_lvs: { level: 2, department: "telecom", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u043c\u043e\u043d\u0442\u0430\u0436\u043d\u0438\u043a \u041b\u0412\u0421" },
    telecom_employee_vols: { level: 1, department: "telecom", label: "\u041c\u043e\u043d\u0442\u0430\u0436\u043d\u0438\u043a \u0412\u041e\u041b\u0421" },
    telecom_employee_lvs: { level: 1, department: "telecom", label: "\u041c\u043e\u043d\u0442\u0430\u0436\u043d\u0438\u043a \u041b\u0412\u0421" },
    skud_head: { level: 3, department: "skud", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u043e\u0442\u0434\u0435\u043b\u0430 \u0442\u0435\u0445\u043d\u0438\u0447\u0435\u0441\u043a\u043e\u0433\u043e \u043e\u0431\u0441\u043b\u0443\u0436\u0438\u0432\u0430\u043d\u0438\u044f \u0421\u041a\u0423\u0414" },
    skud_project_manager: { level: 2, department: "skud", label: "\u041c\u0435\u043d\u0435\u0434\u0436\u0435\u0440 \u043f\u0440\u043e\u0435\u043a\u0442\u043e\u0432 \u0421\u041a\u0423\u0414" },
    skud_senior_service_engineer: { level: 2, department: "skud", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u0441\u0435\u0440\u0432\u0438\u0441\u043d\u044b\u0439 \u0438\u043d\u0436\u0435\u043d\u0435\u0440 \u0421\u041a\u0423\u0414" },
    skud_senior_installer: { level: 2, department: "skud", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u043c\u043e\u043d\u0442\u0430\u0436\u043d\u0438\u043a \u0421\u041a\u0423\u0414" },
    skud_service_engineer: { level: 1, department: "skud", label: "\u0421\u0435\u0440\u0432\u0438\u0441\u043d\u044b\u0439 \u0438\u043d\u0436\u0435\u043d\u0435\u0440 \u0421\u041a\u0423\u0414" },
    skud_installer: { level: 1, department: "skud", label: "\u041c\u043e\u043d\u0442\u0430\u0436\u043d\u0438\u043a \u0421\u041a\u0423\u0414" },
    approval_head: { level: 3, department: "approval", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0441\u043e\u0433\u043b\u0430\u0441\u043e\u0432\u0430\u043d\u0438\u044f" },
    approval_senior: { level: 2, department: "approval", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u043c\u0435\u043d\u0435\u0434\u0436\u0435\u0440 \u0441\u043e\u0433\u043b\u0430\u0441\u043e\u0432\u0430\u043d\u0438\u044f" },
    approval_employee: { level: 1, department: "approval", label: "\u041c\u0435\u043d\u0435\u0434\u0436\u0435\u0440 \u043f\u043e \u0441\u043e\u0433\u043b\u0430\u0441\u043e\u0432\u0430\u043d\u0438\u044e" },
    marketing_head: { level: 3, department: "marketing", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u043e\u0442\u0434\u0435\u043b\u0430 \u0440\u0435\u043a\u043b\u0430\u043c\u044b \u0438 \u043c\u0430\u0440\u043a\u0435\u0442\u0438\u043d\u0433\u0430" },
    marketing_courier: { level: 1, department: "marketing", label: "\u041a\u0443\u0440\u044c\u0435\u0440" },
    commercial_director: { level: 4, department: "commercial", label: "\u0414\u0438\u0440\u0435\u043a\u0442\u043e\u0440 \u043a\u043e\u043c\u043c\u0435\u0440\u0447\u0435\u0441\u043a\u043e\u0433\u043e \u0431\u043b\u043e\u043a\u0430" },
    commercial_subscriber_head: { level: 3, department: "commercial", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0430\u0431\u043e\u043d\u0435\u043d\u0442\u0441\u043a\u043e\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430" },
    commercial_active_sales_head: { level: 3, department: "commercial", label: "\u041c\u0435\u043d\u0435\u0434\u0436\u0435\u0440 \u0430\u043a\u0442\u0438\u0432\u043d\u044b\u0445 \u043f\u0440\u043e\u0434\u0430\u0436" },
    commercial_senior_mrk: { level: 2, department: "commercial", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u041c\u0420\u041a" },
    commercial_senior_mryu: { level: 2, department: "commercial", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u041c\u0420\u042e" },
    commercial_employee_mrk: { level: 1, department: "commercial", label: "\u041c\u0420\u041a" },
    commercial_employee_mryu: { level: 1, department: "commercial", label: "\u041c\u0420\u042e" },
    finance_head: { level: 3, department: "finance", label: "\u0413\u043b. \u0431\u0443\u0445\u0433\u0430\u043b\u0442\u0435\u0440" },
    finance_employee: { level: 1, department: "finance", label: "\u041f\u043e\u043c\u043e\u0449\u043d\u0438\u043a \u0431\u0443\u0445\u0433\u0430\u043b\u0442\u0435\u0440\u0430" },
    legal_employee: { level: 1, department: "legal", label: "\u042e\u0440\u0438\u0441\u0442" },
    development_head: { level: 3, department: "development", label: "\u0420\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044c \u0433\u0440\u0443\u043f\u043f\u044b \u0440\u0430\u0437\u0440\u0430\u0431\u043e\u0442\u043a\u0438" },
    development_senior: { level: 2, department: "development", label: "\u0421\u0442\u0430\u0440\u0448\u0438\u0439 \u0440\u0430\u0437\u0440\u0430\u0431\u043e\u0442\u0447\u0438\u043a" },
    development_employee: { level: 1, department: "development", label: "\u0420\u0430\u0437\u0440\u0430\u0431\u043e\u0442\u0447\u0438\u043a" },
  };
  return catalog[normalizeRole(role)] ?? { level: 0, department: "support", label: "\u041f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u044c" };
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
      return "\u0413\u043b\u043e\u0431\u0430\u043b\u044c\u043d\u044b\u0435 \u0440\u043e\u043b\u0438";
    case "support":
      return "\u0422\u0435\u0445\u043f\u043e\u0434\u0434\u0435\u0440\u0436\u043a\u0430";
    case "technical":
      return "\u0422\u0435\u0445\u043d\u0438\u0447\u0435\u0441\u043a\u0438\u0439 \u043e\u0442\u0434\u0435\u043b";
    case "telecom":
      return "\u0422\u0435\u043b\u0435\u043a\u043e\u043c / \u0412\u041e\u041b\u0421 / \u041b\u0412\u0421";
    case "skud":
      return "\u0421\u041a\u0423\u0414";
    case "approval":
      return "\u0421\u043e\u0433\u043b\u0430\u0441\u043e\u0432\u0430\u043d\u0438\u0435";
    case "marketing":
      return "\u041c\u0430\u0440\u043a\u0435\u0442\u0438\u043d\u0433";
    case "commercial":
      return "\u041a\u043e\u043c\u043c\u0435\u0440\u0447\u0435\u0441\u043a\u0438\u0439 \u0431\u043b\u043e\u043a";
    case "finance":
      return "\u0424\u0438\u043d\u0430\u043d\u0441\u044b";
    case "legal":
      return "\u042e\u0440\u0438\u0434\u0438\u0447\u0435\u0441\u043a\u0438\u0439 \u043e\u0442\u0434\u0435\u043b";
    case "development":
      return "\u0420\u0430\u0437\u0440\u0430\u0431\u043e\u0442\u043a\u0430";
    default:
      return "\u041e\u0442\u0434\u0435\u043b";
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

function normalizeErrorText(text, fallback = "\u041f\u0440\u043e\u0438\u0437\u043e\u0448\u043b\u0430 \u043e\u0448\u0438\u0431\u043a\u0430.") {
  const rawValue = typeof text === "object" && text !== null
    ? (text.message ?? text.error ?? text.reason ?? text.details ?? String(text))
    : text;
  const value = String(rawValue ?? "").replace(/^Error:\s*/, "").trim();
  if (!value || value === "[object Object]") {
    return fallback;
  }
  return value;
}

function setMessage(text, type = "muted") {
  messageNode.textContent = normalizeErrorText(text, "\u041f\u0440\u043e\u0438\u0437\u043e\u0448\u043b\u0430 \u043e\u0448\u0438\u0431\u043a\u0430.");
  messageNode.className = `message ${type}`;
}

function setLoginMessage(text, type = "muted") {
  loginMessage.textContent = normalizeErrorText(text, "\u041e\u0448\u0438\u0431\u043a\u0430 \u0432\u0445\u043e\u0434\u0430. \u041f\u0440\u043e\u0432\u0435\u0440\u044c\u0442\u0435 \u043b\u043e\u0433\u0438\u043d \u0438 \u043f\u0430\u0440\u043e\u043b\u044c.");
  loginMessage.className = `message ${type}`;
}

function setServiceFormMessage(text, type = "muted") {
  serviceFormMessage.textContent = normalizeErrorText(text, "\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u0441\u043e\u0445\u0440\u0430\u043d\u0438\u0442\u044c \u0443\u0441\u043b\u0443\u0433\u0443.");
  serviceFormMessage.className = `message ${type}`;
}

function setUserFormMessage(text, type = "muted") {
  userFormMessage.textContent = normalizeErrorText(text, "\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u0432\u044b\u043f\u043e\u043b\u043d\u0438\u0442\u044c \u0434\u0435\u0439\u0441\u0442\u0432\u0438\u0435 \u0441 \u043f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u0435\u043c.");
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
  const previousUserID = appState.session?.user?.id ?? 0;
  appState.session = data.session ?? { authenticated: false, canManage: false, canAdmin: false, canModerate: false, user: null };
  const currentUserID = appState.session?.user?.id ?? 0;
  if (previousUserID !== currentUserID) {
    appState.exportDraft = {
      actNumber: 1,
      employeeFullName: String(appState.session?.user?.fullName ?? "").trim(),
      contractCode: "1",
      contractSpbksNumber: String(appState.session?.user?.contractSPBKSNumber ?? "").trim(),
      contractGrizablNumber: String(appState.session?.user?.contractGrizablNumber ?? "").trim(),
      contractDate: String(appState.session?.user?.contractSignedAt ?? "").trim(),
    };
  }
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
  syncExportFields();
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

function serviceTakesPartInCalculation(service) {
  return !(typeof service?.allocationPercent === 'number' && service.allocationPercent <= 0);
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
  servicePercentInput.placeholder = `\u041f\u0443\u0441\u0442\u043e = ${placeholderValue} \u043f\u043e \u0443\u043c\u043e\u043b\u0447\u0430\u043d\u0438\u044e`;
  servicePercentHint.textContent = nextPercent !== null && Number.isFinite(nextPercent)
    ? `\u0421\u0435\u0439\u0447\u0430\u0441 \u0434\u043b\u044f \u0443\u0441\u043b\u0443\u0433\u0438 \u0431\u0443\u0434\u0435\u0442 \u0443\u0441\u0442\u0430\u043d\u043e\u0432\u043b\u0435\u043d \u043f\u0440\u043e\u0446\u0435\u043d\u0442 ${formatPercent(nextPercent)}.`
    : `\u0421\u0435\u0439\u0447\u0430\u0441 \u0434\u043b\u044f \u0443\u0441\u043b\u0443\u0433\u0438 \u0434\u0435\u0439\u0441\u0442\u0432\u0443\u0435\u0442 \u043f\u0440\u043e\u0446\u0435\u043d\u0442 \u043f\u043e \u0433\u0440\u0443\u043f\u043f\u0435: ${placeholderValue}. \u041e\u043d \u0440\u0430\u0441\u0441\u0447\u0438\u0442\u0430\u043d \u0441 \u0443\u0447\u0451\u0442\u043e\u043c \u043a\u043e\u043b\u0438\u0447\u0435\u0441\u0442\u0432\u0430 \u0443\u0441\u043b\u0443\u0433 \u0432 \u044d\u0442\u043e\u0439 \u0433\u0440\u0443\u043f\u043f\u0435.`;
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
  serviceUnitInput.value = "ч.";
  serviceRateInput.value = "";
  serviceCategoryInput.value = "primary";
  servicePercentInput.value = "";
  saveServiceButton.textContent = "\u0421\u043e\u0445\u0440\u0430\u043d\u0438\u0442\u044c \u0443\u0441\u043b\u0443\u0433\u0443";
  updateServicePercentHint();
}

function renderResult(result) {
  if (!result) {
    resultTotal.textContent = "0 \u0440";
    summaryTotal.textContent = "0 \u0440";
    summaryItems.textContent = `0/${appState.services.length}`;
    exactStatus.textContent = "\u041e\u0436\u0438\u0434\u0430\u043d\u0438\u0435 \u0440\u0430\u0441\u0447\u0451\u0442\u0430";
    if (downloadPdfButton) {
      downloadPdfButton.disabled = true;
    }
    if (actNumberVisibleInput) {
      actNumberVisibleInput.disabled = true;
    }
    renderActNumberControls();
    resultBody.innerHTML = '<tr><td colspan="7" class="placeholder">\u0420\u0435\u0437\u0443\u043b\u044c\u0442\u0430\u0442\u044b \u043f\u043e\u044f\u0432\u044f\u0442\u0441\u044f \u0437\u0434\u0435\u0441\u044c \u043f\u043e\u0441\u043b\u0435 \u0440\u0430\u0441\u0447\u0451\u0442\u0430.</td></tr>';
    return;
  }

  const weightMap = Object.fromEntries(result.items.map((item) => [item.serviceCode, item.weight ?? appState.weights[item.serviceCode] ?? 0]));
  const effectivePercentMap = buildEffectivePercentMap(result.items, weightMap);
  const isExact = Boolean(result.foundExact ?? result.exactMatch ?? (result.totalAmount === result.targetAmount));
  resultTotal.textContent = formatMoney(result.totalAmount);
  summaryTotal.textContent = formatMoney(result.totalAmount);
  summaryItems.textContent = `${result.activeServices}/${result.items.length}`;
  exactStatus.textContent = isExact ? "\u0422\u043e\u0447\u043d\u043e\u0435 \u0441\u043e\u0432\u043f\u0430\u0434\u0435\u043d\u0438\u0435 \u043d\u0430\u0439\u0434\u0435\u043d\u043e" : "\u0415\u0441\u0442\u044c \u043e\u0442\u043a\u043b\u043e\u043d\u0435\u043d\u0438\u0435 \u043e\u0442 \u0446\u0435\u043b\u0435\u0432\u043e\u0439 \u0441\u0443\u043c\u043c\u044b";
  if (downloadPdfButton) {
    downloadPdfButton.disabled = !isExact;
  }
  if (actNumberVisibleInput) {
    actNumberVisibleInput.disabled = false;
  }
  renderActNumberControls();
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
  defaultPercentages.textContent = `\u041f\u043e \u0443\u043c\u043e\u043b\u0447\u0430\u043d\u0438\u044e: \u043e\u0441\u043d\u043e\u0432\u043d\u044b\u0435 ${Math.round((groups.primary ?? 0) * 100)}%, \u0432\u0442\u043e\u0440\u0438\u0447\u043d\u044b\u0435 ${Math.round((groups.secondary ?? 0) * 100)}%, \u0437\u0430\u043a\u0440\u044b\u0432\u0430\u044e\u0449\u0438\u0435 ${Math.round((groups.closing ?? 0) * 100)}%. \u0412\u043d\u0443\u0442\u0440\u0438 \u0433\u0440\u0443\u043f\u043f\u044b \u044d\u0442\u043e\u0442 \u043f\u0440\u043e\u0446\u0435\u043d\u0442 \u0434\u0435\u043b\u0438\u0442\u0441\u044f \u0440\u0430\u0432\u043d\u043e\u043c\u0435\u0440\u043d\u043e, \u0435\u0441\u043b\u0438 \u0434\u043b\u044f \u0443\u0441\u043b\u0443\u0433\u0438 \u043d\u0435 \u0437\u0430\u0434\u0430\u043d \u0441\u0432\u043e\u0439 \u043f\u0440\u043e\u0446\u0435\u043d\u0442. \u0412\u0435\u0441 \u043c\u043e\u0436\u043d\u043e \u0437\u0430\u0434\u0430\u0432\u0430\u0442\u044c \u043e\u0442 -10 \u0434\u043e 10: \u043f\u043b\u044e\u0441 \u0443\u0441\u0438\u043b\u0438\u0432\u0430\u0435\u0442 \u0443\u0441\u043b\u0443\u0433\u0443, \u043c\u0438\u043d\u0443\u0441 \u043e\u0441\u043b\u0430\u0431\u043b\u044f\u0435\u0442.`;
}

function renderServices() {
  if (!appState.services.length) {
    servicesList.innerHTML = '<div class="service-card">\u0423\u0441\u043b\u0443\u0433\u0438 \u0435\u0449\u0451 \u043d\u0435 \u0441\u043e\u0437\u0434\u0430\u043d\u044b. \u0414\u043e\u0431\u0430\u0432\u044c\u0442\u0435 \u0438\u0445 \u0432\u043e \u0432\u043a\u043b\u0430\u0434\u043a\u0435 \u043d\u0430\u0441\u0442\u0440\u043e\u0435\u043a.</div>';
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
          <div class="service-meta">\u041f\u0440\u043e\u0446\u0435\u043d\u0442: ${percentText(service, effectivePercentMap)}</div>
          <span class="category-pill ${categoryClass(service.category)}">${categoryLabel(service.category)}</span>
        </div>
        <label class="weight-field">
          <span>\u0412\u0435\u0441</span>
          <input type="text" inputmode="numeric" autocomplete="off" value="${weight}" class="${weightClass}" data-weight-code="${service.code}" />
        </label>
      </article>
    `;
  }).join("");


  servicesList.querySelectorAll("[data-weight-code]").forEach((node) => {
    node.addEventListener("input", (event) => {
      const rawValue = String(event.target.value ?? "");
      if (/^-?\d*$/.test(rawValue)) {
        return;
      }
      event.target.value = rawValue.replace(/[^\d-]/g, "").replace(/(?!^)-/g, "");
    });
    const applyWeightValue = (event) => {
      const rawValue = String(event.target.value ?? "").trim();
      const nextValue = rawValue === "" || rawValue === "-" ? 0 : Number(rawValue);
      if (!Number.isFinite(nextValue)) {
        event.target.value = String(appState.weights[node.dataset.weightCode] ?? 0);
        return;
      }
      appState.weights[node.dataset.weightCode] = clampWeightValue(nextValue);
      renderServices();
    };
    node.addEventListener("change", applyWeightValue);
    node.addEventListener("blur", applyWeightValue);
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
  setServiceFormMessage(`\u0420\u0435\u0434\u0430\u043a\u0442\u0438\u0440\u0443\u0435\u0442\u0441\u044f \u0443\u0441\u043b\u0443\u0433\u0430 \u00ab${service.name}\u00bb.`, "muted");
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

async function updateUserFullName(userId, fullName, options = {}) {
  const normalized = String(fullName ?? "").trim();
  const focusNode = options.focusNode ?? null;
  if (!normalized) {
    const fallbackMessage = options.emptyMessage ?? "Введите ФИО сотрудника.";
    if (options.self) {
      setMessage(fallbackMessage, "error");
    } else {
      setUserFormMessage(fallbackMessage, "error");
    }
    focusNode?.focus?.();
    return null;
  }

  try {
    const updated = await api.UpdateUserFullName({ userID: userId, fullName: normalized });
    await refreshBootstrap({ keepArchiveSelection: true });
    const successMessage = options.successMessage ?? "ФИО сотрудника обновлено.";
    if (options.self) {
      setMessage(successMessage, "success");
    } else {
      setUserFormMessage(successMessage, "success");
    }
    return updated;
  } catch (error) {
    if (options.self) {
      setMessage(error, "error");
    } else {
      setUserFormMessage(error, "error");
    }
    return null;
  }
}

async function saveOwnFullName() {
  const currentUser = appState.session?.user;
  if (!currentUser?.id) {
    return;
  }
  await updateUserFullName(currentUser.id, employeeFullNameInput?.value ?? "", {
    self: true,
    focusNode: employeeFullNameInput,
    emptyMessage: "Введите ФИО сотрудника для акта.",
    successMessage: "ФИО сотрудника сохранено.",
  });
}

async function downloadCurrentCalculationPDF() {

  const currentUser = appState.session?.user;

  const result = appState.currentCalculation;

  if (!currentUser?.id || !result) {

    setMessage("Сначала выполните точный расчёт, а затем скачайте акт.", "error");

    return;

  }



  const isExact = Boolean(result.foundExact ?? result.exactMatch ?? (result.totalAmount === result.targetAmount));
  if (!isExact) {
    setMessage("Скачать акт можно только для точного расчёта без отклонений.", "error");
    return;
  }

  const actNumber = clampPositiveInteger(actNumberVisibleInput?.value, appState.exportDraft.actNumber || ((currentUser.lastActNumber || 0) + 1));
  appState.exportDraft.actNumber = actNumber;
  const fullName = String(appState.exportDraft.employeeFullName || "").trim();
  const contractCode = String(appState.exportDraft.contractCode || "1").trim();
  const contractNumber = getSelectedContractNumber();
  const contractDate = String(appState.exportDraft.contractDate || "").trim();

  if (!fullName || !contractDate || !contractNumber) {
    setMessage("Сначала укажите договор, дату подписания и ФИО для формирования акта.", "error");
    openContractModal();
    return;
  }




  try {

    const path = await api.ExportCurrentCalculationPDF({

      actNumber,

      employeeFullName: fullName,

      contractCode,

      contractNumber,

      contractDate,

      generatedAt: result.generatedAt,

      targetAmount: result.targetAmount,

      items: result.items,

    });

    const updated = await api.UpdateUserLastActNumber({
      userID: currentUser.id,
      lastActNumber: actNumber,
    });
    if (updated && appState.session?.user?.id === updated.id) {
      appState.session.user = { ...appState.session.user, ...updated };
    } else if (appState.session?.user) {
      appState.session.user.lastActNumber = actNumber;
    }
    appState.exportDraft.actNumber = actNumber + 1;
    updateHiddenExportInputs();
    renderActNumberControls();
    setMessage(`Акт сохранён: ${path}`, "success");

  } catch (error) {

    setMessage(error, "error");

  }

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
    usersList.innerHTML = '<div class="settings-empty">\u0423\u043f\u0440\u0430\u0432\u043b\u0435\u043d\u0438\u0435 \u043f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u044f\u043c\u0438 \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u043e \u0441\u0442\u0430\u0440\u0448\u0438\u043c \u0438 \u0440\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044f\u043c \u0441\u0432\u043e\u0435\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430, \u0430 \u0442\u0430\u043a\u0436\u0435 \u0430\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440\u0443.</div>';
    setUserFormMessage("\u0421\u043e\u0437\u0434\u0430\u0432\u0430\u0442\u044c \u043f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u0435\u0439 \u043c\u043e\u0436\u043d\u043e \u0442\u043e\u043b\u044c\u043a\u043e \u0432\u043d\u0443\u0442\u0440\u0438 \u0441\u0432\u043e\u0435\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430 \u0438 \u0441\u0432\u043e\u0435\u0439 \u0437\u043e\u043d\u044b \u043e\u0442\u0432\u0435\u0442\u0441\u0442\u0432\u0435\u043d\u043d\u043e\u0441\u0442\u0438.", "muted");
    return;
  }
  if (!sortedUsers.length) {
    usersList.innerHTML = '<div class="settings-empty">\u041f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u0438 \u043f\u043e\u043a\u0430 \u043d\u0435 \u0441\u043e\u0437\u0434\u0430\u043d\u044b.</div>';
    setUserFormMessage("\u041c\u043e\u0436\u043d\u043e \u0441\u043e\u0437\u0434\u0430\u0442\u044c \u043f\u0435\u0440\u0432\u043e\u0433\u043e \u043f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u044f \u0432 \u0440\u0430\u043c\u043a\u0430\u0445 \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u043e\u0439 \u0432\u0430\u043c \u0440\u043e\u043b\u0438.", "muted");
    return;
  }

  usersList.innerHTML = sortedUsers.map((user) => {
    const canChangeRole = Boolean(appState.session?.canAdmin)
      || (appState.session.canManage && roleDepartment(viewerRole) === roleDepartment(user.role));
    const canDelete = canDeleteManagedUser(appState.session.user?.role, user.role);
    const canResetPassword = roleCanCreateUsers(appState.session.user?.role) && canDeleteManagedUser(appState.session.user?.role, user.role);
    const canEditFullName = Boolean(appState.session?.canAdmin) || canResetPassword || canDelete;
    const roleOptions = renderRoleOptions(roleChoices, user.role);

    return `
      <article class="settings-item glass user-item">
        <div>
          <strong>${escapeHtml(user.username)}</strong>
          <div class="service-meta">${roleLabel(user.role)}</div>
          <div class="service-meta">\u0421\u043e\u0437\u0434\u0430\u043d: ${formatDate(user.createdAt)}</div>
          ${canEditFullName ? `<div class="user-fullname-row"><input type="text" class="input-select compact-select inline-fullname-input" placeholder="\u0424\u0418\u041e \u0441\u043e\u0442\u0440\u0443\u0434\u043d\u0438\u043a\u0430" value="${escapeHtml(user.fullName || "")}" data-user-fullname-id="${user.id}" /><button type="button" class="ghost-button compact-action-button" data-save-fullname-id="${user.id}">\u0421\u043e\u0445\u0440\u0430\u043d\u0438\u0442\u044c \u0424\u0418\u041e</button></div>` : ""}
          ${canResetPassword ? `<div class="user-password-row"><input type="password" class="input-select compact-select inline-password-input" placeholder="\u041d\u043e\u0432\u044b\u0439 \u043f\u0430\u0440\u043e\u043b\u044c" data-user-password-id="${user.id}" /><button type="button" class="ghost-button compact-action-button" data-apply-password-id="${user.id}">\u0421\u043c\u0435\u043d\u0438\u0442\u044c \u043f\u0430\u0440\u043e\u043b\u044c</button></div>` : ""}
        </div>
        <div class="settings-item-actions stacked-actions">
          ${canChangeRole ? `<select class="input-select compact-select" data-user-role-id="${user.id}">${roleOptions}</select><button type="button" class="ghost-button" data-apply-role-id="${user.id}">\u0421\u043c\u0435\u043d\u0438\u0442\u044c \u0440\u043e\u043b\u044c</button>` : `<div class="service-meta">\u0421\u043c\u0435\u043d\u0430 \u0440\u043e\u043b\u0435\u0439 \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u0430 \u0442\u043e\u043b\u044c\u043a\u043e \u0440\u0443\u043a\u043e\u0432\u043e\u0434\u0438\u0442\u0435\u043b\u044e \u043e\u0442\u0434\u0435\u043b\u0430 \u0438 \u0430\u0434\u043c\u0438\u043d\u0438\u0441\u0442\u0440\u0430\u0442\u043e\u0440\u0443.</div>`}
          <button type="button" class="danger-button" ${canDelete ? `data-delete-user-id="${user.id}"` : "disabled"}>\u0423\u0434\u0430\u043b\u0438\u0442\u044c</button>
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

  usersList.querySelectorAll("[data-save-fullname-id]").forEach((button) => {
    button.addEventListener("click", async () => {
      const userId = Number(button.dataset.saveFullnameId);
      const fullNameInput = usersList.querySelector(`[data-user-fullname-id="${userId}"]`);
      await updateUserFullName(userId, fullNameInput?.value ?? "", {
        focusNode: fullNameInput,
        successMessage: "\u0424\u0418\u041e \u0441\u043e\u0442\u0440\u0443\u0434\u043d\u0438\u043a\u0430 \u0441\u043e\u0445\u0440\u0430\u043d\u0435\u043d\u043e.",
      });
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

  setUserFormMessage("\u041f\u043e\u043b\u044c\u0437\u043e\u0432\u0430\u0442\u0435\u043b\u0438 \u043e\u0442\u043e\u0431\u0440\u0430\u0436\u0430\u044e\u0442\u0441\u044f \u0438 \u0441\u043e\u0437\u0434\u0430\u044e\u0442\u0441\u044f \u0442\u043e\u043b\u044c\u043a\u043e \u0432 \u043f\u0440\u0435\u0434\u0435\u043b\u0430\u0445 \u0432\u0430\u0448\u0435\u0433\u043e \u043e\u0442\u0434\u0435\u043b\u0430 \u0438 \u0443\u0440\u043e\u0432\u043d\u044f \u0434\u043e\u0441\u0442\u0443\u043f\u0430.", "muted");
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
  const canEditArchive = Boolean(appState.session?.canAdmin && saved);
  const isEditingArchive = Boolean(saved && appState.editingArchiveId === saved.id);
  editArchiveButton?.classList.toggle("hidden", !canEditArchive || isEditingArchive);
  if (editArchiveButton) {
    editArchiveButton.disabled = !canEditArchive || isEditingArchive;
  }
  cancelArchiveEditButton?.classList.toggle("hidden", !isEditingArchive);
  saveArchiveEditButton?.classList.toggle("hidden", !isEditingArchive);
  if (cancelArchiveEditButton) {
    cancelArchiveEditButton.disabled = !isEditingArchive;
  }
  if (saveArchiveEditButton) {
    saveArchiveEditButton.disabled = !isEditingArchive;
  }
  deleteArchiveButton.classList.toggle("hidden", !canDeleteArchive);
  deleteArchiveButton.disabled = !canDeleteArchive || isEditingArchive;
  copyArchiveServicesButton.classList.toggle("hidden", !canCopyArchiveServices);
  copyArchiveServicesButton.disabled = !canCopyArchiveServices || isEditingArchive;
  if (!saved) {
    archiveTitle.textContent = appState.session?.canModerate ? "Выберите расчёт сотрудника" : "Выберите расчёт из архива";
    archiveTotal.textContent = "0 р";
    archiveMeta.textContent = "Здесь появятся дата, автор и состав выбранного расчёта.";
    archiveBody.innerHTML = '<tr><td colspan="7" class="placeholder">Архив пока пуст или расчёт ещё не выбран.</td></tr>';
    return;
  }

  const items = Array.isArray(saved.items) ? saved.items : [];
  archiveTitle.textContent = saved.title || "Архивный расчёт";
  archiveTotal.textContent = formatMoney(saved.totalAmount || 0);
  archiveMeta.innerHTML = `<span>Целевая сумма: <strong>${formatMoney(saved.targetAmount || 0)}</strong></span><span>Автор: <strong>${escapeHtml(saved.createdBy || "Не указан")}</strong></span><span class="history-date">${formatDate(saved.createdAt)}</span>${isEditingArchive ? '<span class="archive-inline-note">Режим редактирования архива активен.</span>' : ""}`;

  let effectivePercentMap = {};
  try {
    const weightMap = Object.fromEntries(items.map((item) => [item.serviceCode ?? item.code ?? "", item.weight ?? 0]));
    effectivePercentMap = buildEffectivePercentMap(items, weightMap);
  } catch (error) {
    console.error("Archive detail render failed", error, saved);
    setMessage("Не удалось корректно отобразить состав архива.", "error");
  }

  if (!items.length) {
    archiveBody.innerHTML = '<tr><td colspan="7" class="placeholder">В этом архиве пока нет услуг для отображения.</td></tr>';
    return;
  }

  archiveBody.innerHTML = items.map((item, index) => {
    const percentKey = item.serviceCode ?? item.code ?? "";
    const percentValue = effectivePercentMap[percentKey] ?? 0;
    const percentInputValue = item.allocationPercent == null ? "" : String(roundToOneDecimal(item.allocationPercent));
    if (isEditingArchive) {
      return `
      <tr class="archive-edit-row ${item.quantity === 0 ? "muted-row" : ""}" data-archive-row="${index}">
        <td>${index + 1}</td>
        <td>
          <div class="archive-edit-stack">
            <input class="archive-edit-input" data-archive-name value="${escapeAttribute(item.name ?? "")}" />
            <div class="archive-edit-meta">
              <input class="archive-edit-input" data-archive-rate type="number" min="1" step="1" value="${Number(item.rate) || 0}" />
              <select class="archive-edit-select" data-archive-unit>
                <option value="ч." ${item.unit === "ч." ? "selected" : ""}>ч.</option>
                <option value="шт." ${item.unit === "шт." ? "selected" : ""}>шт.</option>
              </select>
            </div>
          </div>
        </td>
        <td class="group-cell">
          <select class="archive-edit-select" data-archive-category>
            <option value="primary" ${item.category === "primary" ? "selected" : ""}>Основная</option>
            <option value="secondary" ${item.category === "secondary" ? "selected" : ""}>Вторичная</option>
            <option value="closing" ${item.category === "closing" ? "selected" : ""}>Закрывающая</option>
          </select>
        </td>
        <td class="weight-value"><input class="archive-edit-input" data-archive-percent type="number" min="0" step="0.1" value="${escapeAttribute(percentInputValue)}" placeholder="${formatPercent(percentValue)}" /></td>
        <td class="weight-value"><input class="archive-edit-input" data-archive-weight type="number" min="-10" max="10" step="1" value="${item.weight ?? 0}" /></td>
        <td class="qty"><input class="archive-edit-input" data-archive-quantity type="number" min="0" step="1" value="${Number(item.quantity) || 0}" /></td>
        <td class="money">${formatMoney(Number(item.lineTotal) || 0)}</td>
      </tr>`;
    }
    return `
      <tr class="${item.quantity === 0 ? "muted-row" : ""}">
        <td>${index + 1}</td>
        <td>
          <strong class="truncate-text" title="${escapeHtml(item.name ?? "")}">${escapeHtml(item.name ?? "")}</strong>
          <div class="service-meta">${formatMoney(Number(item.rate) || 0)} / ${escapeHtml(item.unit ?? "")}</div>
        </td>
        <td class="group-cell">${categoryLabel(item.category)}</td>
        <td class="weight-value">${formatPercent(percentValue)}</td>
        <td class="weight-value">${item.weight ?? 0}</td>
        <td class="qty">${Number(item.quantity) || 0} ${escapeHtml(item.unit ?? "")}</td>
        <td class="money">${formatMoney(Number(item.lineTotal) || 0)}</td>
      </tr>`;
  }).join("");
}

function roundToOneDecimal(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return 0;
  }
  return Math.round(numeric * 10) / 10;
}

function getActiveArchiveCalculation() {
  return appState.savedCalculations.find((item) => item.id === appState.activeHistoryId) ?? null;
}

function startArchiveEdit() {
  const current = getActiveArchiveCalculation();
  if (!appState.session?.canAdmin || !current) {
    return;
  }
  appState.editingArchiveId = current.id;
  renderArchiveDetails(current);
  setMessage("Архив переведён в режим редактирования.", "muted");
}

function cancelArchiveEdit() {
  appState.editingArchiveId = null;
  renderArchiveDetails(getActiveArchiveCalculation());
  setMessage("Редактирование архива отменено.", "muted");
}

function collectArchiveEditRequest(saved) {
  const rows = [...archiveBody.querySelectorAll("[data-archive-row]")];
  return {
    id: saved.id,
    items: rows.map((row, index) => {
      const source = saved.items[index] || {};
      const name = row.querySelector("[data-archive-name]")?.value?.trim() || "";
      const rate = Number.parseInt(row.querySelector("[data-archive-rate]")?.value || String(source.rate || 0), 10) || 0;
      const unit = row.querySelector("[data-archive-unit]")?.value || source.unit || "ч.";
      const category = row.querySelector("[data-archive-category]")?.value || source.category || "primary";
      const quantity = Number.parseInt(row.querySelector("[data-archive-quantity]")?.value || String(source.quantity || 0), 10) || 0;
      const weight = Number.parseInt(row.querySelector("[data-archive-weight]")?.value || String(source.weight || 0), 10) || 0;
      const percentRaw = row.querySelector("[data-archive-percent]")?.value?.trim() || "";
      return {
        serviceId: source.serviceId || 0,
        serviceCode: source.serviceCode || `archive-${saved.id}-${index + 1}`,
        name,
        unit,
        rate,
        quantity,
        lineTotal: rate * quantity,
        description: source.description || "",
        weight,
        category,
        allocationPercent: percentRaw === "" ? null : Number.parseFloat(percentRaw),
      };
    }),
  };
}

async function saveArchiveEdit() {
  const current = getActiveArchiveCalculation();
  if (!appState.session?.canAdmin || !current) {
    return;
  }
  try {
    const updated = await api.UpdateCalculationAsAdmin(collectArchiveEditRequest(current));
    appState.savedCalculations = appState.savedCalculations.map((item) => item.id === updated.id ? updated : item);
    appState.editingArchiveId = null;
    renderHistory();
    renderArchiveDetails(updated);
    setMessage("Архивный расчёт сохранён.", "success");
  } catch (error) {
    setMessage(normalizeErrorText(error, "Не удалось сохранить архивный расчёт."), "error");
  }
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
    next[service.code] = 0;
    if (!serviceTakesPartInCalculation(service)) {
      return;
    }
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

    try {
      const saved = await api.SaveCalculation({
        targetAmount: result.targetAmount,
        items: result.items,
      });
      appState.activeHistoryId = saved.id ?? null;
      await refreshBootstrap({ keepArchiveSelection: true });
      setMessage(`Расчёт сохранён в архив за ${saved.title}. Активно ${result.activeServices} из ${result.items.length} услуг на сумму ${formatMoney(result.totalAmount)}.`, "success");
    } catch (error) {
      setMessage(`Расчёт построен, но не удалось сохранить его в архив: ${normalizeErrorText(error, "ошибка сохранения архива")}.`, "error");
    }
  } catch (error) {
    appState.currentCalculation = null;
    renderResult(null);
    setMessage(error, "error");
  } finally {
    calculateButton.disabled = false;
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
  if (!await requestDeleteConfirmation("услугу", service.name)) {
    return;
  }
  try {
    await api.DeleteService(id);
    if (appState.editingServiceId === id) {
      resetServiceForm();
    }
    delete appState.weights[service.code];
    await refreshBootstrap();
    setServiceFormMessage(`Услуга ${service.name} удалена.`, "success");
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
  if (!await requestDeleteConfirmation("пользователя", user.username)) {
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
    setMessage("Введите сумму и при необходимости сместите акценты весами по услугам.", "muted");
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
contractDetailsButton?.addEventListener("click", openContractModal);
saveOwnFullNameButton?.addEventListener("click", saveOwnFullName);
actNumberVisibleInput?.addEventListener("change", persistActNumber);
actNumberVisibleInput?.addEventListener("blur", persistActNumber);
downloadPdfButton?.addEventListener("click", downloadCurrentCalculationPDF);
randomizeButton?.addEventListener("click", randomizeCalculation);
resetWeightsButton.addEventListener("click", resetWeights);
archiveOwnerFilter?.addEventListener("change", () => {
  appState.archiveOwnerFilter = archiveOwnerFilter.value || "all";
  appState.activeHistoryId = null;
  appState.editingArchiveId = null;
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
editArchiveButton?.addEventListener("click", startArchiveEdit);
cancelArchiveEditButton?.addEventListener("click", cancelArchiveEdit);
saveArchiveEditButton?.addEventListener("click", saveArchiveEdit);
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
contractCancel?.addEventListener("click", () => {
  closeContractModal();
});
contractSave?.addEventListener("click", () => {
  saveContractDetails();
});
contractBackdrop?.addEventListener("click", () => {
  closeContractModal();
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && !confirmModal.classList.contains("hidden")) {
    closeConfirmModal(false);
    return;
  }
  if (event.key === "Escape" && contractModal && !contractModal.classList.contains("hidden")) {
    closeContractModal();
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


























