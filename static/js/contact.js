// Format date to a readable string
function formatDate(dateStr) {
    const date = new Date(dateStr);
    return date.toLocaleString();
}

// Load and display messages
async function loadMessages() {
    try {
        const response = await fetch('/api/v1/contacts');
        const result = await response.json();
        console.log('Messages response:', result);
        
        if (!result.data) {
            throw new Error('Invalid response format');
        }

        const messages = Array.isArray(result.data) ? result.data : [];
        const messagesList = document.getElementById('messages-list');

        if (messages.length === 0) {
            messagesList.innerHTML = '<div class="message-card">No messages yet. Be the first to send one!</div>';
            return;
        }

        messagesList.innerHTML = messages
            .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
            .map(msg => `
                <div class="message-card">
                    <div class="message-header">
                        <div class="message-info">
                            <span class="message-name">${msg.name}</span>
                            <span class="message-email">${msg.email}</span>
                        </div>
                        <span class="message-time">${formatDate(msg.created_at)}</span>
                    </div>
                    <p class="message-content">${msg.message}</p>
                </div>
            `)
            .join('');
    } catch (err) {
        console.error('Failed to load messages:', err);
        document.getElementById('messages-list').innerHTML = 
            '<div class="message-card error">Failed to load messages</div>';
    }
}

// Handle form submission
async function handleSubmit(event) {
    event.preventDefault();
    const form = document.getElementById('contact-form');
    const result = document.getElementById('form-result');
    const responseEl = document.getElementById('response');

    try {
        const response = await fetch('/api/v1/contacts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: form.querySelector('#name').value,
                email: form.querySelector('#email').value,
                message: form.querySelector('#message').value,
            }),
        });

        const data = await response.json();
        console.log('Submit response:', data);
        
        result.classList.remove('hidden');
        responseEl.textContent = JSON.stringify(data, null, 2);
        
        if (response.ok) {
            form.reset();
            // Reload messages after successful submission
            loadMessages();
        }
    } catch (err) {
        console.error('Submit error:', err);
        result.classList.remove('hidden');
        responseEl.textContent = `Error: ${err.message}`;
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Load initial messages
    loadMessages();
    
    // Set up form submission handler
    const form = document.getElementById('contact-form');
    form.addEventListener('submit', handleSubmit);
}); 