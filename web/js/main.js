import store from './state/store.js';
import * as userManager from './modules/userManager.js';
import * as clientManager from './modules/clientManager.js';
import * as providerManager from './modules/providerManager.js';
import * as uiManager from './modules/uiManager.js';
import { initUIEffects, showNotification } from './utils/ui-effects.js';

// Global DOM references
const elements = {
    // User elements
    usersList: document.getElementById('users-list'),
    addUserBtn: document.getElementById('add-user-btn'),
    userModal: document.getElementById('user-modal'),
    modalTitle: document.getElementById('modal-title'),
    userForm: document.getElementById('user-form'),
    userId: document.getElementById('user-id'),
    username: document.getElementById('username'),
    password: document.getElementById('password'),

    // Client elements
    clientsList: document.getElementById('clients-list'),
    addClientBtn: document.getElementById('add-client-btn'),
    clientModal: document.getElementById('client-modal'),
    clientModalTitle: document.getElementById('client-modal-title'),
    clientForm: document.getElementById('client-form'),
    clientIdInput: document.getElementById('client-id-input'),
    clientId: document.getElementById('client-id'),
    clientName: document.getElementById('client-name'),
    clientSecret: document.getElementById('client-secret'),
    redirectUris: document.getElementById('redirect-uris'),

    // Provider elements
    providersList: document.getElementById('providers-list'),
    addProviderBtn: document.getElementById('add-provider-btn'),
    providerModal: document.getElementById('provider-modal'),
    providerModalTitle: document.getElementById('provider-modal-title'),
    providerForm: document.getElementById('provider-form'),
    providerId: document.getElementById('provider-id'),
    providerName: document.getElementById('provider-name'),
    providerType: document.getElementById('provider-type'),
    providerClientId: document.getElementById('provider-client-id'),
    providerClientSecret: document.getElementById('provider-client-secret'),
    providerEnabled: document.getElementById('provider-enabled'),
    providerTenantId: document.getElementById('provider-tenant-id'),
    tenantIdGroup: document.querySelector('.tenant-id-group'),

    // Blocked users elements
    blockedUsersList: document.getElementById('blocked-users-list'),
    refreshBlockedBtn: document.getElementById('refresh-blocked-btn'),

    // Delete modal elements
    deleteModal: document.getElementById('delete-modal'),
    confirmDeleteBtn: document.getElementById('confirm-delete'),
    cancelDeleteBtn: document.getElementById('cancel-delete'),
    deleteConfirmationMessage: document.getElementById('delete-confirmation-message'),
};

/**
 * Initialize event listeners for the user interface
 */
function initEventListeners() {
    // Initialize navigation between sections
    uiManager.initNavigation();

    // Initialize theme
    uiManager.initThemeToggle();

    // Configure click events outside modals
    uiManager.setupModalOutsideClicks();

    // Events for user management
    if (elements.addUserBtn) {
        elements.addUserBtn.addEventListener('click', showAddUserModal);
    }
    if (elements.userForm) {
        elements.userForm.addEventListener('submit', handleUserFormSubmit);
    }

    // Event for Cancel button for users
    const cancelUserBtn = document.getElementById('cancel-btn');
    if (cancelUserBtn) {
        cancelUserBtn.addEventListener('click', () => {
            uiManager.toggleModal(elements.userModal, false);
        });
    }

    // Events for client management
    if (elements.addClientBtn) {
        elements.addClientBtn.addEventListener('click', showAddClientModal);
    }
    if (elements.clientForm) {
        elements.clientForm.addEventListener('submit', handleClientFormSubmit);
    }

    // Event for Cancel button for clients
    const cancelClientBtn = document.getElementById('cancel-client-btn');
    if (cancelClientBtn) {
        cancelClientBtn.addEventListener('click', () => {
            uiManager.toggleModal(elements.clientModal, false);
        });
    }

    // Events for provider management
    if (elements.addProviderBtn) {
        elements.addProviderBtn.addEventListener('click', showAddProviderModal);
    }
    if (elements.providerForm) {
        elements.providerForm.addEventListener('submit', handleProviderFormSubmit);
    }
    if (elements.providerType) {
        elements.providerType.addEventListener('change', function () {
            toggleTenantIdField(this.value);
        });
    }

    // Event for Cancel button for providers
    const cancelProviderBtn = document.getElementById('cancel-provider-btn');
    if (cancelProviderBtn) {
        cancelProviderBtn.addEventListener('click', () => {
            uiManager.toggleModal(elements.providerModal, false);
        });
    }

    // Events for blocked users
    if (elements.refreshBlockedBtn) {
        elements.refreshBlockedBtn.addEventListener('click', () => {
            // Forcer le rechargement des utilisateurs bloqués
            userManager.loadBlockedUsers(true);
        });
    }

    // Events for delete modal
    if (elements.confirmDeleteBtn) {
        elements.confirmDeleteBtn.addEventListener('click', confirmDelete);
    }
    if (elements.cancelDeleteBtn) {
        elements.cancelDeleteBtn.addEventListener('click', () => {
            uiManager.toggleModal(elements.deleteModal, false);
        });
    }

    // Add handlers for close buttons in modals
    document.querySelectorAll('.modal .close').forEach(closeBtn => {
        closeBtn.addEventListener('click', () => {
            const modal = closeBtn.closest('.modal');
            if (modal) {
                uiManager.toggleModal(modal, false);
            }
        });
    });
}

