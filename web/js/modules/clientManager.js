/**
 * Module de gestion des clients OAuth
 * Gère toutes les opérations CRUD pour les clients OAuth
 */
import { apiRequest, createJsonRequestOptions, API_ENDPOINTS, handleError } from '../utils/api.js';
import store from '../state/store.js';

/**
 * Charge la liste des clients OAuth depuis l'API
 */
export async function loadClients() {
    try {
        const clients = await apiRequest(API_ENDPOINTS.CLIENTS);
        store.update('clients', { list: clients });
        return clients;
    } catch (error) {
        handleError(error, 'Impossible de charger la liste des clients OAuth');
        return [];
    }
}

/**
 * Crée un nouveau client OAuth
 * @param {Object} clientData - Données du client à créer
 * @returns {Promise<Object>} - Le client créé
 */
export async function createClient(clientData) {
    try {
        const options = createJsonRequestOptions('POST', clientData);
        const result = await apiRequest(API_ENDPOINTS.CLIENTS, options);

        // Recharger la liste pour intégrer le nouveau client
        await loadClients();
        return result;
    } catch (error) {
        handleError(error, 'Impossible de créer le client OAuth');
        throw error;
    }
}

/**
 * Met à jour un client OAuth existant
 * @param {string} clientId - ID du client
 * @param {Object} clientData - Données à mettre à jour
 * @returns {Promise<Object>} - Le client mis à jour
 */
export async function updateClient(clientId, clientData) {
    try {
        const options = createJsonRequestOptions('PUT', clientData);
        const result = await apiRequest(`${API_ENDPOINTS.CLIENTS}/${clientId}`, options);

        // Recharger la liste pour mettre à jour l'UI
        await loadClients();
        return result;
    } catch (error) {
        handleError(error, 'Impossible de mettre à jour le client OAuth');
        throw error;
    }
}

/**
 * Supprime un client OAuth
 * @param {string} clientId - ID du client à supprimer
 * @returns {Promise<void>}
 */
export async function deleteClient(clientId) {
    try {
        await apiRequest(`${API_ENDPOINTS.CLIENTS}/${clientId}`, { method: 'DELETE' });

        // Recharger la liste pour refléter la suppression
        await loadClients();
    } catch (error) {
        handleError(error, 'Impossible de supprimer le client OAuth');
        throw error;
    }
}

/**
 * Formate les URI de redirection d'un client pour l'affichage/modification
 * @param {Array} uris - Tableau d'URIs
 * @param {string} format - Format cible ('string', 'array', 'list')
 * @returns {string|Array} - URIs formatées
 */
export function formatRedirectUris(uris, format = 'string') {
    if (format === 'string') {
        return Array.isArray(uris) ? uris.join(', ') : uris;
    } else if (format === 'list') {
        return Array.isArray(uris) ? uris.join('\n') : uris;
    } else if (format === 'array') {
        if (typeof uris === 'string') {
            return uris.split('\n')
                .map(uri => uri.trim())
                .filter(uri => uri !== '');
        }
        return uris;
    }
    return uris;
}