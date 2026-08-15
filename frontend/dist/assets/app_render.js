// #message is a span inside the single status strip now, so the state is expressed by a
// modifier on the strip rather than by rewriting the span's own class.
function setMessage(text, type = "muted") {

  messageNode.textContent = normalizeErrorText(text, "\u041f\u0440\u043e\u0438\u0437\u043e\u0448\u043b\u0430 \u043e\u0448\u0438\u0431\u043a\u0430.");

  if (!statusStrip) {

    return;

  }

  statusStrip.classList.toggle("is-ready", type === "success");

  statusStrip.classList.toggle("is-error", type === "error");

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

function setSettingsMessage(text, type = "muted") {

  if (!settingsMessage) {

    return;

  }

  settingsMessage.textContent = normalizeErrorText(text, "Не удалось сохранить настройки.");

  settingsMessage.className = `message ${type}`;

}

function setContractMessage(text, type = "muted") {

  if (!contractMessage) {

    return;

  }

  contractMessage.textContent = normalizeErrorText(text, "Не удалось сохранить данные договора.");

  contractMessage.className = `message ${type}`;

}

function renderSettingsPreferences() {

  if (settingsThemeSelect) {

    settingsThemeSelect.value = appState.preferences.theme;

  }

  if (settingsDensitySelect) {

    settingsDensitySelect.value = appState.preferences.density;

  }

  if (settingsFontScaleSelect) {

    settingsFontScaleSelect.value = appState.preferences.fontScale;

  }

}

function renderContractTemplateOptions() {

  if (!settingsBlankSelect) {

    return;

  }

  const current = String(appState.session?.user?.preferredContractCode || appState.exportDraft.contractCode || "1");

  settingsBlankSelect.innerHTML = appState.contractTemplates

    .map((template) => `<option value="${escapeAttribute(template.code)}"${template.code === current ? " selected" : ""}>${escapeHtml(template.title)}</option>`)

    .join("");

  settingsBlankSelect.value = current;

  const active = appState.contractTemplates.find((template) => template.code === current);

  if (settingsBlankDetails) {

    settingsBlankDetails.textContent = active

      ? `Заказчик: ${active.customerName}. Подписант: ${active.directorShort}.`

      : "Список бланков недоступен.";

  }

}

function setPreference(key, value) {

  if (!PREFERENCE_CHOICES[key]?.includes(value)) {

    renderSettingsPreferences();

    return;

  }

  appState.preferences[key] = value;

  applyPreferences(appState.preferences);

  savePreferences();

  setSettingsMessage("Оформление обновлено.", "success");

}

function setActiveTab(tab) {

  // A stale activeTab can survive a logout, so never land on a tab the current session cannot see.
  if (tab === "settings" && settingsTabButton?.classList.contains("hidden")) {

    tab = "calculator";

  }

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

      contractCode: String(appState.session?.user?.preferredContractCode ?? "1").trim() || "1",

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

  // Every user keeps their own service list, so the management tab is for everyone who is
  // logged in — but it must hide again on logout, hence toggle rather than remove.
  settingsTabButton.classList.toggle("hidden", !authenticated);

  if (!authenticated) {

    closeSettingsModal();

    closeContractModal();

    return;

  }

  sessionUsername.textContent = appState.session.user?.username ?? "Гость";

  sessionRole.textContent = roleLabel(appState.session.user?.role);

  usersPanel.classList.toggle("hidden", !appState.session.canModerate);

  settingsBlankTab?.classList.toggle("hidden", !appState.session.canAdmin);

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

function currentPercentMeta(service, percentMap) {

  return `${categoryLabel(service.category)}, текущий процент ${escapeHtml(percentText(service, percentMap))}`;

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

    resultBody.innerHTML = '<tr><td colspan="4" class="placeholder">\u0412\u0432\u0435\u0434\u0438\u0442\u0435 \u0441\u0443\u043c\u043c\u0443 \u0441\u043b\u0435\u0432\u0430 \u0438 \u0437\u0430\u0434\u0430\u0439\u0442\u0435 \u0432\u0435\u0441\u0430 \u2014 \u0440\u0430\u0441\u0447\u0451\u0442 \u0432\u044b\u043f\u043e\u043b\u043d\u0438\u0442\u0441\u044f \u0430\u0432\u0442\u043e\u043c\u0430\u0442\u0438\u0447\u0435\u0441\u043a\u0438. \u041a\u043d\u043e\u043f\u043a\u0430 \u00ab\u0421\u043a\u0430\u0447\u0430\u0442\u044c \u0430\u043a\u0442 PDF\u00bb \u0441\u0442\u0430\u043d\u0435\u0442 \u0430\u043a\u0442\u0438\u0432\u043d\u043e\u0439 \u043f\u0440\u0438 \u0442\u043e\u0447\u043d\u043e\u043c \u0441\u043e\u0432\u043f\u0430\u0434\u0435\u043d\u0438\u0438 \u0441\u0443\u043c\u043c\u044b.</td></tr>';

    return;

  }

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

    servicesList.innerHTML = `

      <div class="empty-state">

        <div class="empty-state-title">\u0423\u0441\u043b\u0443\u0433\u0438 \u0435\u0449\u0451 \u043d\u0435 \u0441\u043e\u0437\u0434\u0430\u043d\u044b</div>

        <p class="empty-state-text">\u0420\u0430\u0441\u0447\u0451\u0442 \u0441\u0442\u0440\u043e\u0438\u0442\u0441\u044f \u0438\u0437 \u0432\u0430\u0448\u0435\u0433\u043e \u043b\u0438\u0447\u043d\u043e\u0433\u043e \u043d\u0430\u0431\u043e\u0440\u0430 \u0443\u0441\u043b\u0443\u0433. \u0414\u043e\u0431\u0430\u0432\u044c\u0442\u0435 \u043f\u0435\u0440\u0432\u0443\u044e \u0443\u0441\u043b\u0443\u0433\u0443 \u0432\u043e \u0432\u043a\u043b\u0430\u0434\u043a\u0435 \u00ab\u0423\u043f\u0440\u0430\u0432\u043b\u0435\u043d\u0438\u0435\u00bb \u2014 \u043e\u043d\u0430 \u0441\u0440\u0430\u0437\u0443 \u043f\u043e\u044f\u0432\u0438\u0442\u0441\u044f \u0437\u0434\u0435\u0441\u044c.</p>

        <button type="button" class="primary empty-state-action" data-goto-services>\u0414\u043e\u0431\u0430\u0432\u0438\u0442\u044c \u043f\u0435\u0440\u0432\u0443\u044e \u0443\u0441\u043b\u0443\u0433\u0443</button>

      </div>`;

    servicesList.querySelector("[data-goto-services]")?.addEventListener("click", () => {

      setActiveTab("settings");

      serviceNameInput?.focus();

    });

    return;

  }

  servicesList.innerHTML = appState.services.map((service, index) => {

    const weight = appState.weights[service.code] ?? 0;

    const weightClass = weight > 0 ? "weight-positive" : (weight < 0 ? "weight-negative" : "weight-neutral");

    const label = `${index + 1}. ${service.name}`;

    return `

      <div class="weight-row">

        <div class="weight-row-main">

          <span class="weight-row-name" title="${escapeAttribute(label)}">${escapeHtml(label)}</span>

          <span class="weight-row-meta">${formatMoney(service.rate)} / ${escapeHtml(service.unit)} · ${categoryLabel(service.category)}</span>

        </div>

        <div class="weight-stepper">

          <button type="button" data-weight-step="-1" data-weight-target="${escapeAttribute(service.code)}" aria-label="Уменьшить вес">−</button>

          <input type="text" inputmode="numeric" autocomplete="off" value="${weight}" class="${weightClass}" data-weight-code="${escapeAttribute(service.code)}" aria-label="Вес услуги" />

          <button type="button" data-weight-step="1" data-weight-target="${escapeAttribute(service.code)}" aria-label="Увеличить вес">+</button>

        </div>

      </div>

    `;

  }).join("");

  servicesList.querySelectorAll("[data-weight-step]").forEach((button) => {

    button.addEventListener("click", () => {

      const code = button.dataset.weightTarget;

      const current = Number(appState.weights[code] ?? 0);

      const next = clampWeightValue(current + Number(button.dataset.weightStep));

      if (next === current) {

        return;

      }

      appState.weights[code] = next;

      renderServices();

    });

  });

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

    servicesAdminList.innerHTML = `

      <div class="empty-state">

        <div class="empty-state-title">\u0421\u043f\u0438\u0441\u043e\u043a \u0443\u0441\u043b\u0443\u0433 \u043f\u0443\u0441\u0442</div>

        <p class="empty-state-text">\u0417\u0430\u043f\u043e\u043b\u043d\u0438\u0442\u0435 \u0444\u043e\u0440\u043c\u0443 \u0441\u043f\u0440\u0430\u0432\u0430: \u043d\u0430\u0437\u0432\u0430\u043d\u0438\u0435, \u0435\u0434\u0438\u043d\u0438\u0446\u0443, \u0441\u0442\u043e\u0438\u043c\u043e\u0441\u0442\u044c \u0438 \u0433\u0440\u0443\u043f\u043f\u0443. \u0423\u0441\u043b\u0443\u0433\u0430 \u0441\u0440\u0430\u0437\u0443 \u0441\u0442\u0430\u043d\u0435\u0442 \u0434\u043e\u0441\u0442\u0443\u043f\u043d\u0430 \u0432 \u0440\u0430\u0441\u0447\u0451\u0442\u0435.</p>

        <button type="button" class="primary empty-state-action" data-focus-service-form>\u0421\u043e\u0437\u0434\u0430\u0442\u044c \u0443\u0441\u043b\u0443\u0433\u0443</button>

      </div>`;

    servicesAdminList.querySelector("[data-focus-service-form]")?.addEventListener("click", () => {

      serviceNameInput?.focus();

    });

    setServiceFormMessage("\u0423 \u0432\u0430\u0441 \u043f\u043e\u043a\u0430 \u043d\u0435\u0442 \u0443\u0441\u043b\u0443\u0433. \u041c\u043e\u0436\u043d\u043e \u0441\u043e\u0437\u0434\u0430\u0442\u044c \u043f\u0435\u0440\u0432\u0443\u044e \u043f\u0440\u044f\u043c\u043e \u0441\u0435\u0439\u0447\u0430\u0441.", "muted");

    updateServicePercentHint();

    return;

  }

  const percentMap = buildEffectivePercentMap(appState.services);

  servicesAdminList.innerHTML = appState.services.map((service) => `

    <article class="settings-item glass">

      <div>

        <strong class="truncate-text" title="${escapeHtml(service.name)}">${escapeHtml(service.name)}</strong>

        <div class="service-meta">${formatMoney(service.rate)} / ${escapeHtml(service.unit)}</div>

        <div class="service-meta">${categoryLabel(service.category)}, \u0442\u0435\u043a\u0443\u0449\u0438\u0439 \u043f\u0440\u043e\u0446\u0435\u043d\u0442 ${escapeHtml(percentText(service, percentMap))}</div>

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

      generatedAt: getSelectedCalculationDateRFC3339(),

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