/**
 * Subscribe to state changes to update UI accordingly
 */
function initStateSubscriptions() {
    // Subscribe to changes in the user list
    store.subscribe('users', (usersState) => {
        if (elements.usersList) {
            uiManager.renderUsersList(usersState.list, elements.usersList);

            // Attach events after rendering the list
            attachUserEventHandlers();
        }
    });

    // Subscribe to changes in the client list
    store.subscribe('clients', (clientsState) => {
        if (elements.clientsList) {
            uiManager.renderClientsList(clientsState.list, elements.clientsList);

            // Attach events after rendering the list
            attachClientEventHandlers();
        }
    });

    // Subscribe to changes in the provider list
    store.subscribe('providers', (providersState) => {
        if (elements.providersList) {
            uiManager.renderProvidersList(providersState.list, elements.providersList);

            // Attach events after rendering the list
            attachProviderEventHandlers();
        }
    });

    // Subscribe to changes in the blocked users list
    store.subscribe('blockedUsers', (blockedState) => {
        if (elements.blockedUsersList) {
            uiManager.renderBlockedUsersList(blockedState.groupedData, elements.blockedUsersList);

            // Attach events after rendering the list
            attachBlockedUsersEventHandlers();
        }
    });
}

/**
 * Attach event handlers to elements in the user list
 */
function attachUserEventHandlers() {
    // Handler for source toggle button
    const toggleBtn = document.getElementById('toggle-source');
    if (toggleBtn) {
        toggleBtn.addEventListener('click', userManager.toggleUserSource);
    }

    // Handlers for edit buttons
    document.querySelectorAll('#users-list .action-btn.edit').forEach(btn => {
        btn.addEventListener('click', () => {
            showEditUserModal(btn.dataset.id, btn.dataset.username);
        });
    });

    // Handlers for delete buttons
    document.querySelectorAll('#users-list .action-btn.delete[data-type="user"]').forEach(btn => {
        btn.addEventListener('click', () => {
            uiManager.showDeleteConfirmation(btn.dataset.id, 'user', btn.dataset.name);
        });
    });
}

/**
 * Attach event handlers to elements in the client list
 */
function attachClientEventHandlers() {
    // Handlers for edit buttons
    document.querySelectorAll('#clients-list .action-btn.edit').forEach(btn => {
        btn.addEventListener('click', () => {
            showEditClientModal(btn.dataset.id, btn.dataset.name, btn.dataset.secret, btn.dataset.uris);
        });
    });

    // Handlers for delete buttons
    document.querySelectorAll('#clients-list .action-btn.delete[data-type="client"]').forEach(btn => {
        btn.addEventListener('click', () => {
            uiManager.showDeleteConfirmation(btn.dataset.id, 'client', btn.dataset.name);
        });
    });
}

/**
 * Attach event handlers to elements in the provider list
 */
function attachProviderEventHandlers() {
    // Handlers for edit buttons
    document.querySelectorAll('#providers-list .action-btn.edit').forEach(btn => {
        btn.addEventListener('click', () => {
            showEditProviderModal(
                btn.dataset.id,
                btn.dataset.name,
                btn.dataset.type,
                btn.dataset.clientId,
                btn.dataset.tenantId,
                btn.dataset.enabled === 'true'
            );
        });
    });

    // Handlers for delete buttons
    document.querySelectorAll('#providers-list .action-btn.delete[data-type="provider"]').forEach(btn => {
        btn.addEventListener('click', () => {
            uiManager.showDeleteConfirmation(btn.dataset.id, 'provider', btn.dataset.name);
        });
    });
}

