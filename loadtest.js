import http from 'k6/http';
import ws from 'k6/ws';
import { check } from 'k6';

export const options = {
    vus: 100,         
    duration: '45s', 
};

export default function () {
    const res = http.post('http://localhost:3000/api/init-guest');
    check(res, { 'logged in successfully': (r) => r.status === 200 });
    const token = res.json('token');

    const url = `ws://localhost:3000/ws?token=${token}`;
    const wsRes = ws.connect(url, null, function (socket) {
        
        let angle = 0;

        socket.on('open', () => {
            
            // 1. THE MISSING PIECE: Tell the server to put this bot in a room!
            socket.send(JSON.stringify({ 
                type: 'join_room', // <-- Change this to match your actual lobby command
                weapon: 'smg' 
            }));

            // 2. Wait 1 second for the server to matchmake, then start shooting
            socket.setTimeout(function() {
                socket.setInterval(function timeout() {
                    angle += 0.2; 
                    if (angle > Math.PI * 2) angle = 0;
                    
                    socket.send(JSON.stringify({ 
                        type: 'input', 
                        angle: angle, 
                        shoot: true 
                    }));
                }, 33);
            }, 1000);
        });

        socket.setTimeout(function () {
            socket.close();
        }, 45000);
    });
}