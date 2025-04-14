/**
 * Utils pour gérer les effets UI de ZenAuth
 */

/**
 * Ajoute un effet de ripple sur les boutons
 */
function addRippleEffect() {
    const buttons = document.querySelectorAll('.btn');

    buttons.forEach(button => {
        button.classList.add('with-ripple');

        button.addEventListener('click', function (e) {
            const rect = button.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;

            const ripple = document.createElement('span');
            ripple.classList.add('ripple');
            ripple.style.left = `${x}px`;
            ripple.style.top = `${y}px`;

            button.appendChild(ripple);

            setTimeout(() => {
                ripple.remove();
            }, 600);
        });
    });
}

/**
 * Anime l'apparition des cartes
 */
function animateCards() {
    const cards = document.querySelectorAll('.card');
    cards.forEach((card, index) => {
        card.classList.add('animate-in');
        card.style.animationDelay = `${index * 0.1}s`;
    });
}

/**
 * Affiche une notification
 * @param {string} message - Message à afficher
 * @param {string} type - Type de notification ('success', 'error', 'info')
 * @param {number} duration - Durée d'affichage en ms
 */
function showNotification(message, type = 'info', duration = 5000) {
    // Vérifier si le wrapper existe, sinon le créer
    let wrapper = document.querySelector('.notification-wrapper');
    if (!wrapper) {
        wrapper = document.createElement('div');
        wrapper.classList.add('notification-wrapper');
        document.body.appendChild(wrapper);
    }

    // Créer la notification
    const notification = document.createElement('div');
    notification.classList.add('notification', type);

    // Ajouter l'icône en fonction du type
    let icon = 'info-circle';
    if (type === 'success') icon = 'check-circle';
    if (type === 'error') icon = 'exclamation-circle';

    notification.innerHTML = `
        <div class="notification-icon">
            <i class="fas fa-${icon}"></i>
        </div>
        <div class="notification-message">${message}</div>
        <button class="notification-close">&times;</button>
    `;

    // Ajouter au wrapper
    wrapper.appendChild(notification);

    // Gérer la fermeture
    const closeBtn = notification.querySelector('.notification-close');
    closeBtn.addEventListener('click', () => {
        closeNotification(notification);
    });

    // Auto-fermeture après durée spécifiée
    if (duration > 0) {
        setTimeout(() => {
            closeNotification(notification);
        }, duration);
    }

    return notification;
}

/**
 * Ferme une notification avec animation
 * @param {HTMLElement} notification - Élément de notification
 */
function closeNotification(notification) {
    notification.classList.add('removing');
    setTimeout(() => {
        notification.remove();
    }, 300);
}

/**
 * Initialise tous les effets UI
 */
function initUIEffects() {
    addRippleEffect();
    animateCards();

    // Observer pour animer les nouvelles cartes ajoutées dynamiquement
    const observer = new MutationObserver(mutations => {
        mutations.forEach(mutation => {
            if (mutation.addedNodes.length) {
                mutation.addedNodes.forEach(node => {
                    if (node.classList && node.classList.contains('card') && !node.classList.contains('animate-in')) {
                        node.classList.add('animate-in');
                    }

                    if (node.querySelectorAll) {
                        const newCards = node.querySelectorAll('.card:not(.animate-in)');
                        newCards.forEach((card, index) => {
                            card.classList.add('animate-in');
                            card.style.animationDelay = `${index * 0.1}s`;
                        });
                    }
                });
            }
        });
    });

    observer.observe(document.body, { childList: true, subtree: true });
}

// Exporter les fonctions
export {
    initUIEffects,
    showNotification,
    closeNotification
};