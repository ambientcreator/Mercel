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

function canDeleteArchiveCalculation(saved) {

  if (!saved || !appState.session?.authenticated) {

    return false;

  }

  return Boolean(appState.session?.canAdmin || appState.session?.canManage || saved.createdBy === appState.session?.user?.username);

}

function renderArchiveDetails(saved) {

  const canCopyArchiveServices = Boolean((appState.session?.canAdmin || appState.session?.canCopyArchiveServices) && saved);

  const canDeleteArchive = canDeleteArchiveCalculation(saved);

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

    archiveBody.innerHTML = '<tr><td colspan="5" class="placeholder">Архив пока пуст или расчёт ещё не выбран.</td></tr>';

    return;

  }

  const items = Array.isArray(saved.items) ? saved.items : [];

  archiveTitle.textContent = saved.title || "Архивный расчёт";

  archiveTotal.textContent = formatMoney(saved.totalAmount || 0);

  archiveMeta.innerHTML = `<span>Целевая сумма: <strong>${formatMoney(saved.targetAmount || 0)}</strong></span><span>Автор: <strong>${escapeHtml(saved.createdBy || "Не указан")}</strong></span><span class="history-date">${formatDate(saved.createdAt)}</span>${isEditingArchive ? '<span class="archive-inline-note">Режим редактирования архива активен.</span>' : ""}`;

  if (!items.length) {

    archiveBody.innerHTML = '<tr><td colspan="5" class="placeholder">В этом архиве пока нет услуг для отображения.</td></tr>';

    return;

  }

  archiveBody.innerHTML = items.map((item, index) => {

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

        allocationPercent: source.allocationPercent ?? null,

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

async function copyArchiveServicesToAdmin(calculationId) {

  if (!(appState.session?.canAdmin || appState.session?.canCopyArchiveServices) || !calculationId) {

    return;

  }

  try {

    const result = await api.CopyArchiveServicesToAdmin(calculationId);

    await refreshBootstrap({ keepArchiveSelection: true });

    setActiveTab("settings");

    const created = Number(result?.created || 0);

    const updated = Number(result?.updated || 0);

    setMessage(`Услуги из архива скопированы к вам. Создано: ${created}, обновлено: ${updated}.`, "success");

  } catch (error) {

    setMessage(normalizeErrorText(error, "\u041d\u0435 \u0443\u0434\u0430\u043b\u043e\u0441\u044c \u0441\u043a\u043e\u043f\u0438\u0440\u043e\u0432\u0430\u0442\u044c \u0443\u0441\u043b\u0443\u0433\u0438 \u0438\u0437 \u0430\u0440\u0445\u0438\u0432\u0430."), "error");

  }

}

async function deleteCalculation(calculationId) {

  const id = Number(calculationId);

  const calculation = appState.savedCalculations.find((item) => item.id === id);

  if (!calculation || !canDeleteArchiveCalculation(calculation)) {

    setMessage("Недостаточно прав для удаления этого архивного расчёта.", "error");

    return;

  }

  if (!await requestDeleteConfirmation("архивный расчёт", calculation.title || `#${id}`)) {

    return;

  }

  try {

    await api.DeleteCalculation(id);

    if (appState.activeHistoryId === id) {

      appState.activeHistoryId = null;

    }

    if (appState.editingArchiveId === id) {

      appState.editingArchiveId = null;

    }

    await refreshBootstrap({ keepArchiveSelection: true });

    setMessage(`Архивный расчёт ${calculation.title || `#${id}`} удалён.`, "success");

  } catch (error) {

    setMessage(normalizeErrorText(error, "Не удалось удалить архивный расчёт."), "error");

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

    if (archiveDateInput) {

      archiveDateInput.value = "";

    }

    renderArchiveDetails(null);

    return;

  }

  if (!visibleCalculations.some((item) => item.id === appState.activeHistoryId)) {

    appState.activeHistoryId = visibleCalculations[0]?.id ?? null;

  }

  historyList.innerHTML = visibleCalculations.map((item) => {

    const canDelete = canDeleteArchiveCalculation(item);

    return `

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

        <button type="button" class="history-delete" ${canDelete ? `data-delete-id="${item.id}"` : "disabled"}>Удалить</button>

      </div>

    </article>

  `;

  }).join("");

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



