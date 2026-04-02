import { useEffect, useRef } from 'react';

import ChatLayout from './components/ChatLayout';
import { useChatWebSocket } from './hooks/useChatWebSocket';
import { useTypingThrottle } from './hooks/useTypingThrottle';

export default function App() {
  const appRef = useRef(null);
  const { currentUser, messages, sendMessage, sendTyping, typingText } = useChatWebSocket();
  const { sendThrottledTyping } = useTypingThrottle(sendTyping);

  useEffect(() => {
    if (!appRef.current) {
      return;
    }

    appRef.current.scrollTo(0, appRef.current.scrollHeight);
  }, [messages]);

  return (
    <ChatLayout
      appRef={appRef}
      currentPid={currentUser.pid}
      messages={messages}
      onSendMessage={sendMessage}
      onTyping={sendThrottledTyping}
      typingText={typingText}
    />
  );
}
