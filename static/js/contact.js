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
    RESPONSE: 'response'
};

const TEMPLATES = {
    NO_MESSAGES: '<div class="message-card">No messages yet. Be the first to send one!</div>',
    ERROR_MESSAGE: '<div class="message-card error">Failed to load messages</div>',
    MESSAGE_CARD: ({ name, email, message, created_at }) => `
        <div class="message-card">
            <div class="message-header">
                <div class="message-info">
                    <span class="message-name">${name ?? 'Anonymous'}</span>
                    <span class="message-email">${email ?? 'No email'}</span>
                </div>
                <span class="message-time">${formatDate(created_at)}</span>
            </div>
            <p class="message-content">${message ?? 'No message'}</p>
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
    console.debug(message, data ?? '');
};

const logError = (message, error) => {
    console.error(message, error);
    if (error?.stack) console.error('Error stack:', error.stack);
};

// DOM Helpers
const getElement = (id) => document.getElementById(id);

const updateElement = (id, content) => {
    const element = getElement(id);
    if (element) element.innerHTML = content;
};

const showResult = (content) => {
    const result = getElement(DOM_IDS.FORM_RESULT);
    const responseEl = getElement(DOM_IDS.RESPONSE);
    if (result) result.classList.remove('hidden');
    if (responseEl) responseEl.textContent = content;
};

// API Functions
const fetchMessages = async () => {
    const response = await fetch(API.CONTACTS);
    logDebug('API Response status:', response.status);
    logDebug('API Response headers:', Object.fromEntries(response.headers.entries()));
    
    const { data: messages = [] } = await response.json();
    logDebug('Processed messages array:', messages);
    return messages;
};

const submitContact = async (formData) => {
    const response = await fetch(API.CONTACTS, {
        method: 'POST',
        headers: API.HEADERS,
        body: JSON.stringify(formData)
    });
    
    logDebug('API Response status:', response.status);
    logDebug('API Response headers:', Object.fromEntries(response.headers.entries()));
    
    const data = await response.json();
    logDebug('API Response data:', data);
    return { ok: response.ok, data };
};

// Message Handling
const sortMessagesByDate = (messages) => 
    [...messages].sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

const renderMessages = (messages) => {
    if (!messages?.length) {
        logDebug('No messages found, showing empty state');
        return TEMPLATES.NO_MESSAGES;
    }

    logDebug('Sorting and rendering messages...');
    return sortMessagesByDate(messages)
        .map(TEMPLATES.MESSAGE_CARD)
        .join('');
};

// Main Functions
const loadMessages = async () => {
    logDebug('Loading messages...');
    try {
        const messages = await fetchMessages();
        updateElement(DOM_IDS.MESSAGES_LIST, renderMessages(messages));
        logDebug('Messages rendered successfully');
    } catch (err) {
        logError('Failed to load messages:', err);
        updateElement(DOM_IDS.MESSAGES_LIST, TEMPLATES.ERROR_MESSAGE);
    }
};

const handleSubmit = async (event) => {
    logDebug('Form submission started...');
    event.preventDefault();
    
    const form = getElement(DOM_IDS.CONTACT_FORM);
    const formData = Object.fromEntries(
        ['name', 'email', 'message'].map(id => [
            id, 
            form.querySelector(`#${id}`).value
        ])
    );
    logDebug('Form data:', formData);

    try {
        const { ok, data } = await submitContact(formData);
        showResult(JSON.stringify(data, null, 2));
        
        if (ok) {
            logDebug('Submission successful, resetting form');
            form.reset();
            await loadMessages();
        } else {
            logError('Submission failed:', data);
        }
    } catch (err) {
        logError('Submit error:', err);
        showResult(`Error: ${err.message}`);
    }
};

// Initialization
const initialize = () => {
    logDebug('DOM loaded, initializing...');
    loadMessages();
    getElement(DOM_IDS.CONTACT_FORM)?.addEventListener('submit', handleSubmit);
    logDebug('Initialization complete');
};

document.addEventListener('DOMContentLoaded', initialize);
