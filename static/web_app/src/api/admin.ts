import { AdminInfo } from "../models/models";
import { validateChannelName, validateTwitchUserName } from "./utils";

export async function getAdminInfo(accessToken: string) : Promise<AdminInfo|null> {
    try {
        const url = '/api/admin/info/';
        const response = await fetch(
            url, 
            {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${accessToken}`,
                }
            }
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

export async function removeRoom(accessToken:string, channelName: string) {
    try {
        if (!validateChannelName(channelName)) {
            throw Error("invalid channel name");
        }

        const url = '/api/rooms/remove/';
        const response = await fetch(
            url, 
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${accessToken}`
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

export async function removeUserFromRoom(accessToken: string,channelName: string, userName: string) {
    try {
        if (!validateChannelName(channelName)) {
            throw Error("invalid channel name");
        }
        if (!validateTwitchUserName(userName)) {
            throw Error("invalid twitch username");
        }

        const url = '/api/rooms/users/remove/';
        const response = await fetch(
            url, 
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${accessToken}`
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