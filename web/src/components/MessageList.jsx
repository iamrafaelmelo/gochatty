import { memo } from 'react';

import MessageBubble from './MessageBubble';

function MessageList({ currentPid, messages }) {
  return (
    <div className="h-full max-h-96 space-y-1 flex flex-col" id="messages-list">
      {messages.map((message, index) => (
        <MessageBubble
          key={`${message.pid}-${message.datetime}-${index}`}
          isOwnMessage={message.pid === currentPid}
          message={message}
        />
      ))}
    </div>
  );
}

export default memo(MessageList);
