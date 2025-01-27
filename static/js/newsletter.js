// Constants and Configuration
const API = {
    SUBSCRIPTIONS: '/api/v1/subscriptions',
    HEADERS: {
        'Content-Type': 'application/json'
    }
};

const DOM_IDS = {
    NEWSLETTER_FORM: 'newsletter-form',
    API_RESPONSE: 'api-response'
};

const TEMPLATES = {
    API_RESPONSE: (type, data) => `
        <div class="api-response ${type}">
            <div class="api-response-header">
                <span class="api-response-type">${type.toUpperCase()}</span>
                <span class="api-response-time">${formatDate(new Date())}</span>
            </div>
            <pre class="api-response-data">${JSON.stringify(data, null, 2)}</pre>
        </div>
    `,
    DEFAULT_RESPONSE: `
        <div class="api-response default">
            <div class="api-response-header">
                <span class="api-response-type">Waiting</span>
            </div>
            <pre class="api-response-data">// Submit the form to see the API response
{
    "status": "waiting",
    "message": "No responses yet"
}</pre>
        </div>
    `
};

// Utility Functions
const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return new Intl.DateTimeFormat('default', {
        dateStyle: 'medium',
        timeStyle: 'short'
    }).format(date);
};

const logDebug = (message, data) => {
    const timestamp = new Date().toISOString();
    console.debug(`[${timestamp}] ${message}`, data ?? '');
};

const logError = (message, error) => {
    const timestamp = new Date().toISOString();
    console.error(`[${timestamp}] ${message}`, error);
    if (error?.stack) console.error(`[${timestamp}] Error stack:`, error.stack);
};

// DOM Helpers
const getElement = (id) => {
    const element = document.getElementById(id);
    if (!element) {
        logDebug(`Element not found: ${id}`);
    }
    return element;
};

// API Response Display Component
class APIResponseDisplay {
    constructor(containerId) {
        this.container = getElement(containerId);
        this.responses = [];
        this.maxResponses = 5;
        this.showDefault();
    }

    showDefault() {
        if (this.container) {
            this.container.innerHTML = TEMPLATES.DEFAULT_RESPONSE;
        }
    }

    addResponse(type, data) {
        this.responses.unshift({ type, data, timestamp: new Date() });
        if (this.responses.length > this.maxResponses) {
            this.responses.pop();
        }
        this.render();
    }

    render() {
        if (!this.container) return;
        
        this.container.innerHTML = this.responses
            .map(({ type, data }) => TEMPLATES.API_RESPONSE(type, data))
            .join('');
    }

    clear() {
        this.responses = [];
        this.showDefault();
    }
}

// Form Handler Component
class NewsletterForm {
    constructor(formId) {
        this.form = getElement(formId);
        this.apiResponse = new APIResponseDisplay(DOM_IDS.API_RESPONSE);
        this.setupListeners();
    }

    setupListeners() {
        if (this.form) {
            this.form.addEventListener('submit', this.handleSubmit.bind(this));
        }
    }

    getFormData() {
        return Object.fromEntries(
            ['name', 'email'].map(id => [
                id, 
                this.form.querySelector(`#${id}`).value
            ])
        );
    }

    async handleSubmit(event) {
        event.preventDefault();
        logDebug('Form submission started...');
        
        const formData = this.getFormData();
        logDebug('Form data:', formData);

        try {
            const response = await fetch(API.SUBSCRIPTIONS, {
                method: 'POST',
                headers: API.HEADERS,
                body: JSON.stringify(formData)
            });
            
            const data = await response.json();
            this.apiResponse.addResponse(response.ok ? 'success' : 'error', data);
            
            if (response.ok) {
                logDebug('Subscription successful, resetting form');
                this.form.reset();
            } else {
                logError('Subscription failed:', data);
            }
        } catch (err) {
            logError('Submit error:', err);
            this.apiResponse.addResponse('error', { 
                error: err.message 
            });
        }
    }
}

// Initialize form handler when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new NewsletterForm(DOM_IDS.NEWSLETTER_FORM);
}); 