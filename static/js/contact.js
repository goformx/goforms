// Format date to a readable string
function formatDate(dateStr) {
    console.debug('Formatting date:', dateStr);
    const date = new Date(dateStr);
    const formatted = date.toLocaleString();
    console.debug('Formatted date:', formatted);
    return formatted;
}

// Load and display messages
async function loadMessages() {
    console.debug('Loading messages...');
    try {
        console.debug('Fetching messages from API...');
        const response = await fetch('/api/v1/contacts');
        console.debug('API Response status:', response.status);
        console.debug('API Response headers:', Object.fromEntries(response.headers.entries()));

        const result = await response.json();
        console.debug('API Response data:', result);
        
        const messages = result.data ?? [];
        console.debug('Processed messages array:', messages);
        const messagesList = document.getElementById('messages-list');

        if (messages.length === 0) {
            console.debug('No messages found, showing empty state');
            messagesList.innerHTML = '<div class="message-card">No messages yet. Be the first to send one!</div>';
            return;
        }

        console.debug('Sorting messages by date...');
        const sortedMessages = messages.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
        console.debug('Sorted messages:', sortedMessages);

        console.debug('Rendering messages to DOM...');
        messagesList.innerHTML = sortedMessages
            .map(msg => {
                console.debug('Rendering message:', msg);
                return `
                    <div class="message-card">
                        <div class="message-header">
                            <div class="message-info">
                                <span class="message-name">${msg.name ?? 'Anonymous'}</span>
                                <span class="message-email">${msg.email ?? 'No email'}</span>
                            </div>
                            <span class="message-time">${formatDate(msg.created_at)}</span>
                        </div>
                        <p class="message-content">${msg.message ?? 'No message'}</p>
                    </div>
                `;
            })
            .join('');
        console.debug('Messages rendered successfully');
    } catch (err) {
        console.error('Failed to load messages:', err);
        console.error('Error stack:', err.stack);
        document.getElementById('messages-list').innerHTML = 
            '<div class="message-card error">Failed to load messages</div>';
    }
}

// Handle form submission
async function handleSubmit(event) {
    console.debug('Form submission started...');
    event.preventDefault();
    
    const form = document.getElementById('contact-form');
    const result = document.getElementById('form-result');
    const responseEl = document.getElementById('response');

    const formData = {
        name: form.querySelector('#name').value,
        email: form.querySelector('#email').value,
        message: form.querySelector('#message').value,
    };
    console.debug('Form data:', formData);

    try {
        console.debug('Sending POST request to API...');
        const response = await fetch('/api/v1/contacts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(formData),
        });
        console.debug('API Response status:', response.status);
        console.debug('API Response headers:', Object.fromEntries(response.headers.entries()));

        const data = await response.json();
        console.debug('API Response data:', data);
        
        result.classList.remove('hidden');
        responseEl.textContent = JSON.stringify(data, null, 2);
        
        if (response.ok) {
            console.debug('Submission successful, resetting form');
            form.reset();
            console.debug('Reloading messages...');
            await loadMessages();
        } else {
            console.error('Submission failed:', data);
        }
    } catch (err) {
        console.error('Submit error:', err);
        console.error('Error stack:', err.stack);
        result.classList.remove('hidden');
        responseEl.textContent = `Error: ${err.message}`;
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.debug('DOM loaded, initializing...');
    
    // Load initial messages
    console.debug('Loading initial messages...');
    loadMessages();
    
    // Set up form submission handler
    console.debug('Setting up form submission handler...');
    const form = document.getElementById('contact-form');
    form.addEventListener('submit', handleSubmit);
    console.debug('Initialization complete');
}); 