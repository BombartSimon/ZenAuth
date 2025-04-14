/**
 * Point d'entrée principal de l'application ZenAuth Admin
 * Orchestre tous les modules et initialise l'application
 */
import store from './state/store.js';
import * as userManager from './modules/userManager.js';
import * as clientManager from './modules/clientManager.js';
import * as providerManager from './modules/providerManager.js';
import * as uiManager from './modules/uiManager.js';

// Références DOM globales
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
 * Initialise les gestionnaires d'événements pour l'interface utilisateur
 */
function initEventListeners() {
    // Initialise la navigation entre les sections
    uiManager.initNavigation();

    // Initialise le thème
    uiManager.initThemeToggle();

    // Configure les événements de clic à l'extérieur des modales
    uiManager.setupModalOutsideClicks();

    // Événements pour la gestion des utilisateurs
    if (elements.addUserBtn) {
        elements.addUserBtn.addEventListener('click', showAddUserModal);
    }
    if (elements.userForm) {
        elements.userForm.addEventListener('submit', handleUserFormSubmit);
    }

    // Événements pour la gestion des clients
    if (elements.addClientBtn) {
        elements.addClientBtn.addEventListener('click', showAddClientModal);
    }
    if (elements.clientForm) {
        elements.clientForm.addEventListener('submit', handleClientFormSubmit);
    }

    // Événements pour la gestion des providers
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

    // Événements pour les utilisateurs bloqués
    if (elements.refreshBlockedBtn) {
        elements.refreshBlockedBtn.addEventListener('click', userManager.loadBlockedUsers);
    }

    // Événements pour la modale de suppression
    if (elements.confirmDeleteBtn) {
        elements.confirmDeleteBtn.addEventListener('click', confirmDelete);
    }
    if (elements.cancelDeleteBtn) {
        elements.cancelDeleteBtn.addEventListener('click', () => {
            uiManager.toggleModal(elements.deleteModal, false);
        });
    }
}

/**
 * S'abonne aux changements d'état pour mettre à jour l'UI en conséquence
 */
function initStateSubscriptions() {
    // S'abonner aux changements dans la liste des utilisateurs
    store.subscribe('users', (usersState) => {
        if (elements.usersList) {
            uiManager.renderUsersList(usersState.list, elements.usersList);

            // Attacher les événements après avoir rendu la liste
            attachUserEventHandlers();
        }
    });

    // S'abonner aux changements dans la liste des clients
    store.subscribe('clients', (clientsState) => {
        if (elements.clientsList) {
            uiManager.renderClientsList(clientsState.list, elements.clientsList);

            // Attacher les événements après avoir rendu la liste
            attachClientEventHandlers();
        }
    });

    // S'abonner aux changements dans la liste des providers
    store.subscribe('providers', (providersState) => {
        if (elements.providersList) {
            uiManager.renderProvidersList(providersState.list, elements.providersList);

            // Attacher les événements après avoir rendu la liste
            attachProviderEventHandlers();
        }
    });

    // S'abonner aux changements dans la liste des utilisateurs bloqués
    store.subscribe('blockedUsers', (blockedState) => {
        if (elements.blockedUsersList) {
            uiManager.renderBlockedUsersList(blockedState.groupedData, elements.blockedUsersList);

            // Attacher les événements après avoir rendu la liste
            attachBlockedUsersEventHandlers();
        }
    });
}

/**
 * Attache les gestionnaires d'événements aux éléments de la liste des utilisateurs
 */
function attachUserEventHandlers() {
    // Gestionnaire pour le bouton de basculement de source
    const toggleBtn = document.getElementById('toggle-source');
    if (toggleBtn) {
        toggleBtn.addEventListener('click', userManager.toggleUserSource);
    }

    // Gestionnaires pour les boutons d'édition
    document.querySelectorAll('#users-list .action-btn.edit').forEach(btn => {
        btn.addEventListener('click', () => {
            showEditUserModal(btn.dataset.id, btn.dataset.username);
        });
    });

    // Gestionnaires pour les boutons de suppression
    document.querySelectorAll('#users-list .action-btn.delete[data-type="user"]').forEach(btn => {
        btn.addEventListener('click', () => {
            uiManager.showDeleteConfirmation(btn.dataset.id, 'user', btn.dataset.name);
        });
    });
}

/**
 * Attache les gestionnaires d'événements aux éléments de la liste des clients
 */
