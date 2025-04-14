/**
 * User interface management module
 * Centralizes DOM manipulation operations and modal management
 */
import store from '../state/store.js';

/**
 * Initializes theme management (light/dark)
 */
export function initThemeToggle() {
    const themeToggle = document.getElementById('theme-toggle');
    const savedTheme = localStorage.getItem('theme') || 'light';

    // Set initial theme from localStorage
    document.documentElement.setAttribute('data-theme', savedTheme);
    themeToggle.checked = savedTheme === 'dark';

    themeToggle.addEventListener('change', function () {
        const theme = this.checked ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);

        // Update state
        store.update('ui', { theme });
    });
}

/**
 * Initializes navigation between sections
 */
export function initNavigation() {
    const menuLinks = document.querySelectorAll('.menu-link');
    const sections = document.querySelectorAll('.section');

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

            // Update state
            store.update('ui', { activeSection: targetSection });
        });
    });
}

/**
 * Manages the display of a generic modal
 * @param {HTMLElement} modal - Modal DOM element
 * @param {boolean} visible - true to show, false to hide
 */
export function toggleModal(modal, visible) {
    if (modal) {
        modal.style.display = visible ? 'block' : 'none';
    }
}

/**
 * Configures close events for modals on outside click
 */
export function setupModalOutsideClicks() {
    const modals = [
        document.getElementById('user-modal'),
        document.getElementById('client-modal'),
        document.getElementById('provider-modal'),
        document.getElementById('delete-modal')
    ];

    window.addEventListener('click', function (event) {
        modals.forEach(modal => {
            if (event.target === modal) {
                toggleModal(modal, false);

                // Update UI state accordingly
                if (modal.id === 'user-modal') {
                    store.update('ui', {
                        modals: {
                            ...store.getState('ui').modals,
                            user: { isOpen: false }
                        }
                    });
                } else if (modal.id === 'client-modal') {
                    store.update('ui', {
                        modals: {
                            ...store.getState('ui').modals,
                            client: { isOpen: false }
                        }
                    });
                } else if (modal.id === 'provider-modal') {
                    store.update('ui', {
                        modals: {
                            ...store.getState('ui').modals,
                            provider: { isOpen: false }
                        }
                    });
                } else if (modal.id === 'delete-modal') {
                    store.update('ui', {
                        modals: {
                            ...store.getState('ui').modals,
                            delete: { isOpen: false }
                        }
                    });
                }
            }
        });
    });
}

/**
 * Displays a delete confirmation modal
 * @param {string} id - ID of the item to delete
 * @param {string} type - Type of item ('user', 'client', 'provider')
 * @param {string} name - Name of the item
 */
export function showDeleteConfirmation(id, type, name) {
    const deleteModal = document.getElementById('delete-modal');
    const deleteConfirmationMessage = document.getElementById('delete-confirmation-message');

    deleteConfirmationMessage.textContent = `Are you sure you want to delete ${type} "${name}"?`;
    toggleModal(deleteModal, true);

    store.update('ui', {
        modals: {
            ...store.getState('ui').modals,
            delete: { isOpen: true, id, type, name }
        }
    });
}

/**
 * Renders a list of users in the DOM
 * @param {Array} users - List of users to display
 * @param {HTMLElement} container - Container DOM element
 */
export function renderUsersList(users, container) {
    if (!container) return;

    const { source } = store.getState('users');

    // Create a toggle button at the top of the table
    const sourceText = source === "external" ? "View local users" : "View external users";

    let html = `
    <tr>
        <td colspan="3" style="text-align: center;">
            <button id="toggle-source" class="btn primary">
                ${sourceText}
            </button>
            <div style="margin-top: 10px;">
                <small>Current source: <strong>${source === "external" ? "External database" : "ZenAuth local"}</strong></small>
            </div>
        </td>
    </tr>`;

    // Add users
    users.forEach(user => {
        // Determine whether to display edit/delete actions
        const actionsCell = source === "external"
            ? `<td><span class="badge">External</span></td>`
            : `<td>
                <button class="action-btn edit" data-id="${user.id}" data-username="${user.username}">
                    <i class="fas fa-edit"></i> Edit
                </button>
                <button class="action-btn delete" data-id="${user.id}" data-type="user" data-name="${user.username}">
                    <i class="fas fa-trash"></i> Delete
                </button>
              </td>`;

        html += `
            <tr>
                <td>${user.id}</td>
                <td>${user.username}${user.email ? ` (${user.email})` : ''}</td>
                ${actionsCell}
            </tr>
        `;
    });

    container.innerHTML = html;
}

/**
 * Renders a list of OAuth clients in the DOM
 * @param {Array} clients - List of clients to display
 * @param {HTMLElement} container - Container DOM element
 */
export function renderClientsList(clients, container) {
    if (!container) return;

    container.innerHTML = '';

    clients.forEach(client => {
        const redirectUrisText = Array.isArray(client.RedirectURIs) ? client.RedirectURIs.join(', ') : '';
        const row = document.createElement('tr');

        row.innerHTML = `
            <td>${client.ID}</td>
            <td>${client.Name}</td>
            <td>${redirectUrisText}</td>
            <td>
                <button class="action-btn edit" data-id="${client.ID}" data-name="${client.Name}" 
                        data-secret="${client.Secret}" data-uris="${Array.isArray(client.RedirectURIs) ? client.RedirectURIs.join('\n') : ''}">
                    <i class="fas fa-edit"></i> Edit
                </button>
                <button class="action-btn delete" data-id="${client.ID}" data-type="client" data-name="${client.Name}">
                    <i class="fas fa-trash"></i> Delete
                </button>
            </td>
        `;

        container.appendChild(row);
    });
}

/**
 * Renders a list of authentication providers in the DOM
 * @param {Array} providers - List of providers to display
 * @param {HTMLElement} container - Container DOM element
 */
export function renderProvidersList(providers, container) {
    if (!container) return;

    container.innerHTML = '';

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

        container.appendChild(row);
    });
}

/**
 * Renders the list of blocked users
 * @param {Object} groupedData - Grouped data of blocked users
 * @param {HTMLElement} container - Container DOM element
 */
export function renderBlockedUsersList(groupedData, container) {
    if (!container) return;

    container.innerHTML = '';

    if (Object.keys(groupedData).length === 0) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td colspan="4" class="empty-state">
                No blocked users or IP addresses
            </td>
        `;
        container.appendChild(row);
        return;
    }

    // Display grouped data
    Object.values(groupedData).forEach(entry => {
        const row = document.createElement('tr');

        // Different display based on type
        if (entry.type === "user") {
            row.innerHTML = `
                <td>
                    <strong>${entry.identifier}</strong>
                    ${entry.associatedIPs && entry.associatedIPs.length > 0 ?
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
                    ${entry.associatedUsers && entry.associatedUsers.length > 0 ?
                    `<div class="associated-data">Associated Users: ${entry.associatedUsers.join(", ")}</div>` :
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

        container.appendChild(row);
    });
}