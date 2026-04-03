import { partitionPresenceUsers } from '../lib/presence.mjs';

export default function UserPresenceList({ currentPid, onlineUsers }) {
  const { currentUser, otherUsers } = partitionPresenceUsers(onlineUsers, currentPid);

  return (
    <aside className="flex w-full flex-col rounded-xl border border-secondary-800 bg-secondary-900/30 p-5 lg:fixed lg:inset-y-0 lg:right-0 lg:w-96 lg:rounded-none lg:border-y-0 lg:border-r-0">
      <div className="flex items-center justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold uppercase tracking-[0.2em] text-secondary-300">Online now</h2>
          <p className="mt-1 text-xs text-secondary-500">Connected users visible in this room.</p>
        </div>
        <div
          className="rounded-full border border-secondary-800 bg-secondary-500/10 px-3 py-1 text-xs font-semibold text-secondary-300"
          aria-label={`${onlineUsers.length} users online`}
        >
          {onlineUsers.length}
        </div>
      </div>

      <div className="mt-5 flex min-h-0 flex-1 flex-col space-y-3 lg:overflow-hidden">
        {currentUser ? (
          <div className="rounded-lg border border-primary-500/30 bg-primary-500/10 p-3">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary-300">You</p>
            <p className="mt-1 text-sm font-medium text-white">{currentUser.username}</p>
          </div>
        ) : null}

        <div className="min-h-0 flex-1 space-y-2 overflow-y-auto pr-1">
          {otherUsers.map((user) => (
            <div
              key={user.pid}
              className="flex items-center justify-between rounded-lg border border-secondary-800 bg-secondary-950/60 px-3 py-2"
            >
              <span className="text-sm text-secondary-100">{user.username}</span>
              <span className="h-2.5 w-2.5 rounded-full bg-emerald-400" aria-hidden="true" />
            </div>
          ))}

          {otherUsers.length === 0 ? (
            <p className="rounded-lg border border-dashed border-secondary-800 px-3 py-4 text-sm text-secondary-500">
              No other users are online yet.
            </p>
          ) : null}
        </div>
      </div>
    </aside>
  );
}
