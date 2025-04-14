/**
 * API call management module
 * Centralizes all HTTP requests to ensure consistency
 */

/**
 * Makes a request to the API with standardized error handling and retry mechanism
 * @param {string} url - Resource URL
 * @param {Object} options - Fetch request options
 * @param {number} retryCount - Remaining attempts (used internally)
 * @returns {Promise} - Promise containing data or rejecting an error
 */
export async function apiRequest(url, options = {}, retryCount = 3) {
    try {
        console.log(`📡 API Request: ${options.method || 'GET'} ${url}`);

        const response = await fetch(url, options);

        if (!response.ok) {
            // Try to retrieve an error message if available
            let errorMessage;
            try {
                const errorData = await response.text();
                errorMessage = errorData || `HTTP error ${response.status}`;
            } catch (e) {
                errorMessage = `HTTP error ${response.status}`;
            }

            throw new Error(errorMessage);
        }

        // Check if the response contains JSON
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

        // Retry attempt if we still have available attempts
        if (retryCount > 0) {
            console.log(`🔄 Retrying... (${retryCount} attempts left)`);
            // Wait a bit before retrying (exponential delay)
            const delay = (3 - retryCount + 1) * 1000;
            await new Promise(resolve => setTimeout(resolve, delay));
            return apiRequest(url, options, retryCount - 1);
        }

        throw error;
    }
}

/**
 * Creates an options object for requests with a JSON body
 * @param {string} method - HTTP method (GET, POST, PUT, DELETE)
 * @param {Object} data - Data to send in JSON format
 * @returns {Object} - Options for fetch
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
 * Collection of API endpoints
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
 * Displays a standardized error notification
 * @param {Error} error - The error to display
 * @param {string} fallbackMessage - Message to display if the error has no message
 */
export function handleError(error, fallbackMessage = 'An error occurred') {
    const message = error.message || fallbackMessage;

    // Create a stylized notification instead of an alert
    const notification = document.createElement('div');
    notification.className = 'error-notification';
    notification.innerHTML = `
        <div class="error-icon">❌</div>
        <div class="error-message">${message}</div>
        <button class="error-close">×</button>
    `;

    document.body.appendChild(notification);

    // Entry animation
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);

    // Add handler to close the notification
    notification.querySelector('.error-close').addEventListener('click', () => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    });

    // Auto-close after 5 seconds
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