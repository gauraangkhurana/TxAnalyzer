const identityForm = document.getElementById('identity-form');
const welcomeEl = document.getElementById('welcome');
const usernameInput = document.getElementById('username-input');
const createUserBtn = document.getElementById('create-user-btn');
const linkCard = document.getElementById('link-card');
const linkBankBtn = document.getElementById('link-bank-btn');
const statusEl = document.getElementById('status');
const accountsCard = document.getElementById('accounts-card');
const accountsList = document.getElementById('accounts-list');
const pullCard = document.getElementById('pull-card');
const accountSelect = document.getElementById('account-select');
const dateFromInput = document.getElementById('date-from-input');
const dateToInput = document.getElementById('date-to-input');
const pullBtn = document.getElementById('pull-btn');
const pullStatusEl = document.getElementById('pull-status');
const transactionsCard = document.getElementById('transactions-card');
const transactionsBody = document.getElementById('transactions-body');
const categorizeCard = document.getElementById('categorize-card');
const categorizeBtn = document.getElementById('categorize-btn');
const categorizeStatusEl = document.getElementById('categorize-status');
const categorizeResults = document.getElementById('categorize-results');
const pieChartEl = document.getElementById('pie-chart');
const categoryLegendEl = document.getElementById('category-legend');
const categoryBarsEl = document.getElementById('category-bars');
const chartTooltipEl = document.getElementById('chart-tooltip');

let accountOptions = [];

// Fixed hue order - never reassigned per category, matches the CSS
// custom properties declared in index.html.
const SERIES_COLORS = [
  'var(--series-1)', 'var(--series-2)', 'var(--series-3)', 'var(--series-4)',
  'var(--series-5)', 'var(--series-6)', 'var(--series-7)', 'var(--series-8)',
];

function setCategorizeStatus(message, isError) {
  categorizeStatusEl.textContent = message || '';
  categorizeStatusEl.classList.toggle('error', Boolean(isError));
}

function formatCategoryName(category) {
  if (category === 'unknown') return 'Unknown';
  if (category === 'Other') return 'Other';
  return category
    .toLowerCase()
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

// Categories arrive pre-sorted, highest spend first. Past the palette's 8
// validated slots, fold the tail into "Other" rather than generating a 9th
// color (a generated hue can't be told apart from an existing one).
function foldToPaletteSize(categories) {
  if (categories.length <= SERIES_COLORS.length) return categories;

  const kept = categories.slice(0, SERIES_COLORS.length - 1);
  const rest = categories.slice(SERIES_COLORS.length - 1);
  const other = rest.reduce(
    (acc, c) => ({
      category: 'Other',
      amount: acc.amount + c.amount,
      count: acc.count + c.count,
      percentage: acc.percentage + c.percentage,
    }),
    { category: 'Other', amount: 0, count: 0, percentage: 0 },
  );
  return [...kept, other];
}

// --- Chart tooltip (shared by the pie slices and the bar rows) ---

function showChartTooltip(x, y, category, amount, percentage) {
  chartTooltipEl.innerHTML = '';

  const value = document.createElement('div');
  value.className = 'tooltip-value';
  value.textContent = `$${amount.toFixed(2)}`;
  chartTooltipEl.appendChild(value);

  const label = document.createElement('div');
  label.className = 'tooltip-label';
  label.textContent = `${formatCategoryName(category)} · ${percentage.toFixed(1)}%`;
  chartTooltipEl.appendChild(label);

  chartTooltipEl.style.left = `${x}px`;
  chartTooltipEl.style.top = `${y}px`;
  chartTooltipEl.classList.remove('hidden');
}

function hideChartTooltip() {
  chartTooltipEl.classList.add('hidden');
}

// --- SVG pie slice geometry (0-100 viewBox, center 50,50, radius 48) ---

const PIE_CENTER = 50;
const PIE_RADIUS = 48;

function polarToCartesian(angleDeg) {
  const rad = ((angleDeg - 90) * Math.PI) / 180;
  return {
    x: PIE_CENTER + PIE_RADIUS * Math.cos(rad),
    y: PIE_CENTER + PIE_RADIUS * Math.sin(rad),
  };
}

function pieSlicePath(startPct, endPct) {
  // A full-circle slice (one category = 100%) needs a special case: SVG
  // arcs can't span a full 360 degrees in one command.
  if (endPct - startPct >= 100) {
    const top = { x: PIE_CENTER, y: PIE_CENTER - PIE_RADIUS };
    return `M ${PIE_CENTER} ${PIE_CENTER} L ${top.x} ${top.y} A ${PIE_RADIUS} ${PIE_RADIUS} 0 1 1 ${top.x - 0.01} ${top.y} Z`;
  }
  const start = polarToCartesian(endPct * 3.6);
  const end = polarToCartesian(startPct * 3.6);
  const largeArc = endPct - startPct > 50 ? 1 : 0;
  return `M ${PIE_CENTER} ${PIE_CENTER} L ${start.x} ${start.y} A ${PIE_RADIUS} ${PIE_RADIUS} 0 ${largeArc} 0 ${end.x} ${end.y} Z`;
}

function setPullStatus(message, isError) {
  pullStatusEl.textContent = message || '';
  pullStatusEl.classList.toggle('error', Boolean(isError));
}

function defaultDateRange() {
  const to = new Date();
  const from = new Date();
  from.setDate(from.getDate() - 90);
  const fmt = (d) => d.toISOString().slice(0, 10);
  return { from: fmt(from), to: fmt(to) };
}

function setStatus(message, isError) {
  statusEl.textContent = message || '';
  statusEl.classList.toggle('error', Boolean(isError));
}

async function apiRequest(path, options) {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(body.error || `Request to ${path} failed (${res.status})`);
  }
  return body;
}

