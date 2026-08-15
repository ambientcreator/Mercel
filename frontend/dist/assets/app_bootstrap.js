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

        title: getSelectedCalculationDateTitle(),

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

function formatTypedDateValue(value) {
  const digits = String(value || "").replace(/\D/g, "").slice(0, 8);
  const parts = [];

  if (digits.length > 0) {
    parts.push(digits.slice(0, 2));
  }
  if (digits.length > 2) {
    parts.push(digits.slice(2, 4));
  }
  if (digits.length > 4) {
    parts.push(digits.slice(4, 8));
  }

  return parts.join(".");
}


let datePopover = null;
let activeDatePopover = null;
let datePopoverView = null;

function ensureDatePopover() {
  if (datePopover) {
    return datePopover;
  }

  datePopover = document.createElement("div");
  datePopover.className = "app-date-popover hidden";
  datePopover.innerHTML = `
    <div class="app-date-popover-header">
      <button type="button" class="app-date-nav" data-action="prev" aria-label="\u041f\u0440\u0435\u0434\u044b\u0434\u0443\u0449\u0438\u0439 \u043c\u0435\u0441\u044f\u0446">&#x2039;</button>
      <div class="app-date-title"></div>
      <button type="button" class="app-date-nav" data-action="next" aria-label="\u0421\u043b\u0435\u0434\u0443\u044e\u0449\u0438\u0439 \u043c\u0435\u0441\u044f\u0446">&#x203a;</button>
    </div>
    <div class="app-date-weekdays"></div>
    <div class="app-date-grid"></div>
    <div class="app-date-footer">
      <button type="button" class="app-date-link" data-action="today">\u0421\u0435\u0433\u043e\u0434\u043d\u044f</button>
      <button type="button" class="app-date-link" data-action="close">\u0417\u0430\u043a\u0440\u044b\u0442\u044c</button>
    </div>
  `;

  document.body.appendChild(datePopover);

  const weekdays = ["\u041f\u043d", "\u0412\u0442", "\u0421\u0440", "\u0427\u0442", "\u041f\u0442", "\u0421\u0431", "\u0412\u0441"];
  const weekdaysNode = datePopover.querySelector(".app-date-weekdays");
  weekdaysNode.innerHTML = weekdays.map((day) => `<span>${day}</span>`).join("");

  datePopover.addEventListener("click", (event) => {
    const button = event.target.closest("button[data-action], button[data-date]");
    if (!button || !activeDatePopover) {
      return;
    }

    const action = button.dataset.action || "";
    const dateValue = button.dataset.date || "";

    if (dateValue) {
      applyDatePopoverValue(dateValue);
      return;
    }

    if (action === "prev") {
      datePopoverView = addMonths(datePopoverView, -1);
      renderDatePopover();
      return;
    }

    if (action === "next") {
      const nextMonth = addMonths(datePopoverView, 1);
      if (!isMonthAfterCurrentYear(nextMonth)) {
        datePopoverView = nextMonth;
        renderDatePopover();
      }
      return;
    }

    if (action === "today") {
      const today = new Date();
      applyDatePopoverValue(formatDateInputValue(today));
      return;
    }

    if (action === "close") {
      closeDatePopover();
    }
  });

  document.addEventListener("mousedown", (event) => {
    if (!activeDatePopover || !datePopover || datePopover.classList.contains("hidden")) {
      return;
    }
    if (datePopover.contains(event.target)) {
      return;
    }
    if (activeDatePopover.wrap?.contains(event.target)) {
      return;
    }
    closeDatePopover();
  });

  window.addEventListener("resize", () => {
    if (activeDatePopover) {
      positionDatePopover();
    }
  });

  document.addEventListener("scroll", () => {
    if (activeDatePopover) {
      positionDatePopover();
    }
  }, true);

  return datePopover;
}

function closeDatePopover() {
  if (!datePopover) {
    return;
  }

  datePopover.classList.add("hidden");
  datePopover.classList.remove("is-above");
  activeDatePopover = null;
  datePopoverView = null;
}

function startOfMonth(date) {
  return new Date(date.getFullYear(), date.getMonth(), 1, 12, 0, 0, 0);
}

function addMonths(date, delta) {
  return new Date(date.getFullYear(), date.getMonth() + delta, 1, 12, 0, 0, 0);
}

