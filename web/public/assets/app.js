const formMessage = document.getElementById('form-message');
const messageInput = document.getElementById('message-input');
const messagesList = document.getElementById('messages-list');
const app = document.getElementById('app');
const typingMessage = document.getElementById('typing');
const websocket = new WebSocket(`ws://${document.location.host}/ws`);

// Controls user typing intervals
const throttleIntervalTime = 200;
const clearIntervalTime = 900;
var canSendMessage = false;
var clearTimerId = null;

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
let keypress = (event) => {
    const keypressed = event.keyCode;
    let isAlphabetLetterKey = (keypressed >= 65 && keypressed || keypressed >= 97 && keypressed <= 122);
    let isNumberKey = (keypressed >= 48 && keypressed <= 57);

    if (!isAlphabetLetterKey && !isNumberKey) {
        return;
    }

    if (!canSendMessage) {
        setTimeout(function () {
            canSendMessage = true;
        }, throttleIntervalTime);

        return;
    }

    const data = JSON.stringify({
        pid: pid,
        type: 'typing',
        username: username,
        content: messageInput.value ?? '',
        datetime: new Date().toLocaleTimeString()
    });

    websocket.send(data);
};

messageInput.onkeyup = keypress;
messageInput.onkeydown = keypress;

// When user send a message to server
formMessage.onsubmit = (event) => {
    event.preventDefault();

    if (messageInput.value === '') {
        return;
    }

    const data = JSON.stringify({
        pid: pid,
        type: 'message',
        username: username,
        content: messageInput.value,
        datetime: new Date().toLocaleTimeString(),
    });

    websocket.send(data);
    appendMessage(JSON.parse(data));
    messageInput.value = '';
};

const appendMessage = (data) => {
    messagesList.insertAdjacentHTML('beforeend', `
        <div class="flex gap-y-0.5 flex-col w-full pb-6 ${data.pid === pid ? 'items-end' : 'items-start'}">
            <p class="text-[0.65rem] text-secondary-500">${data.username} &bullet; ${data.datetime}</p>
            <p class="text-sm font-medium ${data.pid === pid ? 'bg-primary-600/35' : 'bg-secondary-600/35'} rounded-md px-3 py-2 gap-y-2 max-w-96">
                ${data.content}
            </p>
        </div>
    `);

    app.scrollTo(0, messagesList.scrollHeight);
};