function getUserId() {
  const id = localStorage.getItem('userId');
  return id ? Number(id) : null;
}

function showIdentified(username) {
  identityForm.classList.add('hidden');
  welcomeEl.textContent = `Welcome back, ${username}.`;
  welcomeEl.classList.remove('hidden');
  linkCard.classList.remove('hidden');
}

async function loadAccounts() {
  const userId = getUserId();
  if (!userId) return;

  const profile = await apiRequest(`/v1/bank/users/${userId}/banks`);
  const banks = profile.banks || [];

  accountsList.innerHTML = '';
  accountOptions = [];
  accountSelect.innerHTML = '';

  if (banks.length === 0) {
    accountsCard.classList.add('hidden');
    pullCard.classList.add('hidden');
    categorizeCard.classList.add('hidden');
    return;
  }

  banks.forEach((bank) => {
    const bankDiv = document.createElement('div');
    bankDiv.className = 'bank';

    const nameDiv = document.createElement('div');
    nameDiv.className = 'bank-name';
    nameDiv.textContent = bank.bank_name;
    bankDiv.appendChild(nameDiv);

    (bank.accounts || []).forEach((account) => {
      const accountDiv = document.createElement('div');
      accountDiv.className = 'account';
      accountDiv.innerHTML = `<span>${account.account_id}</span><span class="account-type">${account.account_type}</span>`;
      bankDiv.appendChild(accountDiv);

      accountOptions.push({ bankName: bank.bank_name, accountId: account.account_id });
      const option = document.createElement('option');
      option.value = String(accountOptions.length - 1);
      option.textContent = `${bank.bank_name} - ${account.account_id} (${account.account_type})`;
      accountSelect.appendChild(option);
    });

    accountsList.appendChild(bankDiv);
  });

  accountsCard.classList.remove('hidden');
  pullCard.classList.remove('hidden');
  categorizeCard.classList.remove('hidden');

  if (!dateFromInput.value || !dateToInput.value) {
    const { from, to } = defaultDateRange();
    dateFromInput.value = from;
    dateToInput.value = to;
  }
}

