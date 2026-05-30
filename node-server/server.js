'use strict';

const { WebSocketServer } = require('ws');
const { createPlayer, tickPhysics, respawnPlayer, ARENA, WEAPON_DEFS } = require('./physics');

const PORT          = 3000;
const TICK_RATE     = 30;
const TICK_MS       = 1000 / TICK_RATE;
const RESPAWN_DELAY = 2000;

const wss = new WebSocketServer({ port: PORT });

// rooms: { [roomCode]: RoomObject }
const rooms = {};

function send(ws, msg) {
  if (ws.readyState === 1) ws.send(JSON.stringify(msg));
}

function broadcastAll(room, msg) {
  const data = JSON.stringify(msg);
  room.clients.forEach(id => {
    const ws = room.sockets[id];
    if (ws && ws.readyState === 1) ws.send(data);
  });
}

function buildSnapshot(room) {
  return {
    type: 'snapshot',
    players: Object.values(room.players).map(p => ({
      id: p.id, weaponType: p.weaponType, label: p.label, color: p.color,
      x:  Math.round(p.x  * 10) / 10,
      y:  Math.round(p.y  * 10) / 10,
      vx: Math.round(p.vx * 100) / 100,
      vy: Math.round(p.vy * 100) / 100,
      angle: Math.round(p.angle * 1000) / 1000,
      hp: p.hp, alive: p.alive,
      muzzleFlash: Math.round(p.muzzleFlash * 10) / 10,
      score: p.score,
    })),
    bullets: room.bullets.map(b => ({
      id: b.id, ownerId: b.ownerId,
      x:  Math.round(b.x  * 10) / 10,
      y:  Math.round(b.y  * 10) / 10,
      vx: Math.round(b.vx * 10) / 10,
      vy: Math.round(b.vy * 10) / 10,
      r: b.r,
    })),
  };
}

function generateRoomCode() {
  const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  let code = '';
  for (let i = 0; i < 4; i++) code += chars[Math.floor(Math.random() * chars.length)];
  return code;
}

function createRoom(code) {
  rooms[code] = {
    code,
    clients: [],
    sockets: {},
    players: {},
    weapons: {},        // playerId -> chosen weaponType
    bullets: [],
    inputs:  {},
    phase:   'waiting', // waiting | playing
    tickInterval:   null,
    respawnTimers:  {},
  };
  console.log(`[${code}] room created`);
  return rooms[code];
}

function roomList() {
  return Object.values(rooms)
    .filter(r => r.phase === 'waiting' && r.clients.length < 2)
    .map(r => ({ code: r.code, players: r.clients.length }));
}

function startMatch(room) {
  room.phase = 'playing';
  const [id1, id2] = room.clients;
  const spawnPositions = {
    [id1]: { x: ARENA.x + 140, y: ARENA.y + ARENA.h / 2 },
    [id2]: { x: ARENA.x + ARENA.w - 140, y: ARENA.y + ARENA.h / 2 },
  };

  room.clients.forEach(id => {
    const weapon = room.weapons[id] || 'pistol';
    const sp = spawnPositions[id];
    room.players[id] = createPlayer(id, weapon, sp.x, sp.y);
  });

  broadcastAll(room, {
    type: 'match_start',
    players: Object.values(room.players).map(p => ({
      id: p.id, weaponType: p.weaponType, label: p.label, color: p.color, score: p.score,
    })),
  });

  room.tickInterval = setInterval(() => {
    const events = tickPhysics({ players: room.players, bullets: room.bullets }, room.inputs);

    events.forEach(ev => {
      if (ev.type === 'death') {
        broadcastAll(room, { type: 'player_died', playerId: ev.playerId, killerId: ev.killerId });
        const spawnIdx = room.clients.indexOf(ev.playerId);
        const spawnX   = spawnIdx === 0 ? ARENA.x + 140 : ARENA.x + ARENA.w - 140;
        room.respawnTimers[ev.playerId] = setTimeout(() => {
          if (!room.players[ev.playerId]) return;
          respawnPlayer(room.players[ev.playerId], spawnX, ARENA.y + ARENA.h / 2);
          room.bullets = room.bullets.filter(b => b.ownerId !== ev.playerId);
          broadcastAll(room, { type: 'player_respawned', playerId: ev.playerId });
        }, RESPAWN_DELAY);
      }
      if (ev.type === 'hit') {
        broadcastAll(room, { type: 'hit', playerId: ev.playerId, damage: ev.damage, hp: ev.hp });
      }
    });

    broadcastAll(room, buildSnapshot(room));
    room.inputs = {};
  }, TICK_MS);

  console.log(`[${room.code}] match started: ${room.clients.map(id => `${id}(${room.weapons[id]})`).join(' vs ')}`);
}