/**
 * Attach event handlers to elements in the blocked users list
 */
function attachBlockedUsersEventHandlers() {
    // Handlers for unblock buttons
    document.querySelectorAll('#blocked-users-list .action-btn.unblock').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                const result = await userManager.unblockUser(
                    btn.dataset.identifier,
                    btn.dataset.type
                );
                alert(result.message || 'Successfully unblocked');
            } catch (error) {
                // Error is already handled in unblockUser
            }
        });
    });
}

/**
 * User modal management
 */
function showAddUserModal() {
    elements.modalTitle.textContent = 'Add User';
    elements.userId.value = '';
    elements.username.value = '';
    elements.password.value = '';
    elements.password.required = true;

    if (document.getElementById('password-hint')) {
        document.getElementById('password-hint').style.display = 'none';
    }

    uiManager.toggleModal(elements.userModal, true);
}

function showEditUserModal(id, usernameValue) {
    elements.modalTitle.textContent = 'Edit User';
    elements.userId.value = id;
    elements.username.value = usernameValue;
    elements.password.value = '';
    elements.password.required = false;

    if (document.getElementById('password-hint')) {
        document.getElementById('password-hint').style.display = 'block';
    }

    uiManager.toggleModal(elements.userModal, true);
}

/**
 * Client modal management
 */
function showAddClientModal() {
    elements.clientModalTitle.textContent = 'Add OAuth Client';
    elements.clientIdInput.value = '';
    elements.clientId.value = '';
    elements.clientId.readOnly = false;
    elements.clientName.value = '';
    elements.clientSecret.value = '';
    elements.clientSecret.required = true;

    if (document.getElementById('secret-hint')) {
        document.getElementById('secret-hint').style.display = 'none';
    }

    elements.redirectUris.value = '';
    uiManager.toggleModal(elements.clientModal, true);
}

function showEditClientModal(id, name, secret, uris) {
    elements.clientModalTitle.textContent = 'Edit OAuth Client';
    elements.clientIdInput.value = id;
    elements.clientId.value = id;
    elements.clientId.readOnly = true;
    elements.clientName.value = name;
    elements.clientSecret.value = '';
    elements.clientSecret.required = false;

    if (document.getElementById('secret-hint')) {
        document.getElementById('secret-hint').style.display = 'block';
    }

    elements.redirectUris.value = uris;
    uiManager.toggleModal(elements.clientModal, true);
}

/**
 * Provider modal management
 */
function showAddProviderModal() {
    elements.providerModalTitle.textContent = 'Add Authentication Provider';
    elements.providerId.value = '';
    elements.providerName.value = '';
    elements.providerType.value = 'google';
    elements.providerClientId.value = '';
    elements.providerClientSecret.value = '';
    elements.providerClientSecret.required = true;
    elements.providerEnabled.checked = false;
    elements.providerTenantId.value = '';

    if (document.getElementById('provider-secret-hint')) {
        document.getElementById('provider-secret-hint').style.display = 'none';
    }

    toggleTenantIdField(elements.providerType.value);
    uiManager.toggleModal(elements.providerModal, true);
}

function showEditProviderModal(id, name, type, clientId, tenantId, enabled) {
    elements.providerModalTitle.textContent = 'Edit Authentication Provider';
    elements.providerId.value = id;
    elements.providerName.value = name;
    elements.providerType.value = type;
    elements.providerClientId.value = clientId;
    elements.providerClientSecret.value = '';
    elements.providerClientSecret.required = false;
    elements.providerEnabled.checked = enabled;
    elements.providerTenantId.value = tenantId || '';

    if (document.getElementById('provider-secret-hint')) {
        document.getElementById('provider-secret-hint').style.display = 'block';
    }

    toggleTenantIdField(type);
    uiManager.toggleModal(elements.providerModal, true);
}

function toggleTenantIdField(providerType) {
    if (elements.tenantIdGroup) {
        elements.tenantIdGroup.style.display = providerManager.isTenantIdRequired(providerType) ? 'block' : 'none';
    }
}

/**
 * Form submission handlers
 */
