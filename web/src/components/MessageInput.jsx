import { useState } from 'react';

import TypingIndicator from './TypingIndicator';

export default function MessageInput({ onSendMessage, onTyping, typingText }) {
  const [messageContent, setMessageContent] = useState('');

  const handleSubmit = (event) => {
    event.preventDefault();

    if (messageContent === '') {
      return;
    }

    onSendMessage(messageContent);
    setMessageContent('');
  };

  const handleInputChange = (event) => {
    const nextMessage = event.target.value;
    setMessageContent(nextMessage);
    onTyping(nextMessage);
  };

  return (
    <form className="mx-auto w-full max-w-2xl flex flex-col gap-y-2" id="form-message" onSubmit={handleSubmit}>
      <input
        className="w-full rounded-lg text-sm placeholder-secondary-600/85 border-secondary-800 bg-secondary-900 shadow-md focus:border-primary-600 focus:outline-none focus:ring-4 focus:ring-primary-500/15"
        type="text"
        id="message-input"
        placeholder="Type message here... (Press 'Enter' to send a message)"
        value={messageContent}
        onInput={handleInputChange}
      />
      <TypingIndicator text={typingText} />
    </form>
  );
}