async function loadTransactions() {
  const userId = getUserId();
  if (!userId) return;

  const result = await apiRequest(`/v1/transactions?user_id=${userId}`);
  const transactions = result.transactions || [];

  transactionsBody.innerHTML = '';
  if (transactions.length === 0) {
    transactionsCard.classList.add('hidden');
    return;
  }

  transactions.forEach((tx) => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td>${tx.date}</td>
      <td>${tx.bank_name}</td>
      <td>${tx.name}${tx.pending ? ' <span class="pending-tag">PENDING</span>' : ''}</td>
      <td>${tx.category || ''}</td>
      <td class="amount">${tx.amount.toFixed(2)}</td>
    `;
    transactionsBody.appendChild(row);
  });

  transactionsCard.classList.remove('hidden');
}

async function pullTransactions() {
  const userId = getUserId();
  const selected = accountOptions[Number(accountSelect.value)];
  if (!userId || !selected) return;

  const dateFrom = dateFromInput.value;
  const dateTo = dateToInput.value;
  if (!dateFrom || !dateTo) {
    setPullStatus('Pick a date range first.', true);
    return;
  }

  pullBtn.disabled = true;
  setPullStatus('Pulling transactions from Plaid...');

  try {
    const result = await apiRequest('/v1/transactions/pull', {
      method: 'POST',
      body: JSON.stringify({
        user_id: userId,
        bank_name: selected.bankName,
        account_id: selected.accountId,
        date_from: dateFrom,
        date_to: dateTo,
      }),
    });
    setPullStatus(`Pulled ${result.pulled} transaction(s).`);
    await loadTransactions();
  } catch (err) {
    setPullStatus(err.message, true);
  } finally {
    pullBtn.disabled = false;
  }
}

const SVG_NS = 'http://www.w3.org/2000/svg';

function renderCategorySpend(rawCategories) {
  const categories = foldToPaletteSize(rawCategories);

  pieChartEl.replaceChildren();
  categoryLegendEl.replaceChildren();
  categoryBarsEl.replaceChildren();

  // Every category needs its slice, its legend row, and its bar row to
  // agree on the same color and highlight together on hover - built in one
  // pass so they can never drift apart.
  let cumulative = 0;
  categories.forEach((c, i) => {
    const color = SERIES_COLORS[i];
    const label = formatCategoryName(c.category);
    const start = cumulative;
    cumulative += c.percentage;

    // --- pie slice ---
    const slice = document.createElementNS(SVG_NS, 'path');
    slice.setAttribute('d', pieSlicePath(start, cumulative));
    slice.style.fill = color;
    slice.classList.add('pie-slice');
    pieChartEl.appendChild(slice);

    // --- legend row ---
    const legendRow = document.createElement('div');
    legendRow.className = 'legend-row';

    const swatch = document.createElement('span');
    swatch.className = 'legend-swatch';
    swatch.style.background = color;
    legendRow.appendChild(swatch);

    const labelEl = document.createElement('span');
    labelEl.className = 'legend-label';
    labelEl.textContent = label;
    legendRow.appendChild(labelEl);

    const amountEl = document.createElement('span');
    amountEl.className = 'legend-amount';
    amountEl.textContent = `$${c.amount.toFixed(2)} (${c.percentage.toFixed(1)}%)`;
    legendRow.appendChild(amountEl);

    categoryLegendEl.appendChild(legendRow);

    // --- bar row ---
    const barRow = document.createElement('div');
    barRow.className = 'bar-row';

    const barLabel = document.createElement('div');
    barLabel.className = 'bar-row-label';
    const barLabelName = document.createElement('span');
    barLabelName.textContent = label;
    const barLabelAmount = document.createElement('span');
    barLabelAmount.textContent = `$${c.amount.toFixed(2)} (${c.percentage.toFixed(1)}%)`;
    barLabel.appendChild(barLabelName);
    barLabel.appendChild(barLabelAmount);
    barRow.appendChild(barLabel);

    const barTrack = document.createElement('div');
    barTrack.className = 'bar-track';
    const barFill = document.createElement('div');
    barFill.className = 'bar-fill';
    barFill.style.width = `${c.percentage}%`;
    barFill.style.background = color;
    barTrack.appendChild(barFill);
    barRow.appendChild(barTrack);

    categoryBarsEl.appendChild(barRow);

    // --- hover: tooltip + cross-highlight the slice and legend row together ---
    const highlight = (on) => {
      slice.classList.toggle('slice-hover', on);
      legendRow.classList.toggle('slice-hover', on);
    };
    const onEnter = (e) => {
      highlight(true);
      showChartTooltip(e.clientX, e.clientY, c.category, c.amount, c.percentage);
    };
    const onMove = (e) => showChartTooltip(e.clientX, e.clientY, c.category, c.amount, c.percentage);
    const onLeave = () => {
      highlight(false);
      hideChartTooltip();
    };

    slice.addEventListener('pointerenter', onEnter);
    slice.addEventListener('pointermove', onMove);
    slice.addEventListener('pointerleave', onLeave);
    legendRow.addEventListener('pointerenter', onEnter);
    legendRow.addEventListener('pointermove', onMove);
    legendRow.addEventListener('pointerleave', onLeave);
    barRow.addEventListener('pointerenter', onEnter);
    barRow.addEventListener('pointermove', onMove);
    barRow.addEventListener('pointerleave', onLeave);
  });

  categorizeResults.classList.remove('hidden');
}

async function categorizeSpend() {
  const userId = getUserId();
  if (!userId) return;

  categorizeBtn.disabled = true;
  setCategorizeStatus('Categorizing...');

  try {
    const result = await apiRequest(`/v1/transactions/categories?user_id=${userId}`);
    if (!result.categories || result.categories.length === 0) {
      setCategorizeStatus('No spend to categorize yet - pull some transactions first.', true);
      categorizeResults.classList.add('hidden');
      return;
    }
    renderCategorySpend(result.categories);
    setCategorizeStatus('');
  } catch (err) {
    setCategorizeStatus(err.message, true);
  } finally {
    categorizeBtn.disabled = false;
  }
}

async function createUser() {
  const username = usernameInput.value.trim();
  if (!username) {
    setStatus('Enter a name first.', true);
    return;
  }

  createUserBtn.disabled = true;
  try {
    const user = await apiRequest('/v1/users', {
      method: 'POST',
      body: JSON.stringify({ username }),
    });
    localStorage.setItem('userId', String(user.user_id));
    localStorage.setItem('username', user.username);
    showIdentified(user.username);
    await loadAccounts();
  } catch (err) {
    setStatus(err.message, true);
  } finally {
    createUserBtn.disabled = false;
  }
}

async function linkBank() {
  const userId = getUserId();
  if (!userId) return;

  linkBankBtn.disabled = true;
  setStatus('Creating link token...');

  try {
    const { link_token: linkToken } = await apiRequest('/v1/bank/plaid/link-token', {
      method: 'POST',
      body: JSON.stringify({ user_id: userId }),
    });

    const handler = Plaid.create({
      token: linkToken,
      onSuccess: async (publicToken, metadata) => {
        setStatus('Linking account...');
        try {
          const result = await apiRequest('/v1/bank/plaid/exchange', {
            method: 'POST',
            body: JSON.stringify({
              user_id: userId,
              bank_name: metadata.institution.name,
              public_token: publicToken,
            }),
          });
          setStatus(`Linked ${result.accounts.length} account(s) from ${metadata.institution.name}.`);
          await loadAccounts();
        } catch (err) {
          setStatus(err.message, true);
        } finally {
          linkBankBtn.disabled = false;
        }
      },
      onExit: (err) => {
        linkBankBtn.disabled = false;
        if (err) {
          setStatus(err.display_message || err.error_message || 'Link exited with an error.', true);
        } else {
          setStatus('');
        }
      },
    });

    handler.open();
  } catch (err) {
    setStatus(err.message, true);
    linkBankBtn.disabled = false;
  }
}

createUserBtn.addEventListener('click', createUser);
linkBankBtn.addEventListener('click', linkBank);
pullBtn.addEventListener('click', pullTransactions);
categorizeBtn.addEventListener('click', categorizeSpend);

(function init() {
  const userId = getUserId();
  const username = localStorage.getItem('username');
  if (userId && username) {
    showIdentified(username);
    loadAccounts().catch((err) => setStatus(err.message, true));
    loadTransactions().catch((err) => setPullStatus(err.message, true));
  }
})();
