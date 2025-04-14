/**
 * Authentication providers management module
 * Handles all operations related to external providers (Google, Microsoft, GitHub)
 */
import { apiRequest, createJsonRequestOptions, API_ENDPOINTS, handleError } from '../utils/api.js';
import store from '../state/store.js';

/**
 * Loads the list of authentication providers from the API
 */
export async function loadProviders() {
    try {
        const providers = await apiRequest(API_ENDPOINTS.PROVIDERS);
        store.update('providers', { list: providers });
        return providers;
    } catch (error) {
        handleError(error, 'Unable to load authentication providers');
        return [];
    }
}

/**
 * Creates a new authentication provider
 * @param {Object} providerData - Provider data to create
 * @returns {Promise<Object>} - The created provider
 */
export async function createProvider(providerData) {
    try {
        const options = createJsonRequestOptions('POST', providerData);
        const result = await apiRequest(API_ENDPOINTS.PROVIDERS, options);

        // Reload the list to integrate the new provider
        await loadProviders();
        return result;
    } catch (error) {
        handleError(error, 'Unable to create authentication provider');
        throw error;
    }
}

/**
 * Updates an existing authentication provider
 * @param {string} providerId - Provider ID
 * @param {Object} providerData - Data to update
 * @returns {Promise<Object>} - The updated provider
 */
export async function updateProvider(providerId, providerData) {
    try {
        const options = createJsonRequestOptions('PUT', providerData);
        const result = await apiRequest(`${API_ENDPOINTS.PROVIDERS}/${providerId}`, options);

        // Reload the list to update the UI
        await loadProviders();
        return result;
    } catch (error) {
        handleError(error, 'Unable to update authentication provider');
        throw error;
    }
}

/**
 * Deletes an authentication provider
 * @param {string} providerId - ID of provider to delete
 * @returns {Promise<void>}
 */
export async function deleteProvider(providerId) {
    try {
        const options = { method: 'DELETE' };
        await apiRequest(`${API_ENDPOINTS.PROVIDERS}/${providerId}`, options);

        // Reload the list to update the UI
        await loadProviders();
    } catch (error) {
        handleError(error, 'Unable to delete authentication provider');
        throw error;
    }
}

/**
 * Determines if the tenant_id field should be displayed for a provider type
 * @param {string} providerType - Provider type (google, microsoft, github)
 * @returns {boolean} - true if tenant_id should be displayed
 */
export function isTenantIdRequired(providerType) {
    return providerType === 'microsoft';
}