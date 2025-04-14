/**
 * Module de gestion de l'interface utilisateur
 * Centralise les opérations de manipulation du DOM et la gestion des modales
 */
import store from '../state/store.js';

/**
 * Initialise la gestion des thèmes (clair/sombre)
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

        // Mettre à jour l'état
        store.update('ui', { theme });
    });
}

/**
 * Initialise la navigation entre les sections
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

            // Mettre à jour l'état
            store.update('ui', { activeSection: targetSection });
        });
    });
}

/**
 * Gère l'affichage d'une modale générique
 * @param {HTMLElement} modal - Élément DOM de la modale
 * @param {boolean} visible - true pour afficher, false pour masquer
 */
export function toggleModal(modal, visible) {
    if (modal) {
        modal.style.display = visible ? 'block' : 'none';
    }
}

/**
 * Configure les événements de fermeture pour les modales au clic extérieur
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

                // Mettre à jour le state UI en conséquence
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
 * Affiche une modale de confirmation de suppression
 * @param {string} id - Identifiant de l'élément à supprimer
 * @param {string} type - Type d'élément ('user', 'client', 'provider')
 * @param {string} name - Nom de l'élément
 */
export function showDeleteConfirmation(id, type, name) {
    const deleteModal = document.getElementById('delete-modal');
    const deleteConfirmationMessage = document.getElementById('delete-confirmation-message');

    deleteConfirmationMessage.textContent = `Êtes-vous sûr de vouloir supprimer ${type} "${name}"?`;
    toggleModal(deleteModal, true);

    store.update('ui', {
        modals: {
            ...store.getState('ui').modals,
            delete: { isOpen: true, id, type, name }
        }
    });
}

/**
 * Rendu d'une liste d'utilisateurs dans le DOM
 * @param {Array} users - Liste des utilisateurs à afficher
 * @param {HTMLElement} container - Élément DOM conteneur
 */
export function renderUsersList(users, container) {
    if (!container) return;

    const { source } = store.getState('users');

    // Créer un bouton de bascule en haut de la table
    const sourceText = source === "external" ? "Voir utilisateurs locaux" : "Voir utilisateurs externes";

    let html = `
    <tr>
        <td colspan="3" style="text-align: center;">
            <button id="toggle-source" class="btn primary">
                ${sourceText}
            </button>
            <div style="margin-top: 10px;">
                <small>Source actuelle: <strong>${source === "external" ? "Base de données externe" : "ZenAuth local"}</strong></small>
            </div>
        </td>
    </tr>`;

    // Ajouter les utilisateurs
    users.forEach(user => {
        // Déterminer si nous devons afficher les actions d'édition/suppression
        const actionsCell = source === "external"
            ? `<td><span class="badge">Externe</span></td>`
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
 * Rendu d'une liste de clients OAuth dans le DOM
 * @param {Array} clients - Liste des clients à afficher  
 * @param {HTMLElement} container - Élément DOM conteneur
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
 * Rendu d'une liste de fournisseurs d'authentification dans le DOM
 * @param {Array} providers - Liste des fournisseurs à afficher
 * @param {HTMLElement} container - Élément DOM conteneur
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
 * Rendu de la liste des utilisateurs bloqués
 * @param {Object} groupedData - Données regroupées des utilisateurs bloqués
 * @param {HTMLElement} container - Élément DOM conteneur
 */
export function renderBlockedUsersList(groupedData, container) {
    if (!container) return;

    container.innerHTML = '';

    if (Object.keys(groupedData).length === 0) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td colspan="4" class="empty-state">
                Aucun utilisateur ou adresse IP bloqué
            </td>
        `;
        container.appendChild(row);
        return;
    }

    // Afficher les données groupées
    Object.values(groupedData).forEach(entry => {
        const row = document.createElement('tr');

        // Affichage différent selon le type
        if (entry.type === "user") {
            row.innerHTML = `
                <td>
                    <strong>${entry.identifier}</strong>
                    ${entry.associatedIPs && entry.associatedIPs.length > 0 ?
                    `<div class="associated-data">IPs associées: ${entry.associatedIPs.join(", ")}</div>` :
                    ''}
                </td>
                <td><span class="badge user">user</span></td>
                <td>${entry.blockedFor}</td>
                <td>
                    <button class="action-btn unblock" 
                            data-identifier="${entry.identifier}" 
                            data-type="user">
                        <i class="fas fa-unlock"></i> Débloquer Utilisateur & IPs
                    </button>
                </td>
            `;
        } else {
            row.innerHTML = `
                <td>
                    <strong>${entry.identifier}</strong>
                    ${entry.associatedUsers && entry.associatedUsers.length > 0 ?
                    `<div class="associated-data">Utilisateurs associés: ${entry.associatedUsers.join(", ")}</div>` :
                    ''}
                </td>
                <td><span class="badge ip">ip</span></td>
                <td>${entry.blockedFor}</td>
                <td>
                    <button class="action-btn unblock" 
                            data-identifier="${entry.identifier}" 
                            data-type="ip">
                        <i class="fas fa-unlock"></i> Débloquer IP & Utilisateurs
                    </button>
                </td>
            `;
        }

        container.appendChild(row);
    });
}