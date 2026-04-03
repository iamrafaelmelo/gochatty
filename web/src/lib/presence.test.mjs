import test from 'node:test';
import assert from 'node:assert/strict';

import { partitionPresenceUsers } from './presence.mjs';

test('partitionPresenceUsers pins the current user and preserves other online users', () => {
  const result = partitionPresenceUsers([
    { pid: 'b', username: 'Bob' },
    { pid: 'a', username: 'Alice' },
    { pid: 'c', username: 'Carla' },
  ], 'a');

  assert.deepEqual(result.currentUser, { pid: 'a', username: 'Alice' });
  assert.deepEqual(result.otherUsers, [
    { pid: 'b', username: 'Bob' },
    { pid: 'c', username: 'Carla' },
  ]);
});

test('partitionPresenceUsers returns no current user when setup has not arrived yet', () => {
  const result = partitionPresenceUsers([{ pid: 'b', username: 'Bob' }], null);

  assert.equal(result.currentUser, null);
  assert.deepEqual(result.otherUsers, [{ pid: 'b', username: 'Bob' }]);
});
