const identityForm = document.getElementById('identity-form');
const welcomeEl = document.getElementById('welcome');
const usernameInput = document.getElementById('username-input');
const createUserBtn = document.getElementById('create-user-btn');
const linkCard = document.getElementById('link-card');
const linkBankBtn = document.getElementById('link-bank-btn');
const statusEl = document.getElementById('status');
const accountsCard = document.getElementById('accounts-card');
const accountsList = document.getElementById('accounts-list');

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
  if (banks.length === 0) {
    accountsCard.classList.add('hidden');
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
    });

    accountsList.appendChild(bankDiv);
  });

  accountsCard.classList.remove('hidden');
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

(function init() {
  const userId = getUserId();
  const username = localStorage.getItem('username');
  if (userId && username) {
    showIdentified(username);
    loadAccounts().catch((err) => setStatus(err.message, true));
  }
})();
