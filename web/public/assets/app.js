const formMessage = document.getElementById('form-message');
const messageInput = document.getElementById('message-input');
const messagesList = document.getElementById('messages-list');
const app = document.getElementById('app');
const typingMessage = document.getElementById('typing');
const websocketProtocol = document.location.protocol === 'https:' ? 'wss' : 'ws';
const websocket = new WebSocket(`${websocketProtocol}://${document.location.host}/ws`);

// Controls user typing intervals
const throttleIntervalTime = 200;
const clearIntervalTime = 900;
var clearTimerId = null;
var lastTypingSentAt = 0;

// User base informations
var pid = null;
var username = null;

// When receive an message from server
websocket.onmessage = function (event) {
    const data = JSON.parse(event.data);

    switch (data.type) {
        case 'setup':
            pid = data.pid
            username = data.username
            break;
        case 'message':
            appendMessage(data);
            break;
        case 'typing':
            typingMessage.textContent = data.content;
            clearTimeout(clearTimerId);
            clearTimerId = setTimeout(() => { typingMessage.textContent = '' }, clearIntervalTime);
            break;
    }
};

// When close websocker connection (user close browser or unexpected error occurred)
websocket.onclose = function (event) {
    // ...
};

// When user is typing a message
let sendTypingEvent = () => {
    const now = Date.now();
    if (now - lastTypingSentAt < throttleIntervalTime) {
        return;
    }

    lastTypingSentAt = now;

    const data = JSON.stringify({
        type: 'typing',
        content: messageInput.value ?? '',
    });

    websocket.send(data);
};

messageInput.addEventListener('input', sendTypingEvent);

// When user send a message to server
formMessage.onsubmit = (event) => {
    event.preventDefault();

    if (messageInput.value === '') {
        return;
    }

    const data = JSON.stringify({
        type: 'message',
        content: messageInput.value,
    });

    websocket.send(data);
    appendMessage({
        type: 'message',
        pid: pid,
        username: username,
        content: messageInput.value,
        datetime: new Date().toLocaleTimeString(),
    });
    messageInput.value = '';
};

const appendMessage = (data) => {
    const wrapper = document.createElement('div');
    wrapper.className = `flex gap-y-0.5 flex-col w-full pb-6 ${data.pid === pid ? 'items-end' : 'items-start'}`;

    const metadata = document.createElement('p');
    metadata.className = 'text-[0.65rem] text-secondary-500';
    metadata.textContent = `${data.username} • ${data.datetime}`;

    const bubble = document.createElement('p');
    bubble.className = `text-sm font-medium rounded-md px-3 py-2 gap-y-2 max-w-96 ${data.pid === pid ? 'bg-primary-600/35' : 'bg-secondary-600/35'}`;
    bubble.textContent = data.content;

    wrapper.append(metadata, bubble);
    messagesList.append(wrapper);

    app.scrollTo(0, messagesList.scrollHeight);
};