function cleanupRoom(room) {
  if (room.tickInterval) clearInterval(room.tickInterval);
  Object.values(room.respawnTimers).forEach(clearTimeout);
  delete rooms[room.code];
  console.log(`[${room.code}] room cleaned up`);
}

// ── WebSocket connections ──────────────────────────────────────────────────

wss.on('connection', (ws) => {
  let playerId   = null;
  let currentRoom = null;
  let pingStart  = 0;

  send(ws, { type: 'hello', weapons: WEAPON_DEFS, rooms: roomList() });

  ws.on('message', (raw) => {
    let msg;
    try { msg = JSON.parse(raw); } catch { return; }

    // ── ping ──────────────────────────────────────────────────────────────
    if (msg.type === 'ping') {
      send(ws, { type: 'pong' });
      return;
    }

    // ── list rooms (lobby refresh) ────────────────────────────────────────
    if (msg.type === 'list_rooms') {
      send(ws, { type: 'room_list', rooms: roomList() });
      return;
    }

    // ── create room ───────────────────────────────────────────────────────
    if (msg.type === 'create_room') {
      if (currentRoom) return;
      playerId  = `p${Date.now()}`;
      ws.playerId = playerId;

      let code;
      do { code = generateRoomCode(); } while (rooms[code]);
      currentRoom = createRoom(code);

      const weapon = WEAPON_DEFS[msg.weapon] ? msg.weapon : 'pistol';
      currentRoom.clients.push(playerId);
      currentRoom.sockets[playerId] = ws;
      currentRoom.weapons[playerId] = weapon;
      currentRoom.inputs[playerId]  = {};
      ws.roomCode = code;

      send(ws, { type: 'room_created', code, playerId, weapon, waitingForOpponent: true });
      console.log(`[${code}] created by ${playerId} (${weapon})`);
      return;
    }

    // ── join room ─────────────────────────────────────────────────────────
    if (msg.type === 'join_room') {
      if (currentRoom) return;
      const code = (msg.code || '').toUpperCase().trim();
      const room = rooms[code];

      if (!room) { send(ws, { type: 'error', message: 'Room not found.' }); return; }
      if (room.clients.length >= 2) { send(ws, { type: 'error', message: 'Room is full.' }); return; }
      if (room.phase !== 'waiting') { send(ws, { type: 'error', message: 'Match already started.' }); return; }

      playerId = `p${Date.now()}`;
      ws.playerId = playerId;
      ws.roomCode = code;
      currentRoom = room;

      const weapon = WEAPON_DEFS[msg.weapon] ? msg.weapon : 'pistol';
      room.clients.push(playerId);
      room.sockets[playerId] = ws;
      room.weapons[playerId] = weapon;
      room.inputs[playerId]  = {};

      send(ws, { type: 'room_joined', code, playerId, weapon });
      console.log(`[${code}] joined by ${playerId} (${weapon}) — ${room.clients.length}/2`);

      if (room.clients.length === 2) startMatch(room);
      return;
    }

    // ── game input ────────────────────────────────────────────────────────
    if (msg.type === 'input' && currentRoom && playerId) {
      currentRoom.inputs[playerId] = {
        angle: typeof msg.angle === 'number' ? msg.angle : undefined,
        shoot: !!msg.shoot,
      };
      return;
    }
  });

  ws.on('close', () => {
    if (!currentRoom || !playerId) return;
    console.log(`[${currentRoom.code}] ${playerId} disconnected`);
    const other = currentRoom.clients.find(id => id !== playerId);
    if (other && currentRoom.sockets[other]) {
      send(currentRoom.sockets[other], { type: 'opponent_disconnected' });
    }
    cleanupRoom(currentRoom);
    currentRoom = null;
  });

  ws.on('error', err => console.error('ws error:', err.message));
});

console.log(`✦ Recoil Arena server  →  ws://localhost:${PORT}`);
console.log(`  Tick rate: ${TICK_RATE}/sec | Weapons: ${Object.keys(WEAPON_DEFS).join(', ')}`);