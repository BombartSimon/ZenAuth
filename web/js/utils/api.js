/**
 * Module de gestion des appels API
 * Centralise toutes les requêtes HTTP pour assurer une cohérence
 */

/**
 * Effectue une requête vers l'API avec gestion d'erreurs standardisée et mécanisme de retentatives
 * @param {string} url - URL de la ressource
 * @param {Object} options - Options de la requête fetch
 * @param {number} retryCount - Nombre de tentatives restantes (utilisé en interne)
 * @returns {Promise} - Promesse contenant les données ou rejetant une erreur
 */
export async function apiRequest(url, options = {}, retryCount = 3) {
    try {
        console.log(`📡 API Request: ${options.method || 'GET'} ${url}`);

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
            const data = await response.json();
            console.log(`✅ API Response: ${url}`, data);
            return data;
        }

        const text = await response.text();
        console.log(`✅ API Response (text): ${url}`);
        return text;
    } catch (error) {
        console.error(`❌ API request failed: ${url}`, error);

        // Tentative de nouvelle requête si nous avons encore des essais disponibles
        if (retryCount > 0) {
            console.log(`🔄 Retrying... (${retryCount} attempts left)`);
            // Attendre un peu avant de réessayer (délai exponentiel)
            const delay = (3 - retryCount + 1) * 1000;
            await new Promise(resolve => setTimeout(resolve, delay));
            return apiRequest(url, options, retryCount - 1);
        }

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

    // Créer une notification stylisée au lieu d'une alerte
    const notification = document.createElement('div');
    notification.className = 'error-notification';
    notification.innerHTML = `
        <div class="error-icon">❌</div>
        <div class="error-message">${message}</div>
        <button class="error-close">×</button>
    `;

    document.body.appendChild(notification);

    // Animation d'entrée
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);

    // Ajouter un gestionnaire pour fermer la notification
    notification.querySelector('.error-close').addEventListener('click', () => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    });

    // Auto-fermeture après 5 secondes
    setTimeout(() => {
        if (document.body.contains(notification)) {
            notification.classList.remove('show');
            setTimeout(() => {
                if (document.body.contains(notification)) {
                    document.body.removeChild(notification);
                }
            }, 300);
        }
    }, 5000);

    console.error(error);
}