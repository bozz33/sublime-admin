// Main application JavaScript for Nourriture Solidaire Dashboard

// Utility functions
const Utils = {
    // Format number with thousands separator
    formatNumber(num) {
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
    },

    // Format date
    formatDate(date) {
        return new Intl.DateTimeFormat('fr-FR', {
            day: '2-digit',
            month: 'short',
            year: 'numeric'
        }).format(new Date(date));
    },

    // Time ago
    timeAgo(date) {
        const seconds = Math.floor((new Date() - new Date(date)) / 1000);
        const intervals = {
            année: 31536000,
            mois: 2592000,
            semaine: 604800,
            jour: 86400,
            heure: 3600,
            minute: 60
        };
        for (const [unit, secondsInUnit] of Object.entries(intervals)) {
            const interval = Math.floor(seconds / secondsInUnit);
            if (interval >= 1) {
                return `Il y a ${interval} ${unit}${interval > 1 && unit !== 'mois' ? 's' : ''}`;
            }
        }
        return 'À l\'instant';
    },

    // Debounce function
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};

// Table functionality
const DataTable = {
    init(tableId, options = {}) {
        const table = document.getElementById(tableId);
        if (!table) return;

        this.table = table;
        this.options = {
            searchable: true,
            sortable: true,
            pagination: true,
            perPage: 10,
            ...options
        };

        if (this.options.searchable) this.initSearch();
        if (this.options.sortable) this.initSort();
    },

    initSearch() {
        const searchInput = document.querySelector(`[data-table-search="${this.table.id}"]`);
        if (!searchInput) return;

        searchInput.addEventListener('input', Utils.debounce((e) => {
            const searchTerm = e.target.value.toLowerCase();
            const rows = this.table.querySelectorAll('tbody tr');
            
            rows.forEach(row => {
                const text = row.textContent.toLowerCase();
                row.style.display = text.includes(searchTerm) ? '' : 'none';
            });
        }, 300));
    },

    initSort() {
        const headers = this.table.querySelectorAll('th[data-sortable]');
        headers.forEach(header => {
            header.style.cursor = 'pointer';
            header.addEventListener('click', () => this.sortTable(header));
        });
    },

    sortTable(header) {
        const index = Array.from(header.parentNode.children).indexOf(header);
        const rows = Array.from(this.table.querySelectorAll('tbody tr'));
        const isAsc = header.dataset.sortDir !== 'asc';
        
        rows.sort((a, b) => {
            const aVal = a.children[index].textContent.trim();
            const bVal = b.children[index].textContent.trim();
            return isAsc ? aVal.localeCompare(bVal) : bVal.localeCompare(aVal);
        });

        header.dataset.sortDir = isAsc ? 'asc' : 'desc';
        const tbody = this.table.querySelector('tbody');
        rows.forEach(row => tbody.appendChild(row));
    }
};

// Modal functionality
const Modal = {
    open(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('hidden');
            document.body.style.overflow = 'hidden';
        }
    },

    close(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('hidden');
            document.body.style.overflow = '';
        }
    },

    init() {
        // Close modal on backdrop click
        document.querySelectorAll('[data-modal-backdrop]').forEach(backdrop => {
            backdrop.addEventListener('click', (e) => {
                if (e.target === backdrop) {
                    this.close(backdrop.id);
                }
            });
        });

        // Close modal on escape key
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                document.querySelectorAll('[data-modal-backdrop]:not(.hidden)').forEach(modal => {
                    this.close(modal.id);
                });
            }
        });
    }
};

// Toast notifications
const Toast = {
    container: null,

    init() {
        this.container = document.createElement('div');
        this.container.className = 'fixed bottom-4 right-4 z-50 space-y-2';
        document.body.appendChild(this.container);
    },

    show(message, type = 'info', duration = 3000) {
        const colors = {
            success: 'bg-green-500',
            error: 'bg-red-500',
            warning: 'bg-yellow-500',
            info: 'bg-blue-500'
        };

        const icons = {
            success: 'check_circle',
            error: 'error',
            warning: 'warning',
            info: 'info'
        };

        const toast = document.createElement('div');
        toast.className = `flex items-center gap-3 px-4 py-3 rounded-lg text-white ${colors[type]} shadow-lg transform translate-x-full transition-transform duration-300`;
        toast.innerHTML = `
            <span class="material-icons-outlined text-xl">${icons[type]}</span>
            <span class="text-sm font-medium">${message}</span>
            <button class="ml-2 hover:opacity-75" onclick="this.parentElement.remove()">
                <span class="material-icons-outlined text-lg">close</span>
            </button>
        `;

        this.container.appendChild(toast);
        
        // Animate in
        requestAnimationFrame(() => {
            toast.classList.remove('translate-x-full');
        });

        // Auto remove
        setTimeout(() => {
            toast.classList.add('translate-x-full');
            setTimeout(() => toast.remove(), 300);
        }, duration);
    }
};

// Form validation
const FormValidator = {
    validate(form) {
        const inputs = form.querySelectorAll('[required]');
        let isValid = true;

        inputs.forEach(input => {
            if (!input.value.trim()) {
                this.showError(input, 'Ce champ est requis');
                isValid = false;
            } else if (input.type === 'email' && !this.isValidEmail(input.value)) {
                this.showError(input, 'Email invalide');
                isValid = false;
            } else {
                this.clearError(input);
            }
        });

        return isValid;
    },

    isValidEmail(email) {
        return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    },

    showError(input, message) {
        this.clearError(input);
        input.classList.add('border-red-500');
        const error = document.createElement('p');
        error.className = 'text-red-500 text-xs mt-1 error-message';
        error.textContent = message;
        input.parentNode.appendChild(error);
    },

    clearError(input) {
        input.classList.remove('border-red-500');
        const error = input.parentNode.querySelector('.error-message');
        if (error) error.remove();
    }
};

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', () => {
    Modal.init();
    Toast.init();
    
    // Initialize tables with data-table attribute
    document.querySelectorAll('[data-table]').forEach(table => {
        DataTable.init(table.id);
    });
});

// Export for use in other scripts
window.Utils = Utils;
window.DataTable = DataTable;
window.Modal = Modal;
window.Toast = Toast;
window.FormValidator = FormValidator;
