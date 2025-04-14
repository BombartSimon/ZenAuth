/**
 * Module de gestion des utilisateurs
 * Gère toutes les opérations liées aux utilisateurs (chargement, ajout, édition, suppression)
 */
import { apiRequest, createJsonRequestOptions, API_ENDPOINTS, handleError } from '../utils/api.js';
import store from '../state/store.js';

// Temps minimum entre deux requêtes de chargement (en ms)
const MIN_REFRESH_INTERVAL = 1000;
let lastLoadTime = 0;

/**
 * Charge la liste des utilisateurs depuis l'API
 * Prend en compte la source (local ou externe)
 * @param {boolean} force - Force le rechargement même si appelé récemment
 * @returns {Promise<Array>} - Liste des utilisateurs
 */
export async function loadUsers(force = false) {
    try {
        // Éviter trop de requêtes rapprochées pour éviter de surcharger l'API
        const now = Date.now();
        if (!force && now - lastLoadTime < MIN_REFRESH_INTERVAL) {
            console.log('⏱️ Requête trop rapprochée, utilisation du cache');
            return store.getState('users').list;
        }

        lastLoadTime = now;
        const { source } = store.getState('users');
        const url = source === "external" ? API_ENDPOINTS.EXTERNAL_USERS : API_ENDPOINTS.USERS;

        // Mettre à jour l'état pour indiquer le chargement
        store.update('users', { loading: true, error: null });

        const users = await apiRequest(url);

        // Vérifier que les données sont bien un array
        if (!Array.isArray(users)) {
            throw new Error('Format de données utilisateurs invalide');
        }

        // Mettre à jour l'état avec les données
        store.update('users', { list: users, loading: false, lastUpdated: now });

        return users;
    } catch (error) {
        store.update('users', {
            loading: false,
            error: error.message || 'Erreur lors du chargement des utilisateurs'
        });

        handleError(error, 'Impossible de charger la liste des utilisateurs');
        return [];
    }
}

/**
 * Bascule entre les sources d'utilisateurs (local/externe)
 */
export function toggleUserSource() {
    const currentState = store.getState('users');
    const newSource = currentState.source === "external" ? "local" : "external";
    store.update('users', { source: newSource, loading: true });
    loadUsers(true); // Force le rechargement
}

/**
 * Crée un nouvel utilisateur
 * @param {Object} userData - Données de l'utilisateur à créer
 * @returns {Promise<Object>} - L'utilisateur créé
 */
export async function createUser(userData) {
    try {
        const options = createJsonRequestOptions('POST', userData);
        const result = await apiRequest(API_ENDPOINTS.USERS, options);

        // Recharger la liste pour intégrer le nouvel utilisateur
        await loadUsers(true);

        // Afficher un message de succès
        showSuccessMessage('Utilisateur créé avec succès');

        return result;
    } catch (error) {
        handleError(error, 'Impossible de créer l\'utilisateur');
        throw error;
    }
}

/**
 * Met à jour un utilisateur existant
 * @param {string} userId - ID de l'utilisateur
 * @param {Object} userData - Données à mettre à jour
 * @returns {Promise<Object>} - L'utilisateur mis à jour
 */
export async function updateUser(userId, userData) {
    try {
        const options = createJsonRequestOptions('PUT', userData);
        const result = await apiRequest(`${API_ENDPOINTS.USERS}/${userId}`, options);

        // Recharger la liste pour mettre à jour l'UI
        await loadUsers(true);

        // Afficher un message de succès
        showSuccessMessage('Utilisateur mis à jour avec succès');

        return result;
    } catch (error) {
        handleError(error, 'Impossible de mettre à jour l\'utilisateur');
        throw error;
    }
}

/**
 * Supprime un utilisateur
 * @param {string} userId - ID de l'utilisateur à supprimer
 * @returns {Promise<void>}
 */
export async function deleteUser(userId) {
    try {
        await apiRequest(`${API_ENDPOINTS.USERS}/${userId}`, { method: 'DELETE' });

        // Recharger la liste pour refléter la suppression
        await loadUsers(true);

        // Afficher un message de succès
        showSuccessMessage('Utilisateur supprimé avec succès');
    } catch (error) {
        handleError(error, 'Impossible de supprimer l\'utilisateur');
        throw error;
    }
}

/**
 * Charge la liste des utilisateurs bloqués
 * @param {boolean} force - Force le rechargement même si appelé récemment
 * @returns {Promise<Array>} - Liste des utilisateurs bloqués
 */
export async function loadBlockedUsers(force = false) {
    try {
        // Éviter trop de requêtes rapprochées
        const now = Date.now();
        if (!force && now - lastLoadTime < MIN_REFRESH_INTERVAL) {
            return store.getState('blockedUsers').list;
        }

        // Mettre à jour l'état pour indiquer le chargement
        store.update('blockedUsers', { loading: true, error: null });

        const blockedUsers = await apiRequest(API_ENDPOINTS.BLOCKED_USERS);

        // Vérifier que les données sont bien un array
        if (!Array.isArray(blockedUsers)) {
            throw new Error('Format de données d\'utilisateurs bloqués invalide');
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

        store.update('blockedUsers', {
            list: blockedUsers,
            groupedData: groupedData,
            loading: false,
            lastUpdated: now
        });

        return blockedUsers;
    } catch (error) {
        store.update('blockedUsers', {
            loading: false,
            error: error.message || 'Erreur lors du chargement des utilisateurs bloqués'
        });

        handleError(error, 'Impossible de charger les utilisateurs bloqués');
        return [];
    }
}

/**
 * Débloque un utilisateur ou une adresse IP
 * @param {string} identifier - Identifiant à débloquer
 * @param {string} type - Type d'identifiant ('user' ou 'ip')
 * @returns {Promise<Object>} - Résultat de l'opération
 */
export async function unblockUser(identifier, type) {
    try {
        const options = createJsonRequestOptions('POST', { identifier, type });
        const result = await apiRequest(API_ENDPOINTS.UNBLOCK_USER, options);

        // Recharger la liste pour refléter le déblocage
        await loadBlockedUsers(true);

        // Afficher un message de succès
        showSuccessMessage(`${type === 'user' ? 'Utilisateur' : 'Adresse IP'} débloqué avec succès`);

        return result;
    } catch (error) {
        handleError(error, 'Impossible de débloquer cet identifiant');
        throw error;
    }
}

/**
 * Affiche une notification de succès
 * @param {string} message - Message à afficher
 */
function showSuccessMessage(message) {
    const notification = document.createElement('div');
    notification.className = 'success-notification';
    notification.innerHTML = `
        <div class="success-icon">✓</div>
        <div class="success-message">${message}</div>
        <button class="success-close">×</button>
    `;

    document.body.appendChild(notification);

    // Animation d'entrée
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);

    // Ajouter un gestionnaire pour fermer la notification
    notification.querySelector('.success-close').addEventListener('click', () => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    });

    // Auto-fermeture après 3 secondes
    setTimeout(() => {
        if (document.body.contains(notification)) {
            notification.classList.remove('show');
            setTimeout(() => {
                if (document.body.contains(notification)) {
                    document.body.removeChild(notification);
                }
            }, 300);
        }
    }, 3000);
}