function getInitialPopoverDate(value) {
  const parsed = parseDateInputValue(value);
  return parsed || new Date();
}

function isSameDay(left, right) {
  return left.getFullYear() === right.getFullYear() && left.getMonth() === right.getMonth() && left.getDate() === right.getDate();
}

function isMonthAfterCurrentYear(date) {
  const currentYear = new Date().getFullYear();
  return date.getFullYear() > currentYear;
}

function getMonthTitle(date) {
  const title = new Intl.DateTimeFormat("ru-RU", { month: "long", year: "numeric" }).format(date);
  return title.charAt(0).toUpperCase() + title.slice(1);
}

function renderDatePopover() {
  if (!datePopover || !activeDatePopover || !datePopoverView) {
    return;
  }

  const titleNode = datePopover.querySelector(".app-date-title");
  const gridNode = datePopover.querySelector(".app-date-grid");
  const nextButton = datePopover.querySelector('[data-action="next"]');
  const selectedDate = getInitialPopoverDate(activeDatePopover.visibleInput?.value || activeDatePopover.pickerInput?.value || "");
  const today = new Date();
  const viewDate = startOfMonth(datePopoverView);
  const startWeekday = (viewDate.getDay() + 6) % 7;
  const daysInMonth = new Date(viewDate.getFullYear(), viewDate.getMonth() + 1, 0).getDate();
  const dayCells = [];

  titleNode.textContent = getMonthTitle(viewDate);

  for (let index = 0; index < startWeekday; index += 1) {
    dayCells.push('<span class="app-date-cell is-empty"></span>');
  }

  for (let day = 1; day <= daysInMonth; day += 1) {
    const current = new Date(viewDate.getFullYear(), viewDate.getMonth(), day, 12, 0, 0, 0);
    const iso = formatDateInputValue(current);
    const classes = ["app-date-cell", "app-date-day"];

    if (isSameDay(current, selectedDate)) {
      classes.push("is-selected");
    }

    if (isSameDay(current, today)) {
      classes.push("is-today");
    }

    dayCells.push(`<button type="button" class="${classes.join(" ")}" data-date="${iso}">${day}</button>`);
  }

  gridNode.innerHTML = dayCells.join("");
  nextButton.disabled = isMonthAfterCurrentYear(addMonths(viewDate, 1));

  requestAnimationFrame(positionDatePopover);
}

function positionDatePopover() {
  if (!datePopover || !activeDatePopover) {
    return;
  }

  const margin = 8;
  const wrapRect = activeDatePopover.wrap.getBoundingClientRect();
  const popRect = datePopover.getBoundingClientRect();
  const belowTop = wrapRect.bottom + 6;
  const aboveTop = wrapRect.top - popRect.height - 6;
  let top = belowTop;
  let left = wrapRect.left;

  if (belowTop + popRect.height > window.innerHeight - margin && aboveTop >= margin) {
    top = aboveTop;
    datePopover.classList.add("is-above");
  } else {
    datePopover.classList.remove("is-above");
    if (top + popRect.height > window.innerHeight - margin) {
      top = Math.max(margin, window.innerHeight - popRect.height - margin);
    }
  }

  if (left + popRect.width > window.innerWidth - margin) {
    left = window.innerWidth - popRect.width - margin;
  }
  if (left < margin) {
    left = margin;
  }

  // The font-scale preference sets `zoom` on the root element. getBoundingClientRect and
  // window.innerWidth/Height report in the zoomed space, but the popover is a child of the
  // zoomed root, so its left/top are interpreted before scaling — divide to bring them back.
  // At the default scale the divisor is 1 and this is a no-op.
  const rootZoom = Number(getComputedStyle(document.documentElement).zoom) || 1;

  datePopover.style.left = `${left / rootZoom}px`;
  datePopover.style.top = `${top / rootZoom}px`;
}

function applyDatePopoverValue(isoValue) {
  if (!activeDatePopover) {
    return;
  }

  const { visibleInput, pickerInput } = activeDatePopover;
  const clamped = clampDateValueToCurrentYear(isoValue);
  syncDateControls(visibleInput, pickerInput, clamped);
  visibleInput?.dispatchEvent(new Event("change", { bubbles: true }));
  visibleInput?.dispatchEvent(new Event("blur", { bubbles: true }));
  closeDatePopover();
}

