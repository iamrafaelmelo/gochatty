const form = document.getElementById('form');
const message = document.getElementById('message');
const list = document.getElementById('list');
const main = document.getElementById('app');

const websocket = new WebSocket(`ws://${document.location.host}/ws`);

websocket.onmessage = function (event) {
    console.log(`Message received from server: ${event.data}`);
    const data = JSON.parse(event.data);

    switch (data.type) {
        case 'message':
            append(data);
            break;
        case 'typying':
            console.log('typing...');
            break;
    }
};

websocket.onclose = function (event) {
    console.log('Connection closed!');
};

form.onsubmit = (event) => {
    event.preventDefault();
    const data = JSON.stringify({
        type: 'message',
        message: message.value,
        datetime: new Date().toLocaleTimeString(),
    });

    console.log(`Sending to server: ${data}`);
    websocket.send(data);
    append(JSON.parse(data));
    message.value = '';
};


const append = (data) => {
    list.insertAdjacentHTML('beforeend', `
        <div class="flex gap-y-0.5 flex-col w-full items-end">
            <p class="text-[0.65rem] text-secondary-500">You ${data.datetime}</p>
            <p class="text-sm font-medium bg-primary-600/35 rounded-md px-3 py-2 gap-y-2 max-w-96">${data.message}</p>
        </div>
    `);

    main.scrollTo(0, list.scrollHeight);
};
