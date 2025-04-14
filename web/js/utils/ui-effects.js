/**
 * Utils for managing ZenAuth UI effects
 */

/**
 * Adds a ripple effect on buttons
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
 * Animates the appearance of cards
 */
function animateCards() {
    const cards = document.querySelectorAll('.card');
    cards.forEach((card, index) => {
        card.classList.add('animate-in');
        card.style.animationDelay = `${index * 0.1}s`;
    });
}

/**
 * Displays a notification
 * @param {string} message - Message to display
 * @param {string} type - Notification type ('success', 'error', 'info')
 * @param {number} duration - Display duration in ms
 */
function showNotification(message, type = 'info', duration = 5000) {
    // Check if wrapper exists, otherwise create it
    let wrapper = document.querySelector('.notification-wrapper');
    if (!wrapper) {
        wrapper = document.createElement('div');
        wrapper.classList.add('notification-wrapper');
        document.body.appendChild(wrapper);
    }

    // Create notification
    const notification = document.createElement('div');
    notification.classList.add('notification', type);

    // Add icon based on type
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

    // Add to wrapper
    wrapper.appendChild(notification);

    // Handle closing
    const closeBtn = notification.querySelector('.notification-close');
    closeBtn.addEventListener('click', () => {
        closeNotification(notification);
    });

    // Auto-close after specified duration
    if (duration > 0) {
        setTimeout(() => {
            closeNotification(notification);
        }, duration);
    }

    return notification;
}

/**
 * Closes a notification with animation
 * @param {HTMLElement} notification - Notification element
 */
function closeNotification(notification) {
    notification.classList.add('removing');
    setTimeout(() => {
        notification.remove();
    }, 300);
}

/**
 * Initializes all UI effects
 */
function initUIEffects() {
    addRippleEffect();
    animateCards();

    // Observer to animate new cards added dynamically
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

// Export functions
export {
    initUIEffects,
    showNotification,
    closeNotification
};