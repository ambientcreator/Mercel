const appState = {

  session: { authenticated: false, canManage: false, canAdmin: false, canModerate: false, canCopyArchiveServices: false, user: null },

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

  actTemplates: [],

  exportDraft: {

    actNumber: 1,

    employeeFullName: "",

    contractCode: "1",

    contractSpbksNumber: "",

    contractGrizablNumber: "",

    contractDate: "",

  },

  preferences: { theme: "dark", density: "comfortable", fontScale: "medium" },

};

const PREFERENCES_STORAGE_KEY = "mercel.preferences";

const PREFERENCE_CHOICES = {

  theme: ["dark", "light"],

  density: ["comfortable", "compact"],

  fontScale: ["small", "medium", "large"],

};

const PREFERENCE_DEFAULTS = { theme: "dark", density: "comfortable", fontScale: "medium" };

const FONT_SCALE_ZOOM = { small: "0.92", medium: "1", large: "1.1" };

// Preferences are read and applied here, in the first script, so the shell paints
// with the chosen theme instead of flashing the dark default through bootstrap.
function readStoredPreferences() {

  const stored = { ...PREFERENCE_DEFAULTS };

  let raw = null;

  try {

    raw = window.localStorage?.getItem(PREFERENCES_STORAGE_KEY) ?? null;

  } catch (error) {

    return stored;

  }

  if (!raw) {

    return stored;

  }

  let parsed = null;

  try {

    parsed = JSON.parse(raw);

  } catch (error) {

    return stored;

  }

  if (!parsed || typeof parsed !== "object") {

    return stored;

  }

  Object.keys(PREFERENCE_CHOICES).forEach((key) => {

    if (PREFERENCE_CHOICES[key].includes(parsed[key])) {

      stored[key] = parsed[key];

    }

  });

  return stored;

}

function applyPreferences(preferences) {

  const root = document.documentElement;

  root.dataset.theme = preferences.theme;

  root.dataset.density = preferences.density;

  root.dataset.fontScale = preferences.fontScale;

  root.style.zoom = FONT_SCALE_ZOOM[preferences.fontScale] ?? "1";

}

function savePreferences() {

  try {

    window.localStorage?.setItem(PREFERENCES_STORAGE_KEY, JSON.stringify(appState.preferences));

  } catch (error) {

    // Storage can be unavailable in a locked-down webview; the in-memory value still applies.

  }

}

appState.preferences = readStoredPreferences();

applyPreferences(appState.preferences);

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
const archiveDatePickerButton = document.getElementById("archive-date-picker-button");
const archiveDatePicker = document.getElementById("archive-date-picker");

const actNumberInput = document.getElementById("act-number-input");

const employeeFullNameInput = document.getElementById("employee-full-name-input");

const contractCodeInput = document.getElementById("contract-code-input");

const contractSPBKSNumberInput = document.getElementById("contract-spbks-number-input");

const contractGrizablNumberInput = document.getElementById("contract-grizabl-number-input");

const contractDateInput = document.getElementById("contract-date-input");

const contractDetailsButton = document.getElementById("contract-details-button");

const contractDetailsSummary = document.getElementById("contract-details-summary");

const contractCard = document.getElementById("contract-card");

const contractCardAction = contractCard?.querySelector(".contract-card-action") ?? null;

const contractCardIcon = document.getElementById("contract-card-icon");

const saveOwnFullNameButton = document.getElementById("save-own-fullname-button");

const actTemplatePanel = document.getElementById("act-template-panel");

const settingsBlankTab = document.getElementById("settings-blank-tab");

const actTemplateSelect = document.getElementById("act-template-select");

const saveActTemplateButton = document.getElementById("save-act-template-button");

const actTemplateMessage = document.getElementById("act-template-message");

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

const statusStrip = document.getElementById("status-strip");

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

const contractMessage = document.getElementById("contract-message");

const contractSave = document.getElementById("contract-save");

const contractModalSPBKSNumberInput = document.getElementById("contract-modal-spbks-number-input");

const contractModalGrizablNumberInput = document.getElementById("contract-modal-grizabl-number-input");

const contractModalCodeSPBKS = document.getElementById("contract-modal-code-spbks");

const contractModalCodeGrizabl = document.getElementById("contract-modal-code-grizabl");

const contractModalDateInput = document.getElementById("contract-modal-date-input");
const contractModalDatePickerButton = document.getElementById("contract-modal-date-picker-button");
const contractModalDatePicker = document.getElementById("contract-modal-date-picker");

const contractModalFullNameInput = document.getElementById("contract-modal-fullname-input");

const settingsButton = document.getElementById("settings-button");

const settingsModal = document.getElementById("settings-modal");

const settingsBackdrop = document.getElementById("settings-backdrop");

const settingsCloseButton = document.getElementById("settings-close");

const settingsNav = document.getElementById("settings-nav");

const settingsMessage = document.getElementById("settings-message");

const settingsThemeSelect = document.getElementById("settings-theme-select");

const settingsDensitySelect = document.getElementById("settings-density-select");

const settingsFontScaleSelect = document.getElementById("settings-font-scale-select");

const settingsAppVersion = document.getElementById("settings-app-version");

const settingsDbPath = document.getElementById("settings-db-path");

const settingsSupport = document.getElementById("settings-support");

const settingsSectionButtons = document.querySelectorAll("[data-settings-section]");

const settingsSectionPanels = document.querySelectorAll("[data-settings-panel]");

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

function formatDateInputValue(date = new Date()) {

  if (Number.isNaN(date.getTime())) {

    return "";

  }

  const year = String(date.getFullYear());
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;

}

