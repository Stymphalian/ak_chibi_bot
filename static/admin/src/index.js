const queryString = window.location.search;
const searchParams = new URLSearchParams(queryString);
const secret = searchParams.get('secret');

function fillInRoomsTable(data) {
    const tableBody = document.querySelector('#roomsTable tbody');
    // Clear any existing rows
    tableBody.innerHTML = '';

    // Populate table with room data
    data["Rooms"].forEach(room => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>
                <a href="https://twitch.tv/${room.ChannelName}">
                    ${room.ChannelName}
                </a>
            </td>
            <td>${room.LastTimeUsed}</td>
            <td>${room.NumWebsocketConnections}</td>
            <td>
                <table class="sub-table">
                    <thead>
                        <tr>
                            <th class="sub-th">Username</th>
                            <th class="sub-th">Operator</th>
                            <th class="sub-th">LastChatTime</th>
                            <th class="sub-th">Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${room.Chatters.map(chatter => `
                            <tr>
                                <td class="sub-td">${chatter.Username}</td>
                                <td class="sub-td">${chatter.Operator}</td>
                                <td class="sub-td">${chatter.LastChatTime}</td>
                                <td>
                                    <button class="remove-user-button" 
                                        data-room-id="${room.ChannelName}"
                                        data-user-id="${chatter.Username}">
                                        Remove
                                    </button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </td>
            <td>
                <button class="remove-button" data-room-id="${room.ChannelName}">Remove</button>
            </td>
        `;
        tableBody.appendChild(row);
    });

    // Add event listeners to all remove buttons
    document.querySelectorAll('.remove-button').forEach(button => {
        button.addEventListener('click', async (event) => {
            const roomId = event.target.getAttribute('data-room-id');
            try {
                const response = await fetch(`/admin/room/remove?secret=${secret}`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ channel_name: roomId }),
                });

                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }

                // Optionally, refresh the room list after removal
                loadRoomList();
            } catch (error) {
                console.error('Error removing room:', error);
            }
        });
    });

    // Add event listeners to all remove buttons
    document.querySelectorAll('.remove-user-button').forEach(button => {
        button.addEventListener('click', async (event) => {
            const roomId = event.target.getAttribute('data-room-id');
            const userName = event.target.getAttribute('data-user-id');
            try {
                const response = await fetch(`/admin/user/remove?secret=${secret}`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        channel_name: roomId,
                        username: userName 
                    }),
                });

                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }

                // Optionally, refresh the room list after removal
                loadRoomList();
            } catch (error) {
                console.error('Error removing room:', error);
            }
        });
    });
}

function fillGeneralInfoTable(data) {
    const metricsTableBody = document.querySelector('#metricsTable tbody');
    metricsTableBody.innerHTML = `
        <tr>
            <td>Restore</td>
            <td><button class="restore-button normal-button">Restore</button></td>
        </tr>
    `;
    for (const [key, value] of Object.entries(data["Metrics"])) {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td> ${key} </td>
            <td> ${value} </td>
        `;
        metricsTableBody.appendChild(row);
    }

    // Add event listeners to all remove buttons
    document.querySelectorAll('.restore-button').forEach(button => {
        button.addEventListener('click', async (event) => {
            try {
                const response = await fetch(`/admin/rooms/restore?secret=${secret}`, {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                });

                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }

                // Optionally, refresh the room list after removal
                loadRoomList();
            } catch (error) {
                console.error('Error removing room:', error);
            }
        });
    });

}

async function loadRoomList() {
    try {
        const response = await fetch(`/admin/list?secret=${secret}`);
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        const data = await response.json();
        console.log(data)

        fillInRoomsTable(data);
        fillGeneralInfoTable(data);
    } catch (error) {
        console.error('Error fetching room list:', error);
    }
}

// Load room list when the page is loaded
window.onload = loadRoomList;