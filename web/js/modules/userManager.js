/**
 * Module de gestion des utilisateurs
 * Gère toutes les opérations liées aux utilisateurs (chargement, ajout, édition, suppression)
 */
import { apiRequest, createJsonRequestOptions, API_ENDPOINTS, handleError } from '../utils/api.js';
import store from '../state/store.js';

/**
 * Charge la liste des utilisateurs depuis l'API
 * Prend en compte la source (local ou externe)
 */
export async function loadUsers() {
    try {
        const { source } = store.getState('users');
        const url = source === "external" ? API_ENDPOINTS.EXTERNAL_USERS : API_ENDPOINTS.USERS;

        const users = await apiRequest(url);

        store.update('users', { list: users });
        return users;
    } catch (error) {
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
    store.update('users', { source: newSource });
    loadUsers();
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
        await loadUsers();
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
        await loadUsers();
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
        await loadUsers();
    } catch (error) {
        handleError(error, 'Impossible de supprimer l\'utilisateur');
        throw error;
    }
}

/**
 * Charge la liste des utilisateurs bloqués
 * @returns {Promise<Array>} - Liste des utilisateurs bloqués
 */
export async function loadBlockedUsers() {
    try {
        const blockedUsers = await apiRequest(API_ENDPOINTS.BLOCKED_USERS);

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
            groupedData: groupedData
        });

        return blockedUsers;
    } catch (error) {
        handleError(error, 'Impossible de charger les utilisateurs bloqués');
        return [];
    }
}

/**
 * Débloque un utilisateur ou une adresse IP
 * @param {string} identifier - Identifiant à débloquer
 * @param {string} type - Type d'identifiant ('user' ou 'ip')
 */
export async function unblockUser(identifier, type) {
    try {
        const options = createJsonRequestOptions('POST', { identifier, type });
        const result = await apiRequest(API_ENDPOINTS.UNBLOCK_USER, options);

        // Recharger la liste pour refléter le déblocage
        await loadBlockedUsers();

        return result;
    } catch (error) {
        handleError(error, 'Impossible de débloquer cet identifiant');
        throw error;
    }
}