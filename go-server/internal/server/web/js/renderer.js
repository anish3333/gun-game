import { state, ARENA, COLORS, getPlayer } from './state.js';

const canvas = document.getElementById('arena');
const ctx    = canvas.getContext('2d');
const W = 680, H = 480;

let sparks     = [];
let hitNumbers = [];

// ── public effect spawners ────────────────────────────────────────────────

export function spawnSparks(playerId, damage) {
  const p = getPlayer(playerId);
  if (!p) return;
  const color = COLORS[p.weaponType] || '#4af0c8';
  const count = Math.min(Math.ceil(damage / 2), 10);
  for (let i = 0; i < count; i++) {
    const a = Math.random() * Math.PI * 2;
    const sp = Math.random() * 3 + 1;
    sparks.push({ x: p.x, y: p.y, vx: Math.cos(a)*sp, vy: Math.sin(a)*sp, life: 1, color });
  }
}

export function spawnDeathEffect(playerId) {
  const p = getPlayer(playerId);
  if (!p) return;
  const color = COLORS[p.weaponType] || '#4af0c8';
  for (let i = 0; i < 24; i++) {
    const a = Math.random() * Math.PI * 2;
    const sp = Math.random() * 5 + 2;
    sparks.push({ x: p.x, y: p.y, vx: Math.cos(a)*sp, vy: Math.sin(a)*sp, life: 1.5, color });
  }
}

export function spawnHitNumber(playerId, damage) {
  const p = getPlayer(playerId);
  if (!p) return;
  hitNumbers.push({
    x:    p.x + (Math.random() - 0.5) * 20,
    y:    p.y - 20,
    vy:   -1.5,
    text: '-' + damage,
    life: 1,
    color: playerId === state.myId ? '#f04a4a' : '#f0c84a',
  });
}

// ── arena background ──────────────────────────────────────────────────────

function drawArena() {
  ctx.fillStyle = '#060911';
  ctx.fillRect(0, 0, W, H);

  // Grid
  ctx.strokeStyle = '#1a2535';
  ctx.lineWidth = 0.5;
  const gs = 40;
  for (let x = ARENA.x; x <= ARENA.x + ARENA.w; x += gs) {
    ctx.beginPath(); ctx.moveTo(x, ARENA.y); ctx.lineTo(x, ARENA.y + ARENA.h); ctx.stroke();
  }
  for (let y = ARENA.y; y <= ARENA.y + ARENA.h; y += gs) {
    ctx.beginPath(); ctx.moveTo(ARENA.x, y); ctx.lineTo(ARENA.x + ARENA.w, y); ctx.stroke();
  }

  // Border
  ctx.strokeStyle = '#2a3a50';
  ctx.lineWidth = 1.5;
  ctx.strokeRect(ARENA.x, ARENA.y, ARENA.w, ARENA.h);

  // Corner decorations
  ctx.strokeStyle = '#4af0c830';
  ctx.lineWidth = 0.5;
  const cs = 14;
  [
    [ARENA.x,            ARENA.y,            1,  1],
    [ARENA.x + ARENA.w,  ARENA.y,           -1,  1],
    [ARENA.x,            ARENA.y + ARENA.h,  1, -1],
    [ARENA.x + ARENA.w,  ARENA.y + ARENA.h, -1, -1],
  ].forEach(([cx, cy, sx, sy]) => {
    ctx.beginPath();
    ctx.moveTo(cx + sx * cs, cy);
    ctx.lineTo(cx, cy);
    ctx.lineTo(cx, cy + sy * cs);
    ctx.stroke();
  });
}

// ── gun drawing ───────────────────────────────────────────────────────────

function drawGun(p) {
  const color = COLORS[p.weaponType] || '#4af0c8';
  const dim   = {
    pistol:  '#1a6050',
    shotgun: '#604010',
    smg:     '#502070',
    sniper:  '#601010',
  }[p.weaponType] || '#1a6050';

  ctx.save();
  ctx.translate(p.x, p.y);
  ctx.rotate(p.angle);
  if (!p.alive) ctx.globalAlpha = 0.2;

  switch (p.weaponType) {
    case 'pistol':   drawPistol(color, dim, p.muzzleFlash);   break;
    case 'shotgun':  drawShotgun(color, dim, p.muzzleFlash);  break;
    case 'smg':      drawSMG(color, dim, p.muzzleFlash);      break;
    case 'sniper':   drawSniper(color, dim, p.muzzleFlash);   break;
    default:         drawPistol(color, dim, p.muzzleFlash);
  }

  ctx.restore();
}

