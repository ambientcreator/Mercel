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

    const nextValue = clampDateValueToCurrentYear(appState.exportDraft.contractDate || new Date().toISOString().slice(0, 10));

    syncDateControls(contractModalDateInput, contractModalDatePicker, nextValue);

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

  const contractDate = clampDateValueToCurrentYear(String(contractModalDateInput?.value ?? "").trim());

  syncDateControls(contractModalDateInput, contractModalDatePicker, contractDate);

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

