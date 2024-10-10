
export function validateChannelName(channel: string): boolean {
    // validate channel name is alphanumeric
    if (!/^[a-zA-Z0-9_-]{1,100}$/.test(channel)) {
        return false;
    }
    return true;
}

export function validateTwitchUserName(username: string): boolean {
    // validate username name is alphanumeric
    if (!/^[a-zA-Z0-9_-]{1,100}$/.test(username)) {
        return false;
    }
    return true;
}