function openCustomDatePopover(visibleInput, pickerInput, pickerButton) {
  if (!visibleInput) {
    return;
  }

  const wrap = visibleInput.closest(".date-input-wrap");
  if (!wrap) {
    return;
  }

  ensureDatePopover();

  if (activeDatePopover && activeDatePopover.visibleInput === visibleInput && !datePopover.classList.contains("hidden")) {
    closeDatePopover();
    return;
  }

  activeDatePopover = { visibleInput, pickerInput, pickerButton, wrap };
  datePopoverView = startOfMonth(getInitialPopoverDate(visibleInput.value || pickerInput?.value || ""));
  datePopover.classList.remove("hidden");
  renderDatePopover();
}

function wireDatePicker(visibleInput, pickerInput, pickerButton) {
  if (!visibleInput || visibleInput.dataset.datePickerWired === "1") {
    return;
  }

  visibleInput.dataset.datePickerWired = "1";

  const syncVisibleValue = () => {
    const isoValue = clampDateValueToCurrentYear(visibleInput.value);
    syncDateControls(visibleInput, pickerInput, isoValue);
  };

  const openPopover = (event) => {
    event?.preventDefault?.();
    event?.stopPropagation?.();
    openCustomDatePopover(visibleInput, pickerInput, pickerButton);
  };

  visibleInput.addEventListener("input", () => {
    const formatted = formatTypedDateValue(visibleInput.value);
    if (visibleInput.value !== formatted) {
      visibleInput.value = formatted;
    }
  });

  visibleInput.addEventListener("change", syncVisibleValue);
  visibleInput.addEventListener("blur", syncVisibleValue);
  pickerButton?.addEventListener("click", openPopover);
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

    if (archiveDateInput) {

      archiveDateInput.value = "";

    }

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

wireDatePicker(archiveDateInput, archiveDatePicker, archiveDatePickerButton);
wireDatePicker(contractModalDateInput, contractModalDatePicker, contractModalDatePickerButton);

archiveDateInput?.addEventListener("change", persistArchiveDateInput);

actNumberVisibleInput?.addEventListener("change", persistActNumber);

archiveDateInput?.addEventListener("blur", persistArchiveDateInput);

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

contractSave?.addEventListener("click", () => {

  saveContractDetails();

});

contractCancel?.addEventListener("click", () => {

  closeContractModal();

});

contractBackdrop?.addEventListener("click", () => {

  closeContractModal();

});

// The status line on the calculator is the entry point into the contract form now,
// so the whole strip is clickable rather than carrying a separate button.
contractCard?.addEventListener("click", () => {

  openContractModal();

});

contractCard?.addEventListener("keydown", (event) => {

  if (event.key === "Enter" || event.key === " ") {

    event.preventDefault();

    openContractModal();

  }

});

settingsButton?.addEventListener("click", () => {

  openSettingsModal("appearance");

});

settingsCloseButton?.addEventListener("click", () => {

  closeSettingsModal();

});

settingsBackdrop?.addEventListener("click", () => {

  closeSettingsModal();

});

settingsNav?.addEventListener("click", (event) => {

  const button = event.target.closest("[data-settings-section]");

  if (button) {

    setSettingsSection(button.dataset.settingsSection);

  }

});

settingsThemeSelect?.addEventListener("change", () => {

  setPreference("theme", settingsThemeSelect.value);

});

settingsDensitySelect?.addEventListener("change", () => {

  setPreference("density", settingsDensitySelect.value);

});

settingsFontScaleSelect?.addEventListener("change", () => {

  setPreference("fontScale", settingsFontScaleSelect.value);

});

document.addEventListener("keydown", (event) => {

  if (event.key === "Escape" && activeDatePopover) {

    closeDatePopover();

    return;

  }

  if (event.key === "Escape" && !confirmModal.classList.contains("hidden")) {

    closeConfirmModal(false);

    return;

  }

  if (event.key === "Escape" && contractModal && !contractModal.classList.contains("hidden")) {

    closeContractModal();

    return;

  }

  if (event.key === "Escape" && settingsModal && !settingsModal.classList.contains("hidden")) {

    closeSettingsModal();

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