function drawMuzzleFlash(bx, f, color) {
  if (f <= 0.1) return;
  ctx.save();
  ctx.globalAlpha = f * 0.9;
  ctx.fillStyle = '#fff8e0';
  ctx.beginPath(); ctx.ellipse(bx + 14*f, 0, 18*f, 7*f, 0, 0, Math.PI*2); ctx.fill();
  ctx.fillStyle = color;
  ctx.beginPath(); ctx.ellipse(bx + 8*f, 0, 10*f, 4*f, 0, 0, Math.PI*2); ctx.fill();
  ctx.restore();
}

function drawPistol(color, dim, flash) {
  ctx.scale(0.55, 0.55);
  ctx.fillStyle = '#0d1520';
  ctx.beginPath(); ctx.roundRect(-24,-16,80,22,3); ctx.fill();
  ctx.fillStyle = '#1e2d3f'; ctx.fillRect(-22,-16,76,5);
  ctx.fillStyle = dim;       ctx.fillRect(8,-13,28,7);
  ctx.fillStyle = '#0d1520'; ctx.fillRect(-24,-8,62,15);
  ctx.fillStyle = '#0a1018'; ctx.fillRect(-22,7,46,4);
  ctx.strokeStyle = '#1e2d3f'; ctx.lineWidth = 1.5;
  ctx.beginPath(); ctx.moveTo(14,11); ctx.lineTo(14,22); ctx.arc(22,22,8,Math.PI,0); ctx.lineTo(30,11); ctx.stroke();
  ctx.fillStyle = '#f0a84a'; ctx.beginPath(); ctx.ellipse(20,18,2.5,5,0.15,0,Math.PI*2); ctx.fill();
  ctx.fillStyle = '#131c28';
  ctx.beginPath(); ctx.moveTo(-24,11); ctx.lineTo(-24,50); ctx.quadraticCurveTo(-24,56,-16,56); ctx.lineTo(12,56); ctx.quadraticCurveTo(16,56,16,50); ctx.lineTo(16,11); ctx.closePath(); ctx.fill();
  ctx.fillStyle = dim;   ctx.fillRect(-24,11,3,40);
  ctx.fillStyle = color; ctx.fillRect(-24,11,1.5,40);
  ctx.fillStyle = color; ctx.fillRect(56,-12,3,4);
  ctx.fillStyle = dim;   ctx.fillRect(-16,-20,10,6);
  ctx.fillStyle = color; ctx.fillRect(-14,-18,3,4); ctx.fillRect(-8,-18,3,4);
  drawMuzzleFlash(56, flash, color);
}

function drawShotgun(color, dim, flash) {
  ctx.scale(0.65, 0.65);
  ctx.fillStyle = '#1a1a10';
  ctx.beginPath(); ctx.roundRect(-28,-18,90,26,3); ctx.fill();
  ctx.fillStyle = '#282818'; ctx.fillRect(-26,-18,86,6);
  ctx.fillStyle = dim;       ctx.fillRect(0,-14,40,10);
  ctx.fillStyle = '#151510'; ctx.fillRect(-28,-4,70,14);
  ctx.fillStyle = '#0a0a06';
  for (let i = 0; i < 5; i++) ctx.fillRect(-26+i*13, 9, 3, 4);
  ctx.strokeStyle = '#282818'; ctx.lineWidth = 2;
  ctx.beginPath(); ctx.moveTo(16,10); ctx.lineTo(16,24); ctx.arc(26,24,10,Math.PI,0); ctx.lineTo(36,10); ctx.stroke();
  ctx.fillStyle = '#4af0c8'; ctx.beginPath(); ctx.ellipse(24,20,3,6,0.15,0,Math.PI*2); ctx.fill();
  ctx.fillStyle = '#1a1a10';
  ctx.beginPath(); ctx.moveTo(-28,10); ctx.lineTo(-28,55); ctx.quadraticCurveTo(-28,62,-18,62); ctx.lineTo(14,62); ctx.quadraticCurveTo(16,62,16,55); ctx.lineTo(16,10); ctx.closePath(); ctx.fill();
  ctx.fillStyle = dim;   ctx.fillRect(-28,10,3,48);
  ctx.fillStyle = color; ctx.fillRect(-28,10,1.5,48);
  ctx.fillStyle = color; ctx.fillRect(60,-14,3,5);
  ctx.fillStyle = dim;   ctx.fillRect(-18,-24,14,7);
  ctx.fillStyle = color; ctx.fillRect(-16,-22,3,5); ctx.fillRect(-8,-22,3,5);
  drawMuzzleFlash(62, flash, color);
}

