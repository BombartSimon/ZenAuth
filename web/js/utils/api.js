/**
 * Module de gestion des appels API
 * Centralise toutes les requêtes HTTP pour assurer une cohérence
 */

/**
 * Effectue une requête vers l'API avec gestion d'erreurs standardisée
 * @param {string} url - URL de la ressource
 * @param {Object} options - Options de la requête fetch
 * @returns {Promise} - Promesse contenant les données ou rejetant une erreur
 */
export async function apiRequest(url, options = {}) {
    try {
        const response = await fetch(url, options);

        if (!response.ok) {
            // Tenter de récupérer un message d'erreur si disponible
            let errorMessage;
            try {
                const errorData = await response.text();
                errorMessage = errorData || `HTTP error ${response.status}`;
            } catch (e) {
                errorMessage = `HTTP error ${response.status}`;
            }

            throw new Error(errorMessage);
        }

        // Vérifier si la réponse contient du JSON
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
            return await response.json();
        }

        return await response.text();
    } catch (error) {
        console.error(`API request failed: ${url}`, error);
        throw error;
    }
}

/**
 * Crée un objet d'options pour les requêtes avec un corps JSON
 * @param {string} method - Méthode HTTP (GET, POST, PUT, DELETE)
 * @param {Object} data - Données à envoyer au format JSON
 * @returns {Object} - Options pour fetch
 */
export function createJsonRequestOptions(method, data = null) {
    const options = {
        method,
        headers: {
            'Content-Type': 'application/json'
        }
    };

    if (data) {
        options.body = JSON.stringify(data);
    }

    return options;
}

/**
 * Collection des endpoints de l'API
 */
export const API_ENDPOINTS = {
    USERS: '/admin/users',
    EXTERNAL_USERS: '/admin/users?provider=external',
    CLIENTS: '/admin/clients',
    PROVIDERS: '/admin/auth-providers',
    BLOCKED_USERS: '/admin/blocked-users',
    UNBLOCK_USER: '/admin/unblock-user',
};

/**
 * Affiche une notification d'erreur standardisée
 * @param {Error} error - L'erreur à afficher
 * @param {string} fallbackMessage - Message à afficher si l'erreur n'a pas de message
 */
export function handleError(error, fallbackMessage = 'Une erreur est survenue') {
    const message = error.message || fallbackMessage;
    alert(message);
    console.error(error);
}