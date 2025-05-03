// Constants and Configuration
const API = {
    SUBSCRIPTIONS: '/api/v1/subscriptions',
    HEADERS: {
        'Content-Type': 'application/json'
    }
};

const DOM_IDS = {
    DEMO_FORM: 'demo-form',
    API_RESPONSE: 'api-response',
    MESSAGES_LIST: 'messages-list'
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
class DemoForm {
    constructor(formId) {
        this.form = getElement(formId);
        this.apiResponse = new APIResponseDisplay(DOM_IDS.API_RESPONSE);
        this.messagesList = getElement(DOM_IDS.MESSAGES_LIST);
        this.setupListeners();
    }

    setupListeners() {
        if (this.form) {
            this.form.addEventListener('submit', this.handleSubmit.bind(this));
        }
    }

    getFormData() {
        const formData = new FormData(this.form);
        return {
            name: formData.get('name'),
            email: formData.get('email')
        };
    }

    async handleSubmit(event) {
        event.preventDefault();
        logDebug('Form submission started...');
        
        const formData = this.getFormData();
        logDebug('Form data:', formData);

        try {
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 500));

            // Add new submission to the list
            this.addSubmissionToList(formData);

            // Show success message
            this.showMessage('success', 'Form submitted successfully! Check the submissions list.');
            this.form.reset();
        } catch (err) {
            logError('Submit error:', err);
            this.apiResponse.addResponse('error', { 
                error: err.message 
            });
            this.showMessage('error', 'Failed to submit form. Please try again.');
        }
    }

    addSubmissionToList(submission) {
        const item = document.createElement('div');
        item.className = 'message-item';
        item.innerHTML = `
            <div class="message-header">
                <strong>${submission.name}</strong>
                <span class="timestamp">${submission.timestamp}</span>
            </div>
            <div class="message-content">
                <span class="email">${submission.email}</span>
            </div>
        `;
        this.messagesList.insertBefore(item, this.messagesList.firstChild);
    }

    showMessage(type, text) {
        const message = document.createElement('div');
        message.className = `alert alert-${type}`;
        message.textContent = text;
        this.form.parentNode.insertBefore(message, this.form);

        setTimeout(() => {
            message.remove();
        }, 5000);
    }
}

// Initialize form handler when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('demo-form');
    const messagesList = document.getElementById('messages-list');

    // Demo submissions to show in the list
    const demoSubmissions = [
        { name: 'John Smith', email: 'john@example.com', timestamp: '2 minutes ago' },
        { name: 'Sarah Wilson', email: 's.wilson@example.com', timestamp: '5 minutes ago' },
        { name: 'Mike Johnson', email: 'mike.j@example.com', timestamp: '10 minutes ago' }
    ];

    // Display demo submissions
    demoSubmissions.forEach(submission => {
        addSubmissionToList(submission);
    });

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const formData = new FormData(form);
        const data = {
            name: formData.get('name'),
            email: formData.get('email')
        };

        try {
            // Simulate API call
            await new Promise(resolve => setTimeout(resolve, 500));

            // Add new submission to the list
            addSubmissionToList(data);

            // Show success message
            showMessage('success', 'Form submitted successfully! Check the submissions list.');
            form.reset();

        } catch (error) {
            showMessage('error', 'Failed to submit form. Please try again.');
        }
    });

    function addSubmissionToList(submission) {
        const item = document.createElement('div');
        item.className = 'message-item';
        item.innerHTML = `
            <div class="message-header">
                <strong>${submission.name}</strong>
                <span class="timestamp">${submission.timestamp}</span>
            </div>
            <div class="message-content">
                <span class="email">${submission.email}</span>
            </div>
        `;
        messagesList.insertBefore(item, messagesList.firstChild);
    }

    function showMessage(type, text) {
        const message = document.createElement('div');
        message.className = `alert alert-${type}`;
        message.textContent = text;
        form.parentNode.insertBefore(message, form);

        setTimeout(() => {
            message.remove();
        }, 5000);
    }
}); 