function drawSMG(color, dim, flash) {
  ctx.scale(0.52, 0.52);
  // Compact body
  ctx.fillStyle = '#12101a';
  ctx.beginPath(); ctx.roundRect(-18,-14,72,20,3); ctx.fill();
  ctx.fillStyle = '#1e1a2e'; ctx.fillRect(-16,-14,68,5);
  ctx.fillStyle = dim;       ctx.fillRect(4,-11,24,8);
  ctx.fillStyle = '#0e0c16'; ctx.fillRect(-18,-4,56,12);
  // Foregrip
  ctx.fillStyle = '#0a0812';
  ctx.beginPath(); ctx.roundRect(-16,8,16,28,2); ctx.fill();
  ctx.fillStyle = color; ctx.fillRect(-16,8,1.5,28);
  // Trigger guard
  ctx.strokeStyle = '#1e1a2e'; ctx.lineWidth = 1.5;
  ctx.beginPath(); ctx.moveTo(10,8); ctx.lineTo(10,18); ctx.arc(17,18,7,Math.PI,0); ctx.lineTo(24,8); ctx.stroke();
  ctx.fillStyle = '#d08030'; ctx.beginPath(); ctx.ellipse(16,15,2,4,0.15,0,Math.PI*2); ctx.fill();
  // Grip
  ctx.fillStyle = '#0a0812';
  ctx.beginPath(); ctx.moveTo(-18,8); ctx.lineTo(-18,44); ctx.quadraticCurveTo(-18,50,-10,50); ctx.lineTo(10,50); ctx.quadraticCurveTo(14,50,14,44); ctx.lineTo(14,8); ctx.closePath(); ctx.fill();
  ctx.fillStyle = dim;   ctx.fillRect(-18,8,3,38);
  ctx.fillStyle = color; ctx.fillRect(-18,8,1.5,38);
  // Extended mag
  ctx.fillStyle = '#0a0812'; ctx.beginPath(); ctx.roundRect(-12,46,24,14,2); ctx.fill();
  ctx.fillStyle = dim;   ctx.fillRect(-10,46,4,14);
  ctx.fillStyle = color; ctx.fillRect(52,-10,3,4);
  ctx.fillStyle = dim;   ctx.fillRect(-10,-18,8,6);
  ctx.fillStyle = color; ctx.fillRect(-9,-17,2.5,4); ctx.fillRect(-5,-17,2.5,4);
  drawMuzzleFlash(52, flash, color);
}

function drawSniper(color, dim, flash) {
  ctx.scale(0.58, 0.58);
  // Long barrel
  ctx.fillStyle = '#0a0a08';
  ctx.beginPath(); ctx.roundRect(-20,-8,120,14,2); ctx.fill();
  ctx.fillStyle = '#181810'; ctx.fillRect(-18,-8,116,4);
  // Suppressor tip
  ctx.fillStyle = '#0e0e0a';
  ctx.beginPath(); ctx.roundRect(96,-10,10,18,3); ctx.fill();
  ctx.fillStyle = dim; ctx.fillRect(97,-8,8,6);
  // Receiver
  ctx.fillStyle = '#141410';
  ctx.beginPath(); ctx.roundRect(-20,-14,74,26,2); ctx.fill();
  ctx.fillStyle = '#201e10'; ctx.fillRect(-18,-14,70,6);
  // Bolt
  ctx.fillStyle = dim; ctx.fillRect(20,-12,16,8);
  ctx.fillStyle = color; ctx.fillRect(32,-9,6,2);
  // Scope
  ctx.fillStyle = '#0c0c0a';
  ctx.beginPath(); ctx.roundRect(-4,-20,36,8,3); ctx.fill();
  ctx.fillStyle = dim; ctx.fillRect(-2,-19,32,3);
  ctx.fillStyle = color; ctx.fillRect(12,-20,2,3); ctx.fillRect(20,-20,2,3);
  // Trigger guard
  ctx.strokeStyle = '#201e10'; ctx.lineWidth = 1.5;
  ctx.beginPath(); ctx.moveTo(14,12); ctx.lineTo(14,22); ctx.arc(22,22,8,Math.PI,0); ctx.lineTo(30,12); ctx.stroke();
  ctx.fillStyle = '#d08030'; ctx.beginPath(); ctx.ellipse(20,19,2.5,5,0.1,0,Math.PI*2); ctx.fill();
  // Grip
  ctx.fillStyle = '#0e0e08';
  ctx.beginPath(); ctx.moveTo(-20,12); ctx.lineTo(-20,58); ctx.quadraticCurveTo(-20,64,-12,64); ctx.lineTo(12,64); ctx.quadraticCurveTo(16,64,16,58); ctx.lineTo(16,12); ctx.closePath(); ctx.fill();
  ctx.fillStyle = dim;   ctx.fillRect(-20,12,3,48);
  ctx.fillStyle = color; ctx.fillRect(-20,12,1.5,48);
  // Bipod legs
  ctx.strokeStyle = dim; ctx.lineWidth = 2;
  ctx.beginPath(); ctx.moveTo(-14,12); ctx.lineTo(-22,32); ctx.stroke();
  ctx.beginPath(); ctx.moveTo(-6,12); ctx.lineTo(-2,32); ctx.stroke();
  drawMuzzleFlash(100, flash, color);
}

