import MessageInput from './MessageInput';
import MessageList from './MessageList';
import UserPresenceList from './UserPresenceList';

export default function ChatLayout({ appRef, currentPid, messages, onlineUsers, onSendMessage, onTyping, typingText }) {
  return (
    <div className="flex min-h-screen w-full flex-col gap-y-10 bg-secondary-950 p-8 text-white lg:pr-[26rem]">
      <header className="flex flex-col gap-y-1 text-center">
        <h1 className="text-xl font-bold">Go Chatty</h1>
        <small className="text-secondary-500">A simple Golang websocket chat</small>
      </header>

      <section className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6">
        <main
          className="flex min-h-[26rem] flex-1 flex-col overflow-y-auto overflow-x-hidden rounded-xl border border-secondary-800 p-6"
          id="app"
          ref={appRef}
        >
          <MessageList currentPid={currentPid} messages={messages} />
        </main>
      </section>

      <UserPresenceList currentPid={currentPid} onlineUsers={onlineUsers} />

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
