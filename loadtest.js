import http from 'k6/http';
import ws from 'k6/ws';
import { sleep } from 'k6';

export const options = {
    vus: 100,         // UNLEASH THE HORDE (100 bots = 50 simultaneous matches)
    duration: '45s', 
};

export default function () {
    // Stagger connections over 10 seconds to gently spin up the 50 rooms
    sleep(Math.random() * 10); 

    const res = http.post('http://localhost:3000/api/init-guest');
    const token = res.json('token');

    ws.connect(`ws://localhost:3000/ws?token=${token}`, null, function (socket) {
        
        socket.on('message', (rawMsg) => {
            const msg = JSON.parse(rawMsg);
            if (msg.type === 'snapshot' || msg.Type === 'snapshot') return;
            
            // 1. Handshake & Room Joining (Now battle-tested)
            if (msg.type === 'room_list' || msg.type === 'hello') {
                const rooms = msg.rooms || [];
                let foundRoom = null;
                
                for (let r of rooms) {
                    if (r.players < 2) {
                        foundRoom = r.code;
                        break;
                    }
                }

                if (foundRoom) {
                    socket.send(JSON.stringify({ type: 'join_room', code: foundRoom, weapon: 'smg' }));
                } else {
                    socket.send(JSON.stringify({ type: 'create_room', weapon: 'smg' }));
                }
            }

            // 2. The Fierce Combat AI
            if (msg.type === 'match_start') {
                let angle = Math.random() * Math.PI * 2; // Start aiming in a random direction
                let sweepDirection = 1;
                
                socket.setInterval(() => {
                    // --- THE FIERCE AIMING LOGIC ---
                    if (Math.random() < 0.05) {
                        // 5% chance to perform an instant "Flick Shot" to a completely random angle
                        angle = Math.random() * Math.PI * 2;
                        sweepDirection = Math.random() > 0.5 ? 1 : -1;
                    } else {
                        // Otherwise, perform a "Sweeping Spray" (dragging the gun back and forth)
                        angle += (0.15 * sweepDirection);
                        if (Math.random() < 0.1) sweepDirection *= -1; // Randomly change sweep direction
                    }
                    
                    // 90% chance to hold the trigger down, 10% chance to pause (simulates burst firing/reloading)
                    const isShooting = Math.random() < 0.90; 

                    socket.send(JSON.stringify({ type: 'input', angle: angle, shoot: isShooting }));
                }, 33); // 30 inputs per second
            }

            // 3. Retry on collision
            if (msg.type === 'error') {
                sleep(1);
                socket.send(JSON.stringify({ type: 'list_rooms' }));
            }
        });

        socket.setTimeout(() => socket.close(), 40000);
    });
}