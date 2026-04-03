export function partitionPresenceUsers(users, currentPid) {
  const normalizedUsers = Array.isArray(users) ? users : [];
  const currentUser = normalizedUsers.find((user) => user.pid === currentPid) ?? null;
  const otherUsers = normalizedUsers.filter((user) => user.pid !== currentPid);

  return {
    currentUser,
    otherUsers,
  };
}
