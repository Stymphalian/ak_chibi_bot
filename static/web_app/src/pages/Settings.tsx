import React from "react";
import RoomSettingsForm from "../components/RoomSettings";
import { getUserChannelSettings } from "../api/real";
import { useAuth } from "../contexts/auth";
import { ChannelSettings } from "../models/models";
import { LoaderBlock } from "../components/LoaderBlock";

export function SettingsPage() {
    const [channelSettings, setChannelSettings] = React.useState<ChannelSettings>({
        channelName: "",
        minAnimationSpeed: 0,
        maxAnimationSpeed: 0,
        minVelocity: 0,
        maxVelocity: 0,
        minSpriteScale: 0,
        maxSpriteScale: 0,
        maxSpritePixelSize: 0,
        usernamesBlacklist: ""
    })
    const [loading, setLoading] = React.useState(true)
    const auth = useAuth()

    React.useEffect(() => {
        async function fetchData() {
            let accessToken = await auth.getAccessToken();
            const settings: ChannelSettings | null = await getUserChannelSettings(accessToken, auth.userName);
            if (settings) {
                setChannelSettings(settings);
                setLoading(false);
            }
        }
        fetchData();
    }, [])

    return (
        <div>
            <h1>Room Settings</h1>
            <LoaderBlock loading={loading}>
                <RoomSettingsForm channelSettings={channelSettings}/>
            </LoaderBlock>
        </div>
    );
}
