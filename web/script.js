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
});