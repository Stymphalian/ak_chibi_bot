import { AdminInfo } from "../models/models";

export async function getAdminInfo() : Promise<AdminInfo|null> {
    try {
        const url = '/api/admin/info/';
        const response = await fetch(
            url, 
            {method: 'GET'}
        );
        if (!response.ok) {
            return null;
        }
        const jsonBody = await response.json();
        return jsonBody;
    } catch (error) {
        console.error("Error fetching admin info", error);
        return null;
    }
}

export async function removeRoom(channelName: string) {
    try {
        const url = '/api/rooms/remove/';
        const response = await fetch(
            url, 
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    'channel_name': channelName
                })
            }
        );
        if (!response.ok) {
            return null;
        }
        return null;
    } catch (error) {
        console.error("Error removing room", error);
        return null;
    }
}

export async function removeUserFromRoom(channelName: string, userName: string) {
    try {
        const url = '/api/rooms/users/remove/';
        const response = await fetch(
            url, 
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    'channel_name': channelName,
                    'username': userName
                })
            }
        );
        if (!response.ok) {
            return null;
        }
    } catch (error) {
        console.error("Error removing user from room", error);  
        return null;
    }
}