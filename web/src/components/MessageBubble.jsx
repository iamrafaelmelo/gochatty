export default function MessageBubble({ isOwnMessage, message }) {
  return (
    <div className={`flex gap-y-0.5 flex-col w-full pb-6 ${isOwnMessage ? 'items-end' : 'items-start'}`}>
      <p className="text-[0.65rem] text-secondary-500">{message.username} • {message.datetime}</p>
      <p
        className={`text-sm font-medium rounded-md px-3 py-2 gap-y-2 max-w-96 ${
          isOwnMessage ? 'bg-primary-600/35' : 'bg-secondary-600/35'
        }`}
      >
        {message.content}
      </p>
    </div>
  );
}
