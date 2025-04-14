/**
 * User management module
 * Handles all user-related operations (loading, adding, editing, deletion)
 */
import { apiRequest, createJsonRequestOptions, API_ENDPOINTS, handleError } from '../utils/api.js';
import store from '../state/store.js';

// Minimum time between two load requests (in ms)
const MIN_REFRESH_INTERVAL = 1000;
let lastLoadTime = 0;

/**
 * Loads the list of users from the API
 * Takes into account the source (local or external)
 * @param {boolean} force - Force reload even if called recently
 * @returns {Promise<Array>} - List of users
 */
export async function loadUsers(force = false) {
    try {
        // Avoid too many requests in a short time to prevent API overload
        const now = Date.now();
        if (!force && now - lastLoadTime < MIN_REFRESH_INTERVAL) {
            console.log('⏱️ Request too close, using cache');
            return store.getState('users').list;
        }

        lastLoadTime = now;
        const { source } = store.getState('users');
        const url = source === "external" ? API_ENDPOINTS.EXTERNAL_USERS : API_ENDPOINTS.USERS;

        // Update state to indicate loading
        store.update('users', { loading: true, error: null });

        const users = await apiRequest(url);

        // Verify that data is an array
        if (!Array.isArray(users)) {
            throw new Error('Invalid user data format');
        }

        // Update state with data
        store.update('users', { list: users, loading: false, lastUpdated: now });

        return users;
    } catch (error) {
        store.update('users', {
            loading: false,
            error: error.message || 'Error loading users'
        });

        handleError(error, 'Unable to load user list');
        return [];
    }
}

/**
 * Toggle between user sources (local/external)
 */
export function toggleUserSource() {
    const currentState = store.getState('users');
    const newSource = currentState.source === "external" ? "local" : "external";
    store.update('users', { source: newSource, loading: true });
    loadUsers(true); // Force reload
}

/**
 * Creates a new user
 * @param {Object} userData - User data to create
 * @returns {Promise<Object>} - The created user
 */
export async function createUser(userData) {
    try {
        const options = createJsonRequestOptions('POST', userData);
        const result = await apiRequest(API_ENDPOINTS.USERS, options);

        // Reload the list to include the new user
        await loadUsers(true);

        // Display success message
        showSuccessMessage('User created successfully');

        return result;
    } catch (error) {
        handleError(error, 'Unable to create user');
        throw error;
    }
}

/**
 * Updates an existing user
 * @param {string} userId - User ID
 * @param {Object} userData - Data to update
 * @returns {Promise<Object>} - The updated user
 */
export async function updateUser(userId, userData) {
    try {
        const options = createJsonRequestOptions('PUT', userData);
        const result = await apiRequest(`${API_ENDPOINTS.USERS}/${userId}`, options);

        // Reload the list to update UI
        await loadUsers(true);

        // Display success message
        showSuccessMessage('User updated successfully');

        return result;
    } catch (error) {
        handleError(error, 'Unable to update user');
        throw error;
    }
}

/**
 * Deletes a user
 * @param {string} userId - ID of the user to delete
 * @returns {Promise<void>}
 */
export async function deleteUser(userId) {
    try {
        await apiRequest(`${API_ENDPOINTS.USERS}/${userId}`, { method: 'DELETE' });

        // Reload the list to reflect deletion
        await loadUsers(true);

        // Display success message
        showSuccessMessage('User deleted successfully');
    } catch (error) {
        handleError(error, 'Unable to delete user');
        throw error;
    }
}

/**
 * Loads the list of blocked users
 * @param {boolean} force - Force reload even if called recently
 * @returns {Promise<Array>} - List of blocked users
 */
export async function loadBlockedUsers(force = false) {
    try {
        // Avoid too many close requests
        const now = Date.now();
        if (!force && now - lastLoadTime < MIN_REFRESH_INTERVAL) {
            return store.getState('blockedUsers').list;
        }

        // Update state to indicate loading
        store.update('blockedUsers', { loading: true, error: null });

        const blockedUsers = await apiRequest(API_ENDPOINTS.BLOCKED_USERS);

        // Verify that data is an array
        if (!Array.isArray(blockedUsers)) {
            throw new Error('Invalid blocked users data format');
        }

        // Organize data to group users and their IPs
        const groupedData = {};

        blockedUsers.forEach(user => {
            // If it's a user, we process it separately
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
            // If it's an IP, check if it's associated with a user
            else if (user.type === "ip") {
                let found = false;

                // Search in existing associations
                Object.values(groupedData).forEach(entry => {
                    if (entry.associatedIPs && entry.associatedIPs.includes(user.identifier)) {
                        found = true;
                    }
                });

                // If this IP is not already associated, create a new entry
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
            error: error.message || 'Error loading blocked users'
        });

        handleError(error, 'Unable to load blocked users');
        return [];
    }
}

/**
 * Unblocks a user or IP address
 * @param {string} identifier - Identifier to unblock
 * @param {string} type - Type of identifier ('user' or 'ip')
 * @returns {Promise<Object>} - Result of the operation
 */
export async function unblockUser(identifier, type) {
    try {
        const options = createJsonRequestOptions('POST', { identifier, type });
        const result = await apiRequest(API_ENDPOINTS.UNBLOCK_USER, options);

        // Reload the list to reflect unblocking
        await loadBlockedUsers(true);

        // Display success message
        showSuccessMessage(`${type === 'user' ? 'User' : 'IP Address'} unblocked successfully`);

        return result;
    } catch (error) {
        handleError(error, 'Unable to unblock this identifier');
        throw error;
    }
}

/**
 * Displays a success notification
 * @param {string} message - Message to display
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

    // Entry animation
    setTimeout(() => {
        notification.classList.add('show');
    }, 10);

    // Add handler to close notification
    notification.querySelector('.success-close').addEventListener('click', () => {
        notification.classList.remove('show');
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    });

    // Auto-close after 3 seconds
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