function attachClientEventHandlers() {
    // Gestionnaires pour les boutons d'édition
    document.querySelectorAll('#clients-list .action-btn.edit').forEach(btn => {
        btn.addEventListener('click', () => {
            showEditClientModal(btn.dataset.id, btn.dataset.name, btn.dataset.secret, btn.dataset.uris);
        });
    });

    // Gestionnaires pour les boutons de suppression
    document.querySelectorAll('#clients-list .action-btn.delete[data-type="client"]').forEach(btn => {
        btn.addEventListener('click', () => {
            uiManager.showDeleteConfirmation(btn.dataset.id, 'client', btn.dataset.name);
        });
    });
}

/**
 * Attache les gestionnaires d'événements aux éléments de la liste des providers
 */
function attachProviderEventHandlers() {
    // Gestionnaires pour les boutons d'édition
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

    // Gestionnaires pour les boutons de suppression
    document.querySelectorAll('#providers-list .action-btn.delete[data-type="provider"]').forEach(btn => {
        btn.addEventListener('click', () => {
            uiManager.showDeleteConfirmation(btn.dataset.id, 'provider', btn.dataset.name);
        });
    });
}

/**
 * Attache les gestionnaires d'événements aux éléments de la liste des utilisateurs bloqués
 */
function attachBlockedUsersEventHandlers() {
    // Gestionnaires pour les boutons de déblocage
    document.querySelectorAll('#blocked-users-list .action-btn.unblock').forEach(btn => {
        btn.addEventListener('click', async () => {
            try {
                const result = await userManager.unblockUser(
                    btn.dataset.identifier,
                    btn.dataset.type
                );
                alert(result.message || 'Débloqué avec succès');
            } catch (error) {
                // L'erreur est déjà gérée dans unblockUser
            }
        });
    });
}

/**
 * Gestion des modales utilisateur
 */
function showAddUserModal() {
    elements.modalTitle.textContent = 'Ajouter un utilisateur';
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
    elements.modalTitle.textContent = 'Modifier un utilisateur';
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
 * Gestion des modales client
 */
function showAddClientModal() {
    elements.clientModalTitle.textContent = 'Ajouter un client OAuth';
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
    elements.clientModalTitle.textContent = 'Modifier un client OAuth';
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
 * Gestion des modales provider
 */
function showAddProviderModal() {
    elements.providerModalTitle.textContent = 'Ajouter un fournisseur d\'authentification';
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
    elements.providerModalTitle.textContent = 'Modifier un fournisseur d\'authentification';
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
 * Gestionnaires de soumission des formulaires
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
        // L'erreur est déjà gérée dans les méthodes userManager
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
        // L'erreur est déjà gérée dans les méthodes clientManager
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
        // L'erreur est déjà gérée dans les méthodes providerManager
    }
}

/**
 * Confirme la suppression d'un élément
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
        // L'erreur est déjà gérée dans les méthodes des managers
    }
}

/**
 * Charge les données initiales lors du chargement de la page
 */
function loadInitialData() {
    // Afficher un indicateur de chargement global
    const loadingIndicator = document.createElement('div');
    loadingIndicator.className = 'global-loading-indicator';
    loadingIndicator.innerHTML = '<div class="spinner"></div><span>Chargement des données...</span>';
    document.body.appendChild(loadingIndicator);

    // Charger toutes les données en parallèle, indépendamment de la section active
    Promise.all([
        userManager.loadUsers(),
        clientManager.loadClients(),
        providerManager.loadProviders(),
        userManager.loadBlockedUsers()
    ])
        .then(() => {
            console.log('✅ Toutes les données ont été chargées avec succès');

            // Masquer l'indicateur de chargement avec une transition
            loadingIndicator.classList.add('fade-out');
            setTimeout(() => {
                document.body.removeChild(loadingIndicator);
            }, 500);
        })
        .catch(error => {
            console.error('❌ Erreur lors du chargement des données:', error);

            // Transformer l'indicateur de chargement en message d'erreur
            loadingIndicator.innerHTML = `
            <div class="error-icon">❌</div>
            <span>Erreur lors du chargement des données. <button id="retry-load">Réessayer</button></span>
        `;
            loadingIndicator.className = 'global-error-indicator';

            // Ajouter un bouton pour réessayer
            document.getElementById('retry-load').addEventListener('click', () => {
                document.body.removeChild(loadingIndicator);
                loadInitialData();
            });
        });
}

/**
 * Initialisation de l'application
 */
document.addEventListener('DOMContentLoaded', function () {
    // Initialiser les gestionnaires d'événements
    initEventListeners();

    // S'abonner aux changements d'état
    initStateSubscriptions();

    // Charger les données initiales
    loadInitialData();
});