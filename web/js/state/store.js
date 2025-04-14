/**
 * Simple state manager for ZenAuth
 * Centralizes state management between different parts of the application
 */
const store = (function () {
    // Private application state
    const state = {
        users: {
            list: [],
            source: 'local', // 'local' or 'external'
        },
        clients: {
            list: []
        },
        providers: {
            list: []
        },
        blockedUsers: {
            list: [],
            groupedData: {}
        },
        ui: {
            theme: localStorage.getItem('theme') || 'light',
            activeSection: 'users',
            modals: {
                user: { isOpen: false, mode: 'add', data: {} },
                client: { isOpen: false, mode: 'add', data: {} },
                provider: { isOpen: false, mode: 'add', data: {} },
                delete: { isOpen: false, type: null, id: null, name: '' }
            }
        }
    };

    // Callbacks for subscribers
    const subscribers = {
        users: [],
        clients: [],
        providers: [],
        blockedUsers: [],
        ui: []
    };

    /**
     * Updates a part of the state and notifies subscribers
     * @param {string} section - The section to update ('users', 'clients', etc.)
     * @param {object} newData - The new data to merge
     */
    function update(section, newData) {
        if (!state[section]) {
            console.error(`Section ${section} doesn't exist in the store`);
            return;
        }

        // Data merging
        state[section] = { ...state[section], ...newData };

        // Notify subscribers
        if (subscribers[section]) {
            subscribers[section].forEach(callback => callback(state[section]));
        }
    }

    /**
     * Subscribes to changes in a section of the state
     * @param {string} section - The section to observe
     * @param {function} callback - Function called on changes
     * @returns {function} - Function to unsubscribe
     */
    function subscribe(section, callback) {
        if (!subscribers[section]) {
            subscribers[section] = [];
        }

        subscribers[section].push(callback);

        // Return an unsubscribe function
        return () => {
            const index = subscribers[section].indexOf(callback);
            if (index !== -1) {
                subscribers[section].splice(index, 1);
            }
        };
    }

    /**
     * Gets the current state of a section
     * @param {string} section - The section to retrieve
     * @returns {object} - The state of the section
     */
    function getState(section) {
        return section ? { ...state[section] } : { ...state };
    }

    // Public API
    return {
        update,
        subscribe,
        getState
    };
})();

export default store;