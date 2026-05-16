'use strict';

const ARENA = { x: 30, y: 30, w: 620, h: 420 };

// Physics tuning — lighter, floatier, more chaotic
const DAMPING     = 0.992;  // was 0.97 — guns keep momentum much longer
const WALL_BOUNCE = 0.82;   // was 0.55 — snappy bounces
const GRAVITY     = 0.04;   // gentle downward pull
const GUN_RADIUS  = 32;
const BULLET_DECAY = 0.010;

const WEAPON_DEFS = {
  pistol: {
    label: 'PULSAR-9', type: 'pistol',
    fireRate: 20, recoilForce: 3.6, bulletSpeed: 9,
    bulletRadius: 2, pellets: 1, spread: 0.06, color: '#4af0c8',
    desc: 'fast fire · floaty kicks · high mobility',
  },
  shotgun: {
    label: 'SLEDGE-X', type: 'shotgun',
    fireRate: 58, recoilForce: 10, bulletSpeed: 7,
    bulletRadius: 3, pellets: 4, spread: 0.38, color: '#f0a84a',
    desc: 'massive kick · spread shot · slow rate',
  },
  smg: {
    label: 'WASP-7', type: 'smg',
    fireRate: 8, recoilForce: 1.8, bulletSpeed: 11,
    bulletRadius: 1.5, pellets: 1, spread: 0.10, color: '#a84af0',
    desc: 'rapid micro-kicks · constant drift · fast bullets',
  },
  sniper: {
    label: 'LANCE-1', type: 'sniper',
    fireRate: 90, recoilForce: 16, bulletSpeed: 18,
    bulletRadius: 2, pellets: 1, spread: 0.01, color: '#f04a4a',
    desc: 'one giant kick · extreme range · high damage',
  },
};

let bulletIdCounter = 0;

function createPlayer(id, weaponType, startX, startY) {
  const def = WEAPON_DEFS[weaponType] || WEAPON_DEFS.pistol;
  return {
    id, weaponType,
    label: def.label, color: def.color,
    x: startX, y: startY,
    vx: (Math.random() - 0.5) * 1.5,
    vy: (Math.random() - 0.5) * 1.5,
    angle: 0, hp: 100, alive: true,
    fireTimer: Math.floor(def.fireRate * 0.5),
    fireRate: def.fireRate, recoilForce: def.recoilForce,
    bulletSpeed: def.bulletSpeed, bulletRadius: def.bulletRadius,
    pellets: def.pellets, spread: def.spread,
    muzzleFlash: 0, score: 0,
  };
}

function muzzlePos(player) {
  const len = { pistol:32, shotgun:40, smg:26, sniper:52 }[player.weaponType] || 32;
  return { x: player.x + Math.cos(player.angle)*len, y: player.y + Math.sin(player.angle)*len };
}

function pseudoRandom(seed) {
  const x = Math.sin(seed + 1) * 10000;
  return x - Math.floor(x);
}

function spawnBullets(player, bullets) {
  const mp = muzzlePos(player);
  for (let i = 0; i < player.pellets; i++) {
    const a = player.angle + (pseudoRandom(bulletIdCounter) - 0.5) * player.spread;
    bullets.push({
      id: bulletIdCounter++, ownerId: player.id,
      x: mp.x, y: mp.y,
      vx: Math.cos(a) * player.bulletSpeed,
      vy: Math.sin(a) * player.bulletSpeed,
      life: 1, r: player.bulletRadius,
    });
  }
  player.vx -= Math.cos(player.angle) * player.recoilForce;
  player.vy -= Math.sin(player.angle) * player.recoilForce;
  player.muzzleFlash = 1;
}

function tickPhysics(state, inputs) {
  const { players, bullets } = state;
  const playerList = Object.values(players);
  const events = [];

  playerList.forEach(player => {
    if (!player.alive) return;
    const input = inputs[player.id] || {};
    if (typeof input.angle === 'number') player.angle = input.angle;

    player.fireTimer++;
    if (input.shoot && player.fireTimer >= player.fireRate) {
      spawnBullets(player, bullets);
      player.fireTimer = 0;
      events.push({ type: 'shoot', playerId: player.id });
    }

    player.muzzleFlash = Math.max(0, player.muzzleFlash - 0.10);
    player.vy += GRAVITY;
    player.vx *= DAMPING;
    player.vy *= DAMPING;
    player.x += player.vx;
    player.y += player.vy;

    const minX = ARENA.x + GUN_RADIUS, maxX = ARENA.x + ARENA.w - GUN_RADIUS;
    const minY = ARENA.y + GUN_RADIUS, maxY = ARENA.y + ARENA.h - GUN_RADIUS;
    if (player.x < minX) { player.x = minX; player.vx = Math.abs(player.vx) * WALL_BOUNCE; events.push({ type: 'wall_bounce', playerId: player.id }); }
    if (player.x > maxX) { player.x = maxX; player.vx = -Math.abs(player.vx) * WALL_BOUNCE; events.push({ type: 'wall_bounce', playerId: player.id }); }
    if (player.y < minY) { player.y = minY; player.vy = Math.abs(player.vy) * WALL_BOUNCE; events.push({ type: 'wall_bounce', playerId: player.id }); }
    if (player.y > maxY) { player.y = maxY; player.vy = -Math.abs(player.vy) * WALL_BOUNCE; events.push({ type: 'wall_bounce', playerId: player.id }); }
  });

  for (let i = bullets.length - 1; i >= 0; i--) {
    const b = bullets[i];
    b.x += b.vx; b.y += b.vy; b.life -= BULLET_DECAY;
    if (b.x < ARENA.x)           { b.vx *= -0.9; b.x = ARENA.x; }
    if (b.x > ARENA.x + ARENA.w) { b.vx *= -0.9; b.x = ARENA.x + ARENA.w; }
    if (b.y < ARENA.y)           { b.vy *= -0.9; b.y = ARENA.y; }
    if (b.y > ARENA.y + ARENA.h) { b.vy *= -0.9; b.y = ARENA.y + ARENA.h; }

    playerList.forEach(player => {
      if (!player.alive || b.ownerId === player.id || b.life <= 0) return;
      const dx = b.x - player.x, dy = b.y - player.y;
      if (Math.sqrt(dx*dx + dy*dy) < 22 + b.r) {
        const dmg = { pistol:12, shotgun:18, smg:7, sniper:40 }[player.weaponType] || 10;
        player.hp -= dmg;
        player.vx += b.vx * 0.55;
        player.vy += b.vy * 0.55;
        b.life = -1;
        events.push({ type: 'hit', playerId: player.id, damage: dmg, hp: Math.max(0, player.hp) });
        if (player.hp <= 0) {
          player.alive = false; player.hp = 0;
          const killer = players[b.ownerId];
          if (killer) killer.score++;
          events.push({ type: 'death', playerId: player.id, killerId: b.ownerId });
        }
      }
    });
    if (b.life <= 0) bullets.splice(i, 1);
  }
  return events;
}

function respawnPlayer(player, x, y) {
  player.x = x; player.y = y;
  player.vx = (Math.random() - 0.5) * 2;
  player.vy = (Math.random() - 0.5) * 2;
  player.hp = 100; player.alive = true;
  player.fireTimer = Math.floor(player.fireRate * 0.4);
}

module.exports = { createPlayer, tickPhysics, respawnPlayer, ARENA, WEAPON_DEFS };