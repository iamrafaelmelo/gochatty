import MessageInput from './MessageInput';
import MessageList from './MessageList';

export default function ChatLayout({ appRef, currentPid, messages, onSendMessage, onTyping, typingText }) {
  return (
    <div className="flex min-h-screen w-full flex-col gap-y-10 bg-secondary-950 p-8 text-white">
      <header className="flex flex-col gap-y-1 text-center">
        <h1 className="text-xl font-bold">Go Chatty</h1>
        <small className="text-secondary-500">A simple Golang websocket chat</small>
      </header>

      <main
        className="mx-auto flex w-full p-6 max-w-4xl flex-1 flex-col rounded-xl border border-secondary-800 overflow-y-auto overflow-x-hidden relative"
        id="app"
        ref={appRef}
      >
        <MessageList currentPid={currentPid} messages={messages} />
      </main>

      <footer className="space-y-3 text-center">
        <MessageInput onSendMessage={onSendMessage} onTyping={onTyping} typingText={typingText} />
        <div className="flex items-center gap-x-2 text-xs text-secondary-600 justify-center">
          <span>
            Powered by{' '}
            <a className="underline hover:no-underline" href="https://gofiber.io" target="_blank" rel="noreferrer">
              Fiber
            </a>
          </span>
          <span>&bullet;</span>
          <a
            className="underline hover:no-underline"
            href="https://github.com/iamrafaelmelo/gochatty"
            target="_blank"
            rel="noreferrer"
          >
            Source code
          </a>
        </div>
      </footer>
    </div>
  );
}
