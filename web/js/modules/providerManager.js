/**
 * Module de gestion des fournisseurs d'authentification
 * Gère toutes les opérations liées aux providers externes (Google, Microsoft, GitHub)
 */
import { apiRequest, createJsonRequestOptions, API_ENDPOINTS, handleError } from '../utils/api.js';
import store from '../state/store.js';

/**
 * Charge la liste des fournisseurs d'authentification depuis l'API
 */
export async function loadProviders() {
    try {
        const providers = await apiRequest(API_ENDPOINTS.PROVIDERS);
        store.update('providers', { list: providers });
        return providers;
    } catch (error) {
        handleError(error, 'Impossible de charger les fournisseurs d\'authentification');
        return [];
    }
}

/**
 * Crée un nouveau fournisseur d'authentification
 * @param {Object} providerData - Données du fournisseur à créer
 * @returns {Promise<Object>} - Le fournisseur créé
 */
export async function createProvider(providerData) {
    try {
        const options = createJsonRequestOptions('POST', providerData);
        const result = await apiRequest(API_ENDPOINTS.PROVIDERS, options);

        // Recharger la liste pour intégrer le nouveau fournisseur
        await loadProviders();
        return result;
    } catch (error) {
        handleError(error, 'Impossible de créer le fournisseur d\'authentification');
        throw error;
    }
}

/**
 * Met à jour un fournisseur d'authentification existant
 * @param {string} providerId - ID du fournisseur
 * @param {Object} providerData - Données à mettre à jour
 * @returns {Promise<Object>} - Le fournisseur mis à jour
 */
export async function updateProvider(providerId, providerData) {
    try {
        const options = createJsonRequestOptions('PUT', providerData);
        const result = await apiRequest(`${API_ENDPOINTS.PROVIDERS}/${providerId}`, options);

        // Recharger la liste pour mettre à jour l'UI
        await loadProviders();
        return result;
    } catch (error) {
        handleError(error, 'Impossible de mettre à jour le fournisseur d\'authentification');
        throw error;
    }
}

/**
 * Supprime un fournisseur d'authentification
 * @param {string} providerId - ID du fournisseur à supprimer
 * @returns {Promise<void>}
 */
export async function deleteProvider(providerId) {
    try {
        await apiRequest(`${API_ENDPOINTS.PROVIDERS}/${providerId}`, { method: 'DELETE' });

        // Recharger la liste pour refléter la suppression
        await loadProviders();
    } catch (error) {
        handleError(error, 'Impossible de supprimer le fournisseur d\'authentification');
        throw error;
    }
}

/**
 * Détermine si le champ tenant_id doit être affiché pour un type de fournisseur
 * @param {string} providerType - Type de fournisseur (google, microsoft, github)
 * @returns {boolean} - true si tenant_id doit être affiché
 */
export function isTenantIdRequired(providerType) {
    return providerType === 'microsoft';
}