// ── main render loop ──────────────────────────────────────────────────────

export function initRenderer() {
  drawFrame();
}

function drawFrame() {
  if (state.gamePhase !== 'in-game') {
    requestAnimationFrame(drawFrame);
    return;
  }

  drawArena();

  // Bullets
  state.gameState.bullets.forEach(b => {
    const owner = state.gameState.players.find(p => p.id === b.ownerId);
    const color = owner ? (COLORS[owner.weaponType] || '#4af0c8') : '#4af0c8';
    ctx.globalAlpha = 0.85;
    ctx.fillStyle = color;
    ctx.beginPath(); ctx.arc(b.x, b.y, b.r, 0, Math.PI*2); ctx.fill();
    ctx.globalAlpha = 0.25;
    ctx.strokeStyle = color; ctx.lineWidth = 1;
    ctx.beginPath(); ctx.moveTo(b.x, b.y); ctx.lineTo(b.x - b.vx*3, b.y - b.vy*3); ctx.stroke();
    ctx.globalAlpha = 1;
  });

  // Sparks
  for (let i = sparks.length - 1; i >= 0; i--) {
    const s = sparks[i];
    s.x += s.vx; s.y += s.vy; s.vy += 0.12; s.life -= 0.04;
    if (s.life <= 0) { sparks.splice(i, 1); continue; }
    ctx.globalAlpha = s.life;
    ctx.fillStyle = s.color;
    ctx.beginPath(); ctx.arc(s.x, s.y, 1.5 * s.life, 0, Math.PI*2); ctx.fill();
  }
  ctx.globalAlpha = 1;

  // Hit numbers
  for (let i = hitNumbers.length - 1; i >= 0; i--) {
    const h = hitNumbers[i];
    h.y += h.vy; h.life -= 0.025;
    if (h.life <= 0) { hitNumbers.splice(i, 1); continue; }
    ctx.globalAlpha = h.life;
    ctx.fillStyle = h.color;
    ctx.font = 'bold 13px "Share Tech Mono"';
    ctx.textAlign = 'center';
    ctx.fillText(h.text, h.x, h.y);
  }
  ctx.globalAlpha = 1;

  // Guns
  state.gameState.players.forEach(p => {
    // Selection ring for local player
    if (p.id === state.myId) {
      ctx.save();
      ctx.strokeStyle = '#4af0c840';
      ctx.lineWidth = 1;
      ctx.setLineDash([3, 4]);
      ctx.beginPath(); ctx.arc(p.x, p.y, 28, 0, Math.PI*2); ctx.stroke();
      ctx.setLineDash([]);
      ctx.restore();
    }
    drawGun(p);
  });

  // Aim line
  const me = getPlayer(state.myId);
  if (me && me.alive) {
    ctx.save();
    ctx.strokeStyle = '#4af0c825';
    ctx.lineWidth = 0.5;
    ctx.setLineDash([4, 6]);
    ctx.beginPath();
    ctx.moveTo(me.x + Math.cos(state.mouseAngle)*30, me.y + Math.sin(state.mouseAngle)*30);
    ctx.lineTo(me.x + Math.cos(state.mouseAngle)*120, me.y + Math.sin(state.mouseAngle)*120);
    ctx.stroke();
    ctx.setLineDash([]);
    ctx.restore();
  }

  requestAnimationFrame(drawFrame);
}