// Constants and Configuration
const API = {
    CONTACTS: '/api/v1/contacts',
    HEADERS: {
        'Content-Type': 'application/json'
    }
};

const DOM_IDS = {
    CONTACT_FORM: 'contact-form',
    MESSAGES_LIST: 'messages-list',
    FORM_RESULT: 'form-result',
    RESPONSE: 'response',
    API_RESPONSE: 'api-response'
};

const TEMPLATES = {
    NO_MESSAGES: '<div class="message-card">No messages yet. Be the first to send one!</div>',
    ERROR_MESSAGE: '<div class="message-card error">Failed to load messages</div>',
    MESSAGE_CARD: ({ name, message, created_at }) => `
        <div class="message-card">
            <div class="message-header">
                <span class="message-name">${name ?? 'Anonymous'}</span>
                <span class="message-time">${formatDate(created_at)}</span>
            </div>
            <p class="message-content">${message ?? 'No message'}</p>
        </div>
    `,
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

const updateElement = (id, content) => {
    const element = getElement(id);
    if (element) element.innerHTML = content;
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

// Message Display Component
class MessageDisplay {
    constructor(containerId) {
        this.container = getElement(containerId);
        this.messages = [];
    }

    setMessages(messages) {
        this.messages = messages;
        this.render();
    }

    render() {
        if (!this.container) return;
        updateElement(DOM_IDS.MESSAGES_LIST, renderMessages(this.messages));
    }

    showError() {
        if (!this.container) return;
        updateElement(DOM_IDS.MESSAGES_LIST, TEMPLATES.ERROR_MESSAGE);
    }
}

// API Functions
const fetchMessages = async () => {
    logDebug('Fetching messages from API...');
    try {
        logDebug('Making GET request to:', API.CONTACTS);
        const response = await fetch(API.CONTACTS);
        logDebug('Response status:', response.status);
        
        if (!response.ok) {
            logError('Failed to fetch messages:', `HTTP ${response.status} - ${response.statusText}`);
            return [];
        }
        
        const responseData = await response.json();
        logDebug('Response data:', responseData);
        const messages = responseData.data || [];
        logDebug('Extracted messages:', messages.length);
        return messages;
    } catch (err) {
        logError('Failed to fetch messages:', err);
        return [];
    }
};

const submitContact = async (formData) => {
    logDebug('Submitting contact form:', formData);
    try {
        const response = await fetch(API.CONTACTS, {
            method: 'POST',
            headers: API.HEADERS,
            body: JSON.stringify(formData)
        });
        logDebug('Submit response status:', response.status);
        
        const data = await response.json();
        logDebug('Submit response data:', data);
        return { ok: response.ok, data };
    } catch (err) {
        logError('Submit request failed:', err);
        throw err;
    }
};

// Message Handling
const sortMessagesByDate = (messages) => {
    logDebug('Sorting messages by date, count:', messages.length);
    return [...messages].sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
};

const renderMessages = (messages) => {
    if (!messages?.length) {
        logDebug('No messages to render, showing empty state');
        return TEMPLATES.NO_MESSAGES;
    }

    logDebug('Rendering messages, count:', messages.length);
    const sorted = sortMessagesByDate(messages);
    logDebug('Messages sorted, rendering cards');
    return sorted.map(TEMPLATES.MESSAGE_CARD).join('');
};

// Form Handler Component
class ContactForm {
    constructor(formId, onSubmitSuccess) {
        this.form = getElement(formId);
        this.onSubmitSuccess = onSubmitSuccess;
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
            ['name', 'email', 'message'].map(id => [
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
            const { ok, data } = await submitContact(formData);
            this.apiResponse.addResponse(ok ? 'success' : 'error', data);
            
            if (ok) {
                logDebug('Submission successful, resetting form');
                this.form.reset();
                await this.onSubmitSuccess();
            } else {
                logError('Submission failed:', data);
            }
        } catch (err) {
            logError('Submit error:', err);
            this.apiResponse.addResponse('error', { 
                error: err.message 
            });
        }
    }
}

// Main Functions
const loadMessages = async (display) => {
    logDebug('Starting message load...');
    try {
        const messages = await fetchMessages();
        logDebug('Messages fetched successfully, updating display');
        display.setMessages(messages);
        logDebug('Display updated with messages:', messages.length);
    } catch (err) {
        logError('Message load failed:', err);
        display.showError();
    }
};

// Initialization
const initialize = () => {
    logDebug('DOM loaded, initializing...');
    
    const messageDisplay = new MessageDisplay(DOM_IDS.MESSAGES_LIST);
    const contactForm = new ContactForm(
        DOM_IDS.CONTACT_FORM, 
        () => loadMessages(messageDisplay)
    );
    
    loadMessages(messageDisplay);
    logDebug('Initialization complete');
};

document.addEventListener('DOMContentLoaded', initialize);