async function handleUserFormSubmit(e) {
    e.preventDefault();

    const isEditing = elements.userId.value !== '';
    const userData = {
        username: elements.username.value
    };

    if (elements.password.value) {
        userData.password = elements.password.value;
    }

    try {
        if (isEditing) {
            await userManager.updateUser(elements.userId.value, userData);
        } else {
            await userManager.createUser(userData);
        }

        uiManager.toggleModal(elements.userModal, false);
    } catch (error) {
        // Error is already handled in userManager methods
    }
}

async function handleClientFormSubmit(e) {
    e.preventDefault();

    const isEditing = elements.clientIdInput.value !== '';
    const redirectUrisList = clientManager.formatRedirectUris(elements.redirectUris.value, 'array');

    const clientData = {
        id: elements.clientId.value,
        name: elements.clientName.value,
        redirect_uris: redirectUrisList
    };

    if (elements.clientSecret.value) {
        clientData.secret = elements.clientSecret.value;
    }

    try {
        if (isEditing) {
            await clientManager.updateClient(elements.clientIdInput.value, clientData);
        } else {
            await clientManager.createClient(clientData);
        }

        uiManager.toggleModal(elements.clientModal, false);
    } catch (error) {
        // Error is already handled in clientManager methods
    }
}

async function handleProviderFormSubmit(e) {
    e.preventDefault();

    const isEditing = elements.providerId.value !== '';
    const providerData = {
        name: elements.providerName.value,
        type: elements.providerType.value,
        client_id: elements.providerClientId.value,
        enabled: elements.providerEnabled.checked
    };

    // Only include tenant_id for Microsoft providers
    if (elements.providerType.value === 'microsoft' && elements.providerTenantId.value) {
        providerData.tenant_id = elements.providerTenantId.value;
    }

    // Only include secret if provided
    if (elements.providerClientSecret.value) {
        providerData.client_secret = elements.providerClientSecret.value;
    }

    try {
        if (isEditing) {
            await providerManager.updateProvider(elements.providerId.value, providerData);
        } else {
            await providerManager.createProvider(providerData);
        }

        uiManager.toggleModal(elements.providerModal, false);
    } catch (error) {
        // Error is already handled in providerManager methods
    }
}

/**
 * Confirm item deletion
 */
async function confirmDelete() {
    const deleteState = store.getState('ui').modals.delete;
    if (!deleteState.id || !deleteState.type) return;

    try {
        switch (deleteState.type) {
            case 'user':
                await userManager.deleteUser(deleteState.id);
                break;
            case 'client':
                await clientManager.deleteClient(deleteState.id);
                break;
            case 'provider':
                await providerManager.deleteProvider(deleteState.id);
                break;
        }

        uiManager.toggleModal(elements.deleteModal, false);
    } catch (error) {
        // Error is already handled in manager methods
    }
}

/**
 * Load initial data when page loads
 */
function loadInitialData() {
    // Display a global loading indicator
    const loadingIndicator = document.createElement('div');
    loadingIndicator.className = 'global-loading-indicator';
    loadingIndicator.innerHTML = '<div class="spinner"></div><span>Loading data...</span>';
    document.body.appendChild(loadingIndicator);

    // Load all data in parallel, independently of the active section
    Promise.all([
        userManager.loadUsers(),
        clientManager.loadClients(),
        providerManager.loadProviders(),
        userManager.loadBlockedUsers()
    ])
        .then(() => {
            console.log('✅ All data successfully loaded');

            // Hide the loading indicator with a transition
            loadingIndicator.classList.add('fade-out');
            setTimeout(() => {
                document.body.removeChild(loadingIndicator);
            }, 500);
        })
        .catch(error => {
            console.error('❌ Error loading data:', error);

            // Transform loading indicator into an error message
            loadingIndicator.innerHTML = `
            <div class="error-icon">❌</div>
            <span>Error loading data. <button id="retry-load">Retry</button></span>
        `;
            loadingIndicator.className = 'global-error-indicator';

            // Add button to retry
            document.getElementById('retry-load').addEventListener('click', () => {
                document.body.removeChild(loadingIndicator);
                loadInitialData();
            });
        });
}

/**
 * Application initialization
 */
document.addEventListener('DOMContentLoaded', function () {
    // Initialize event handlers
    initEventListeners();

    // Subscribe to state changes
    initStateSubscriptions();

    // Initialize UI effects
    initUIEffects();

    // Load initial data
    loadInitialData();

    // Display welcome message
    setTimeout(() => {
        showNotification('Welcome to ZenAuth Admin Console', 'info', 5000);
    }, 1000);
});