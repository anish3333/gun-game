import ws from 'k6/ws';
import { check } from 'k6';

export const options = {
    vus: 500, // 500 players = 250 concurrent matches
    duration: '30s', 
};

export default function () {
    const url = 'ws://localhost:3000/ws';

    const res = ws.connect(url, null, function (socket) {
        let inGame = false;

        socket.on('message', (msg) => {
            const data = JSON.parse(msg);

            // 1. When connected, ask for the room list
            if (data.type === 'hello') {
                socket.send(JSON.stringify({ type: 'list_rooms' }));
            }

            // 2. Try to join a random waiting room, or create one if lobby is empty
            if (data.type === 'room_list') {
                if (data.rooms?.length > 0) {
                    // Pick a random room to avoid all 500 VUs trying to join the exact same one
                    let randomRoom = data.rooms[Math.floor(Math.random() * data.rooms.length)];
                    socket.send(JSON.stringify({ type: 'join_room', code: randomRoom.code, weapon: 'sniper' }));
                } else {
                    socket.send(JSON.stringify({ type: 'create_room', weapon: 'smg' }));
                }
            }

            // 3. If we tried to join a room but someone else beat us to it, create a new one
            if (data.type === 'error' && data.message === 'Room is full.') {
                socket.send(JSON.stringify({ type: 'create_room', weapon: 'smg' }));
            }

            // 4. The match has begun!
            if (data.type === 'match_start') {
                inGame = true;
            }
        });

        // 5. Spam inputs ONLY if the game has actually started
        socket.setInterval(function timeout() {
            if (inGame) {
                // Spinning and shooting wildly
                socket.send(JSON.stringify({ type: 'input', angle: Math.random() * 6.28, shoot: true }));
            }
        }, 33);

        socket.setTimeout(function () {
            socket.close();
        }, 30000); 
    });

    check(res, { 'status is 101': (r) => r && r.status === 101 });
}