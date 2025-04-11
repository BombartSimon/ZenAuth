document.addEventListener('DOMContentLoaded', function () {
    // Theme switching
    const themeToggle = document.getElementById('theme-toggle');
    const savedTheme = localStorage.getItem('theme') || 'light';

    // Set initial theme from localStorage
    document.documentElement.setAttribute('data-theme', savedTheme);
    themeToggle.checked = savedTheme === 'dark';

    themeToggle.addEventListener('change', function () {
        const theme = this.checked ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
    });

    // Navigation
    const menuLinks = document.querySelectorAll('.menu-link');
    const sections = document.querySelectorAll('.section');

    let userSource = "local";

    menuLinks.forEach(link => {
        link.addEventListener('click', function (e) {
            e.preventDefault();
            const targetSection = this.getAttribute('data-section');

            // Update active menu link
            menuLinks.forEach(menuLink => {
                menuLink.classList.remove('active');
            });
            this.classList.add('active');

            // Show the target section, hide others
            sections.forEach(section => {
                section.classList.remove('active');
                if (section.id === `${targetSection}-section`) {
                    section.classList.add('active');
                }
            });

            // Load appropriate data
            if (targetSection === 'users') {
                loadUsers();
            } else if (targetSection === 'clients') {
                loadClients();
            } else if (targetSection === 'providers') {
                loadProviders();
            }
        });
    });

    // User Management Elements
    const usersList = document.getElementById('users-list');
    const addUserBtn = document.getElementById('add-user-btn');
    const userModal = document.getElementById('user-modal');
    const modalTitle = document.getElementById('modal-title');
    const userForm = document.getElementById('user-form');
    const userId = document.getElementById('user-id');
    const username = document.getElementById('username');
    const password = document.getElementById('password');
    const passwordHint = document.getElementById('password-hint');
    const cancelBtn = document.getElementById('cancel-btn');
    const closeBtn = userModal.querySelector('.close');

    // Client Management Elements
    const clientsList = document.getElementById('clients-list');
    const addClientBtn = document.getElementById('add-client-btn');
    const clientModal = document.getElementById('client-modal');
    const clientModalTitle = document.getElementById('client-modal-title');
    const clientForm = document.getElementById('client-form');
    const clientIdInput = document.getElementById('client-id-input');
    const clientId = document.getElementById('client-id');
    const clientName = document.getElementById('client-name');
    const clientSecret = document.getElementById('client-secret');
    const redirectUris = document.getElementById('redirect-uris');
    const secretHint = document.getElementById('secret-hint');
    const cancelClientBtn = document.getElementById('cancel-client-btn');
    const closeClientBtn = clientModal.querySelector('.close');

    // Provider Management Elements
    const providersSection = document.getElementById('providers-section');
    const providersList = document.getElementById('providers-list');
    const addProviderBtn = document.getElementById('add-provider-btn');
    const providerModal = document.getElementById('provider-modal');
    const providerModalTitle = document.getElementById('provider-modal-title');
    const providerForm = document.getElementById('provider-form');
    const providerId = document.getElementById('provider-id');
    const providerName = document.getElementById('provider-name');
    const providerType = document.getElementById('provider-type');
    const providerTenantId = document.getElementById('provider-tenant-id');
    const tenantIdGroup = document.querySelector('.tenant-id-group');

    const providerClientId = document.getElementById('provider-client-id');
    const providerClientSecret = document.getElementById('provider-client-secret');
    const providerEnabled = document.getElementById('provider-enabled');
    const secretProviderHint = document.getElementById('provider-secret-hint');
    const cancelProviderBtn = document.getElementById('cancel-provider-btn');
    const closeProviderBtn = providerModal.querySelector('.close');


    // Blocked Users Section
    const blockedUsersSection = document.getElementById('blocked-users-section');
    const blockedUsersList = document.getElementById('blocked-users-list');
    const refreshBlockedBtn = document.getElementById('refresh-blocked-btn');

    document.querySelector('.menu-link[data-section="blocked-users"]').addEventListener('click', () => {
        loadBlockedUsers();
    });

    refreshBlockedBtn.addEventListener('click', loadBlockedUsers);





    // Delete Modal Elements
    const deleteModal = document.getElementById('delete-modal');
    const confirmDeleteBtn = document.getElementById('confirm-delete');
    const cancelDeleteBtn = document.getElementById('cancel-delete');
    const deleteConfirmationMessage = document.getElementById('delete-confirmation-message');

    let deleteItemId = null;
    let deleteItemType = null;

    // Load initial data
    loadUsers();

    // User Event Listeners
    addUserBtn.addEventListener('click', showAddUserModal);
    userForm.addEventListener('submit', handleUserFormSubmit);
    cancelBtn.addEventListener('click', closeUserModal);
    closeBtn.addEventListener('click', closeUserModal);

    // Client Event Listeners
    addClientBtn.addEventListener('click', showAddClientModal);
    clientForm.addEventListener('submit', handleClientFormSubmit);
    cancelClientBtn.addEventListener('click', closeClientModal);
    closeClientBtn.addEventListener('click', closeClientModal);

    // Provider Event Listeners
    addProviderBtn.addEventListener('click', showAddProviderModal);
    providerForm.addEventListener('submit', handleProviderFormSubmit);
    cancelProviderBtn.addEventListener('click', closeProviderModal);
    closeProviderBtn.addEventListener('click', closeProviderModal);

    // Delete Event Listeners
    confirmDeleteBtn.addEventListener('click', confirmDelete);
    cancelDeleteBtn.addEventListener('click', closeDeleteModal);

    // When clicking outside modals, close them
    window.addEventListener('click', function (event) {
        if (event.target === userModal) {
            closeUserModal();
        }
        if (event.target === clientModal) {
            closeClientModal();
        }
        if (event.target === deleteModal) {
            closeDeleteModal();
        }
    });


    // User Management Functions
    async function loadUsers() {
        try {
            const url = userSource === "external" ?
                '/admin/users?provider=external' :
                '/admin/users';

            const response = await fetch(url);
            if (!response.ok) throw new Error('Failed to load users');
            const users = await response.json();

            // Créer un bouton de bascule en haut de la table
            const sourceText = userSource === "external" ? "Voir utilisateurs locaux" : "Voir utilisateurs externes";

            usersList.innerHTML = `
            <tr>
                <td colspan="3" style="text-align: center;">
                    <button id="toggle-source" class="btn primary">
                        ${sourceText}
                    </button>
                    <div style="margin-top: 10px;">
                        <small>Source actuelle: <strong>${userSource === "external" ? "Base de données externe" : "ZenAuth local"}</strong></small>
                    </div>
                </td>
            </tr>`;

            // Ajouter les utilisateurs après le bouton de bascule
            users.forEach(user => {
                const row = document.createElement('tr');

                // Déterminer si nous devons afficher les actions d'édition/suppression
                const actionsCell = userSource === "external"
                    ? `<td><span class="badge">Externe</span></td>`
                    : `<td>
                        <button class="action-btn edit" data-id="${user.id}" data-username="${user.username}">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-btn delete" data-id="${user.id}" data-type="user" data-name="${user.username}">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                      </td>`;

                row.innerHTML = `
                    <td>${user.id}</td>
                    <td>${user.username}${user.email ? ` (${user.email})` : ''}</td>
                    ${actionsCell}
                `;
                usersList.appendChild(row);
            });

            // Ajouter l'événement de bascule
            document.getElementById('toggle-source').addEventListener('click', () => {
                userSource = userSource === "external" ? "local" : "external";
                loadUsers();
            });




            // Attacher les gestionnaires d'événements aux boutons d'édition et de suppression
            document.querySelectorAll('#users-list .action-btn.edit').forEach(btn => {
                btn.addEventListener('click', () => {
                    showEditUserModal(btn.dataset.id, btn.dataset.username);
                });
            });

            if (userSource === "local") {
                // Attacher les gestionnaires d'événements aux boutons d'édition et de suppression
                document.querySelectorAll('#users-list .action-btn.edit').forEach(btn => {
                    btn.addEventListener('click', () => {
                        showEditUserModal(btn.dataset.id, btn.dataset.username);
                    });
                });

                document.querySelectorAll('#users-list .action-btn.delete[data-type="user"]').forEach(btn => {
                    btn.addEventListener('click', () => {
                        showDeleteConfirmation(btn.dataset.id, 'user', btn.dataset.name);
                    });
                });
            }

            document.querySelectorAll('#users-list .action-btn.delete[data-type="user"]').forEach(btn => {
                btn.addEventListener('click', () => {
                    showDeleteConfirmation(btn.dataset.id, 'user', btn.dataset.name);
                });
            });
        } catch (error) {
            console.error('Error loading users:', error);
            alert('Failed to load users. Please refresh the page.');
        }
    }

    async function loadBlockedUsers() {
        try {
            const response = await fetch('/admin/blocked-users');
            if (!response.ok) throw new Error('Failed to load blocked users');
            const blockedUsers = await response.json();

            blockedUsersList.innerHTML = '';

            if (blockedUsers.length === 0) {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td colspan="4" class="empty-state">
                        No blocked users or IP addresses
                    </td>
                `;
                blockedUsersList.appendChild(row);
                return;
            }

            // Organiser les données pour regrouper les utilisateurs et leurs IPs
            const groupedData = {};

            blockedUsers.forEach(user => {
                // Si c'est un utilisateur, on le traite séparément
                if (user.type === "user") {
                    const key = user.identifier;
                    if (!groupedData[key]) {
                        groupedData[key] = {
                            identifier: user.identifier,
                            type: "user",
                            blockedFor: user.blocked_for,
                            associatedIPs: []
                        };
                    }
                }
                // Si c'est une IP, on cherche si elle est associée à un utilisateur
                else if (user.type === "ip") {
                    let found = false;

                    // Rechercher dans les associations existantes
                    Object.values(groupedData).forEach(entry => {
                        if (entry.associatedIPs && entry.associatedIPs.includes(user.identifier)) {
                            found = true;
                        }
                    });

                    // Si cette IP n'est pas déjà associée, créer une nouvelle entrée
                    if (!found) {
                        groupedData[`ip_${user.identifier}`] = {
                            identifier: user.identifier,
                            type: "ip",
                            blockedFor: user.blocked_for,
                            associatedUsers: []
                        };
                    }
                }
            });

            // Afficher les données groupées
            Object.values(groupedData).forEach(entry => {
                const row = document.createElement('tr');

                // Affichage différent selon le type
                if (entry.type === "user") {
                    row.innerHTML = `
                        <td>
                            <strong>${entry.identifier}</strong>
                            ${entry.associatedIPs.length > 0 ?
                            `<div class="associated-data">Associated IPs: ${entry.associatedIPs.join(", ")}</div>` :
                            ''}
                        </td>
                        <td><span class="badge user">user</span></td>
                        <td>${entry.blockedFor}</td>
                        <td>
                            <button class="action-btn unblock" 
                                    data-identifier="${entry.identifier}" 
                                    data-type="user">
                                <i class="fas fa-unlock"></i> Unblock User & IPs
                            </button>
                        </td>
                    `;
                } else {
                    row.innerHTML = `
                        <td>
                            <strong>${entry.identifier}</strong>
                            ${entry.associatedUsers.length > 0 ?
                            `<div class="associated-data">Associated users: ${entry.associatedUsers.join(", ")}</div>` :
                            ''}
                        </td>
                        <td><span class="badge ip">ip</span></td>
                        <td>${entry.blockedFor}</td>
                        <td>
                            <button class="action-btn unblock" 
                                    data-identifier="${entry.identifier}" 
                                    data-type="ip">
                                <i class="fas fa-unlock"></i> Unblock IP & Users
                            </button>
                        </td>
                    `;
                }

                blockedUsersList.appendChild(row);
            });

            // Ajouter les gestionnaires d'événements pour les boutons de déblocage
            document.querySelectorAll('#blocked-users-list .action-btn.unblock').forEach(btn => {
                btn.addEventListener('click', async () => {
                    try {
                        const response = await fetch('/admin/unblock-user', {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json'
                            },
                            body: JSON.stringify({
                                identifier: btn.dataset.identifier,
                                type: btn.dataset.type
                            })
                        });

                        if (response.ok) {
                            const result = await response.json();
                            // Afficher un message de confirmation
                            alert(result.message);
                            // Recharger la liste après le déblocage
                            loadBlockedUsers();
                        } else {
                            const error = await response.text();
                            throw new Error(error);
                        }
                    } catch (error) {
                        console.error('Error unblocking:', error);
                        alert(`Failed to unblock: ${error.message}`);
                    }
                });
            });
        } catch (error) {
            console.error('Error loading blocked users:', error);
            alert('Failed to load blocked users. Please refresh the page.');
        }
    }

    async function loadClients() {
        try {
            const response = await fetch('/admin/clients');
            if (!response.ok) throw new Error('Failed to load clients');
            const clients = await response.json();

            clientsList.innerHTML = '';
            clients.forEach(client => {

                const redirectUrisText = client.RedirectURIs.join(', ');
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${client.ID}</td>
                    <td>${client.Name}</td>
                    <td>${redirectUrisText}</td>
                    <td>
                        <button class="action-btn edit" data-id="${client.ID}" data-name="${client.Name}" 
                                data-secret="${client.Secret}" data-uris="${client.RedirectURIs.join('\n')}">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-btn delete" data-id="${client.ID}" data-type="client" data-name="${client.Name}">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </td>
                `;
                clientsList.appendChild(row);
            });

            // Add event listeners to the client buttons
            document.querySelectorAll('.action-btn.edit[data-id]').forEach(btn => {
                btn.addEventListener('click', () => {
                    if (btn.closest('#clients-table')) {
                        showEditClientModal(btn.dataset.id, btn.dataset.name, btn.dataset.secret, btn.dataset.uris);
                    }
                });
            });

            document.querySelectorAll('.action-btn.delete[data-type="client"]').forEach(btn => {
                btn.addEventListener('click', () => {
                    showDeleteConfirmation(btn.dataset.id, 'client', btn.dataset.name);
                });
            });
        } catch (error) {
            console.error('Error loading clients:', error);
            alert('Failed to load clients. Please refresh the page.');
        }
    }

    // Auth Provider Management Functions
    async function loadProviders() {
        try {
            const response = await fetch('/admin/auth-providers');
            if (!response.ok) throw new Error('Failed to load providers');
            const providers = await response.json();

            providersList.innerHTML = '';

            providers.forEach(provider => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${provider.id}</td>
                    <td>${provider.name}</td>
                    <td><span class="badge provider-type ${provider.type}">${provider.type}</span></td>
                    <td>${provider.client_id}</td>
                    <td>
                        <span class="status-indicator ${provider.enabled ? 'enabled' : 'disabled'}">
                            ${provider.enabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </td>
                    <td>
                        <button class="action-btn edit" data-id="${provider.id}" 
                            data-name="${provider.name}" data-type="${provider.type}"
                            data-client-id="${provider.client_id}" data-tenant-id="${provider.tenant_id || ''}"
                            data-enabled="${provider.enabled}">
                            <i class="fas fa-edit"></i> Edit
                        </button>
                        <button class="action-btn delete" data-id="${provider.id}" data-type="provider" data-name="${provider.name}">
                            <i class="fas fa-trash"></i> Delete
                        </button>
                    </td>
                `;
                providersList.appendChild(row);
            });

            // Attach the event handlers
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

            document.querySelectorAll('#providers-list .action-btn.delete').forEach(btn => {
                btn.addEventListener('click', () => {
                    showDeleteConfirmation(btn.dataset.id, 'provider', btn.dataset.name);
                });
            });
        } catch (error) {
            console.error('Error loading providers:', error);
            alert('Failed to load authentication providers. Please refresh the page.');
        }
    }


    // User Modal Functions
    function showAddUserModal() {
        modalTitle.textContent = 'Add User';
        userId.value = '';
        username.value = '';
        password.value = '';
        password.required = true;
        passwordHint.style.display = 'none';
        userModal.style.display = 'block';
    }

    function showEditUserModal(id, usernameValue) {
        modalTitle.textContent = 'Edit User';
        userId.value = id;
        username.value = usernameValue;
        password.value = '';
        password.required = false;
        passwordHint.style.display = 'block';
        userModal.style.display = 'block';
    }

    function closeUserModal() {
        userModal.style.display = 'none';
    }

    // Client Modal Functions
    function showAddClientModal() {
        clientModalTitle.textContent = 'Add Client';
        clientIdInput.value = '';
        clientId.value = '';
        clientId.readOnly = false;
        clientName.value = '';
        clientSecret.value = '';
        clientSecret.required = true;
        secretHint.style.display = 'none';
        redirectUris.value = '';
        clientModal.style.display = 'block';
    }

    function showEditClientModal(id, name, secret, uris) {
        clientModalTitle.textContent = 'Edit Client';
        clientIdInput.value = id;
        clientId.value = id;
        clientId.readOnly = true;
        clientName.value = name;
        clientSecret.value = '';
        clientSecret.required = false;
        secretHint.style.display = 'block';
        redirectUris.value = uris;
        clientModal.style.display = 'block';
    }

    function closeClientModal() {
        clientModal.style.display = 'none';
    }

    // Provider Modal Functions
    // Update showAddProviderModal
    function showAddProviderModal() {
        providerModalTitle.textContent = 'Add Authentication Provider';
        providerId.value = '';
        providerName.value = '';
        providerType.value = 'google';
        providerClientId.value = '';
        providerClientSecret.value = '';
        providerClientSecret.required = true;
        secretProviderHint.style.display = 'none';
        providerTenantId.value = '';
        providerEnabled.checked = false;

        // Show/hide tenant ID field based on provider type
        toggleTenantIdField(providerType.value);
        providerModal.style.display = 'block';
    }

    function showEditProviderModal(id, name, type, clientId, tenantId, enabled) {
        providerModalTitle.textContent = 'Edit Authentication Provider';
        providerId.value = id;
        providerName.value = name;
        providerType.value = type;
        providerClientId.value = clientId;
        providerClientSecret.value = '';
        providerClientSecret.required = false;
        secretProviderHint.style.display = 'block';
        providerTenantId.value = tenantId || '';
        providerEnabled.checked = enabled;

        // Show/hide tenant ID field based on provider type
        toggleTenantIdField(providerType.value);
        providerModal.style.display = 'block';
    }
    function closeProviderModal() {
        providerModal.style.display = 'none';
    }

    providerType.addEventListener('change', function () {
        toggleTenantIdField(this.value);
    });

    function toggleTenantIdField(providerType) {
        if (providerType === 'microsoft') {
            tenantIdGroup.style.display = 'block';
        } else {
            tenantIdGroup.style.display = 'none';
            providerTenantId.value = '';
        }
    }

    // Delete Modal Functions
    function showDeleteConfirmation(id, type, name) {
        deleteItemId = id;
        deleteItemType = type;
        deleteConfirmationMessage.textContent = `Are you sure you want to delete ${type} "${name}"?`;
        deleteModal.style.display = 'block';
    }

    function closeDeleteModal() {
        deleteModal.style.display = 'none';
    }

    // Form Submit Handlers
    async function handleUserFormSubmit(e) {
        e.preventDefault();

        const isEditing = userId.value !== '';
        const userData = {
            username: username.value
        };

        // Only include password if provided (for editing) or required (for adding)
        if (password.value) {
            userData.password = password.value;
        }

        try {
            let response;

            if (isEditing) {
                // Update existing user
                response = await fetch(`/admin/users/${userId.value}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(userData)
                });
            } else {
                // Create new user
                response = await fetch('/admin/users', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(userData)
                });
            }

            if (response.ok) {
                closeUserModal();
                loadUsers();
            } else {
                const error = await response.text();
                throw new Error(error);
            }
        } catch (error) {
            console.error('Error saving user:', error);
            alert(`Failed to save user: ${error.message}`);
        }
    }

    async function handleClientFormSubmit(e) {
        e.preventDefault();

        const isEditing = clientIdInput.value !== '';
        const redirectUrisList = redirectUris.value
            .split('\n')
            .map(uri => uri.trim())
            .filter(uri => uri !== '');

        const clientData = {
            id: clientId.value,
            name: clientName.value,
            redirect_uris: redirectUrisList
        };

        // Only include secret if provided (for editing) or required (for adding)
        if (clientSecret.value) {
            clientData.secret = clientSecret.value;
        }

        try {
            let response;

            if (isEditing) {
                // Update existing client
                response = await fetch(`/admin/clients/${clientIdInput.value}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(clientData)
                });
            } else {
                // Create new client
                response = await fetch('/admin/clients', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(clientData)
                });
            }

            if (response.ok) {
                closeClientModal();
                loadClients();
            } else {
                const error = await response.text();
                throw new Error(error);
            }
        } catch (error) {
            console.error('Error saving client:', error);
            alert(`Failed to save client: ${error.message}`);
        }
    }

    // Form Submit Handler for Providers
    async function handleProviderFormSubmit(e) {
        e.preventDefault();

        const isEditing = providerId.value !== '';
        const providerData = {
            name: providerName.value,
            type: providerType.value,
            client_id: providerClientId.value,
            enabled: providerEnabled.checked,
            tenant_id: providerTenantId.value
        };

        // Only include secret if provided (for editing) or required (for adding)
        if (providerClientSecret.value) {
            providerData.client_secret = providerClientSecret.value;
        }

        try {
            let response;

            if (isEditing) {
                // Update existing provider
                response = await fetch(`/admin/auth-providers/${providerId.value}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(providerData)
                });
            } else {
                // Create new provider
                response = await fetch('/admin/auth-providers', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(providerData)
                });
            }

            if (response.ok) {
                closeProviderModal();
                loadProviders();
            } else {
                const error = await response.text();
                throw new Error(error);
            }
        } catch (error) {
            console.error('Error saving provider:', error);
            alert(`Failed to save provider: ${error.message}`);
        }
    }


    async function confirmDelete() {
        try {
            let response;

            if (deleteItemType === 'user') {
                response = await fetch(`/admin/users/${deleteItemId}`, {
                    method: 'DELETE'
                });
            } else if (deleteItemType === 'client') {
                response = await fetch(`/admin/clients/${deleteItemId}`, {
                    method: 'DELETE'
                });
            }

            if (response.ok) {
                closeDeleteModal();
                if (deleteItemType === 'user') {
                    loadUsers();
                } else if (deleteItemType === 'client') {
                    loadClients();
                }
            } else {
                const error = await response.text();
                throw new Error(error);
            }
        } catch (error) {
            console.error(`Error deleting ${deleteItemType}:`, error);
            alert(`Failed to delete ${deleteItemType}: ${error.message}`);
        }
    }
    const originalConfirmDelete = confirmDelete;
    confirmDelete = async function () {
        if (deleteItemType === 'provider') {
            try {
                const response = await fetch(`/admin/auth-providers/${deleteItemId}`, {
                    method: 'DELETE'
                });

                if (response.ok) {
                    closeDeleteModal();
                    loadProviders();
                } else {
                    throw new Error('Failed to delete provider');
                }
            } catch (error) {
                console.error('Error:', error);
                alert('Failed to delete provider');
            }
        } else {
            // Call the original implementation for other item types
            originalConfirmDelete();
        }
    };
});