function parseDateInputValue(value) {

  const rawValue = String(value || "").trim();

  if (!rawValue) {
    return null;
  }

  let match = rawValue.match(/^(\d{2})\.(\d{2})\.(\d{4})$/);
  if (match) {
    const [, day, month, year] = match;
    const parsed = new Date(`${year}-${month}-${day}T12:00:00`);
    if (!Number.isNaN(parsed.getTime()) && parsed.getDate() === Number(day) && parsed.getMonth() + 1 === Number(month) && parsed.getFullYear() === Number(year)) {
      return parsed;
    }
    return null;
  }

  match = rawValue.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (match) {
    const [, year, month, day] = match;
    const parsed = new Date(`${year}-${month}-${day}T12:00:00`);
    if (!Number.isNaN(parsed.getTime()) && parsed.getDate() === Number(day) && parsed.getMonth() + 1 === Number(month) && parsed.getFullYear() === Number(year)) {
      return parsed;
    }
  }

  return null;

}

function formatDateDisplayValue(value) {

  const parsed = parseDateInputValue(value);

  if (!parsed) {
    return "";
  }

  const day = String(parsed.getDate()).padStart(2, "0");
  const month = String(parsed.getMonth() + 1).padStart(2, "0");
  const year = String(parsed.getFullYear());
  return `${day}.${month}.${year}`;

}

function getCurrentYearMaxDateValue() {

  return `${new Date().getFullYear()}-12-31`;

}

function clampDateValueToCurrentYear(value) {

  const parsed = parseDateInputValue(value);

  if (!parsed) {
    return "";
  }

  const maxValue = getCurrentYearMaxDateValue();
  const maxDate = new Date(`${maxValue}T12:00:00`);

  if (parsed.getFullYear() > new Date().getFullYear() || parsed > maxDate) {
    return maxValue;
  }

  return formatDateInputValue(parsed);

}

function syncDateControls(visibleInput, pickerInput, isoValue) {

  if (visibleInput) {
    visibleInput.value = formatDateDisplayValue(isoValue);
  }

  if (pickerInput) {
    pickerInput.max = getCurrentYearMaxDateValue();
    pickerInput.value = isoValue || "";
  }

}

function getSelectedCalculationDate() {

  const rawValue = clampDateValueToCurrentYear(archiveDateInput?.value || "");
  const parsed = rawValue ? new Date(`${rawValue}T12:00:00`) : new Date();

  return Number.isNaN(parsed.getTime()) ? new Date() : parsed;

}

function getSelectedCalculationDateTitle() {

  return formatArchiveTitle(getSelectedCalculationDate());

}

function getSelectedCalculationDateRFC3339() {

  return getSelectedCalculationDate().toISOString();

}

function syncArchiveDateInput() {

  if (archiveDateInput && !String(archiveDateInput.value || "").trim()) {
    const nextValue = clampDateValueToCurrentYear(formatDateInputValue(new Date()));
    syncDateControls(archiveDateInput, archiveDatePicker, nextValue);
  }

}

function persistArchiveDateInput() {

  if (!archiveDateInput) {
    return;
  }

  const nextValue = clampDateValueToCurrentYear(archiveDateInput.value || formatDateInputValue(getSelectedCalculationDate()));
  syncDateControls(archiveDateInput, archiveDatePicker, nextValue);

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

function contractTemplateShortLabel(code) {

  return code === "2" ? "\u0413\u0440\u0438\u0437\u0430\u0431\u043b\u044c" : "\u0421\u041f\u0431\u041a\u0421";

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

  // The card carries the state; the value node only carries the text, so assigning
  // textContent here cannot wipe the label and the action hint next to it.
  const complete = Boolean(selectedNumber && contractDate && fullName);

  contractCard?.classList.toggle("is-complete", complete);

  if (!complete) {

    contractDetailsSummary.textContent = "\u041d\u0435 \u0437\u0430\u043f\u043e\u043b\u043d\u0435\u043d \u2014 \u043d\u0430\u0436\u043c\u0438\u0442\u0435, \u0447\u0442\u043e\u0431\u044b \u0443\u043a\u0430\u0437\u0430\u0442\u044c";

    if (contractCard) {

      contractCard.title = "\u0414\u043b\u044f \u0432\u044b\u0433\u0440\u0443\u0437\u043a\u0438 \u0430\u043a\u0442\u0430 \u0443\u043a\u0430\u0436\u0438\u0442\u0435 \u043d\u043e\u043c\u0435\u0440 \u0434\u043e\u0433\u043e\u0432\u043e\u0440\u0430, \u0434\u0430\u0442\u0443 \u043f\u043e\u0434\u043f\u0438\u0441\u0430\u043d\u0438\u044f \u0438 \u0432\u0430\u0448\u0435 \u0424\u0418\u041e.";

    }

    if (contractCardAction) {

      contractCardAction.textContent = "\u0417\u0430\u043f\u043e\u043b\u043d\u0438\u0442\u044c";

    }

    if (contractCardIcon) {

      contractCardIcon.textContent = "!";

    }

    return;

  }

  contractDetailsSummary.textContent = `${contractTemplateShortLabel(appState.exportDraft.contractCode)} \u2116${selectedNumber} \u043e\u0442 ${formatArchiveTitle(new Date(contractDate))} \u00b7 ${fullName}`;

  if (contractCard) {

    contractCard.title = `\u0412\u044b\u0431\u0440\u0430\u043d \u0434\u043e\u0433\u043e\u0432\u043e\u0440: ${contractTemplateLabel(appState.exportDraft.contractCode)} \u2116${selectedNumber}. \u0424\u0418\u041e: ${fullName}.`;

  }

  if (contractCardAction) {

    contractCardAction.textContent = "\u0418\u0437\u043c\u0435\u043d\u0438\u0442\u044c";

  }

  if (contractCardIcon) {

    contractCardIcon.textContent = "\u2713";

  }

}

