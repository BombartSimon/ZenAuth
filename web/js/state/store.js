/**
 * Gestionnaire d'état simple pour ZenAuth
 * Permet de centraliser la gestion des états entre les différentes parties de l'application
 */
const store = (function () {
    // État privé de l'application
    const state = {
        users: {
            list: [],
            source: 'local', // 'local' ou 'external'
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

    // Callbacks pour les souscripteurs
    const subscribers = {
        users: [],
        clients: [],
        providers: [],
        blockedUsers: [],
        ui: []
    };

    /**
     * Met à jour une partie du state et notifie les souscripteurs
     * @param {string} section - La section à mettre à jour ('users', 'clients', etc.)
     * @param {object} newData - Les nouvelles données à fusionner
     */
    function update(section, newData) {
        if (!state[section]) {
            console.error(`Section ${section} n'existe pas dans le store`);
            return;
        }

        // Fusion des données
        state[section] = { ...state[section], ...newData };

        // Notification des souscripteurs
        if (subscribers[section]) {
            subscribers[section].forEach(callback => callback(state[section]));
        }
    }

    /**
     * S'abonne aux changements d'une section du state
     * @param {string} section - La section à observer
     * @param {function} callback - Fonction appelée lors des changements
     * @returns {function} - Fonction pour se désabonner
     */
    function subscribe(section, callback) {
        if (!subscribers[section]) {
            subscribers[section] = [];
        }

        subscribers[section].push(callback);

        // Retourner une fonction de désabonnement
        return () => {
            const index = subscribers[section].indexOf(callback);
            if (index !== -1) {
                subscribers[section].splice(index, 1);
            }
        };
    }

    /**
     * Récupère l'état actuel d'une section
     * @param {string} section - La section à récupérer
     * @returns {object} - L'état de la section
     */
    function getState(section) {
        return section ? { ...state[section] } : { ...state };
    }

    // API publique
    return {
        update,
        subscribe,
        getState
    };
})();

export default store;