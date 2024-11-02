const form = document.getElementById('form');
const message = document.getElementById('message');
const list = document.getElementById('list');
const main = document.getElementById('app');

const websocket = new WebSocket(`ws://${document.location.host}/ws`);
var pid = null;

websocket.onmessage = function (event) {
    const data = JSON.parse(event.data);

    switch (data.type) {
        case 'setup':
            pid = data.pid
            console.log(pid)
            break;
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
        pid: pid,
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
        <div class="flex gap-y-0.5 flex-col w-full pb-6 ${data.pid === pid ? 'items-end' : 'items-start'}">
            <p class="text-[0.65rem] text-secondary-500">${data.pid === pid ? 'You' : 'Other'} ${data.datetime}</p>
            <p class="text-sm font-medium ${data.pid === pid ? 'bg-primary-600/35' : 'bg-secondary-600/35'} rounded-md px-3 py-2 gap-y-2 max-w-96">
                ${data.message}
            </p>
        </div>
    `);

    main.scrollTo(0, list.scrollHeight);
};
