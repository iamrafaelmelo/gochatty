import { useEffect, useRef, useState } from 'react';

import {
  MESSAGE_TYPE_MESSAGE,
  MESSAGE_TYPE_PRESENCE,
  MESSAGE_TYPE_SETUP,
  MESSAGE_TYPE_TYPING,
  TYPING_CLEAR_TIMEOUT_MS,
} from '../types/chat';

function resolveWebSocketUrl() {
  const configuredUrl = import.meta.env.APP_WEBSOCKET_URL?.trim();

  if (!configuredUrl) {
    throw new Error('APP_WEBSOCKET_URL is required and must include the ws:// or wss:// protocol and host.');
  }

  let parsedUrl;
  try {
    parsedUrl = new URL(configuredUrl);
  } catch {
    throw new Error('APP_WEBSOCKET_URL must be a valid absolute websocket URL including protocol and host.');
  }

  if ((parsedUrl.protocol !== 'ws:' && parsedUrl.protocol !== 'wss:') || parsedUrl.host === '') {
    throw new Error('APP_WEBSOCKET_URL must use ws:// or wss:// and include a host.');
  }

  return parsedUrl.toString();
}

export function useChatWebSocket() {
  const websocketRef = useRef(null);
  const typingClearTimeoutRef = useRef(null);

  const [messages, setMessages] = useState([]);
  const [onlineUsers, setOnlineUsers] = useState([]);
  const [typingText, setTypingText] = useState('');
  const [currentUser, setCurrentUser] = useState({ pid: null, username: null });

  useEffect(() => {
    const websocket = new WebSocket(resolveWebSocketUrl());
    websocketRef.current = websocket;

    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);

      switch (data.type) {
        case MESSAGE_TYPE_SETUP:
          setCurrentUser({ pid: data.pid, username: data.username });
          break;
        case MESSAGE_TYPE_PRESENCE:
          setOnlineUsers(Array.isArray(data.users) ? data.users : []);
          break;
        case MESSAGE_TYPE_MESSAGE:
          setMessages((previousMessages) => [...previousMessages, data]);
          break;
        case MESSAGE_TYPE_TYPING:
          setTypingText(data.content);
          clearTimeout(typingClearTimeoutRef.current);
          typingClearTimeoutRef.current = setTimeout(() => {
            setTypingText('');
          }, TYPING_CLEAR_TIMEOUT_MS);
          break;
        default:
          break;
      }
    };

    websocket.onclose = () => {
      // noop: preserve legacy behavior.
    };

    return () => {
      clearTimeout(typingClearTimeoutRef.current);
      websocket.close();
    };
  }, []);

  const sendMessage = (content) => {
    if (content === '') {
      return;
    }

    websocketRef.current?.send(JSON.stringify({
      type: MESSAGE_TYPE_MESSAGE,
      content,
    }));

    setMessages((previousMessages) => [
      ...previousMessages,
      {
        type: MESSAGE_TYPE_MESSAGE,
        pid: currentUser.pid,
        username: currentUser.username,
        content,
        datetime: new Date().toLocaleTimeString(),
      },
    ]);
  };

  const sendTyping = (content) => {
    if (!websocketRef.current || !content) {
      return;
    }

    websocketRef.current.send(JSON.stringify({
      type: MESSAGE_TYPE_TYPING,
      content,
    }));
  };

  return {
    currentUser,
    messages,
    onlineUsers,
    sendMessage,
    sendTyping,
    typingText,
  };
}
