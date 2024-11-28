export type ChannelSettings = {
    channelName: string,
    minAnimationSpeed: number,
    maxAnimationSpeed: number,
    minVelocity: number,
    maxVelocity: number,
    minSpriteScale: number,
    maxSpriteScale: number,
    maxSpritePixelSize: number,
    usernamesBlacklist ?: string
};

export type AdminChatterInfo = {
    username: string
    operator: string
    last_chat_time: string
}

export type AdminRoomInfo = {
    channel_name: string
    created_at: string
    last_time_used: string
    chatters: AdminChatterInfo[]
    next_gc_time: string
    num_websocket_connections: number,
    connection_average_fps: Map<string, number>
}

export type AdminInfo = {
    rooms: AdminRoomInfo[]
    next_gc_time: string
    metrics: Map<string, any>
}


export interface AuthInfo {
    
}