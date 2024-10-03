import { ChannelSettings } from "../models/models";

export async function getUserChannelSettings(channelName:string) {
    try {
        const url = '/api/rooms/settings/?';
        const query = new URLSearchParams({
            'channel_name': channelName
        });
        const response = await fetch(
            url + query.toString(), 
            {method: 'GET'}
        );
        if (!response.ok) {
            return null;
        }
        const jsonBody = await response.json();
        return {
            channelName: channelName,
            minAnimationSpeed: jsonBody["min_animation_speed"],
            maxAnimationSpeed: jsonBody["max_animation_speed"],
            minVelocity: jsonBody["min_movement_speed"],
            maxVelocity: jsonBody["max_movement_speed"],
            minSpriteScale: jsonBody["min_sprite_size"],
            maxSpriteScale: jsonBody["max_sprite_size"],
            maxSpritePixelSize: jsonBody["max_sprite_pixel_size"]
        };
    } catch (error) {
        console.error("Error fetching room settings for channel " + channelName, error);
        return null;
    }
}

export async function updateUserChannelSettings(channelName:string, updates:ChannelSettings) {
    try {
        const url = '/api/rooms/settings/';
        const response = await fetch(
            url, 
            {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    'channel_name': channelName,
                    'min_animation_speed': updates.minAnimationSpeed,
                    'max_animation_speed': updates.maxAnimationSpeed,
                    'min_movement_speed': updates.minVelocity,
                    'max_movement_speed': updates.maxVelocity,
                    'min_sprite_size': updates.minSpriteScale,
                    'max_sprite_size': updates.maxSpriteScale,
                    'max_sprite_pixel_size': updates.maxSpritePixelSize
                })
            }
        );
        if (!response.ok) {
            return null;
        }
        return updates;
    } catch (error) {
        console.error("Error updating room settings for channel " + channelName, error);
        return null;
    }
}