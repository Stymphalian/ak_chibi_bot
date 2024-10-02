import React from "react";
import RoomSettingsForm from "../components/RoomSettings";
import { getUserChannelSettings, updateUserChannelSettings } from "../api/real";
import { redirect, useLoaderData } from "react-router-dom";
import { useAuth } from "../contexts/auth";
import { ChannelSettings } from "../models/models";
import { LoaderBlock } from "../components/LoaderBlock";

export async function action(data: any) {
    const request = data.request;
    const jsonData = await request.json();
    await updateUserChannelSettings(jsonData.channelName, jsonData);
    return redirect(`/settings`);
}

export function SettingsPage() {
    const [channelSettings, setChannelSettings] = React.useState<ChannelSettings>({
        channelName: "",
        minAnimationSpeed: 0,
        maxAnimationSpeed: 0,
        minVelocity: 0,
        maxVelocity: 0,
        minSpriteScale: 0,
        maxSpriteScale: 0,
        maxSpritePixelSize: 0
    })
    const [loading, setLoading] = React.useState(true)
    const auth = useAuth()

    React.useEffect(() => {
        async function fetchData() {
            const settings: ChannelSettings | null = await getUserChannelSettings(auth.userName);
            if (settings) {
                setChannelSettings(settings);
                setLoading(false);
            }
        }
        fetchData();
    }, [loading])

    return (
        <div>
            <h1>Room Settings</h1>
            <LoaderBlock loading={loading}>
                <RoomSettingsForm channelSettings={channelSettings}/>
            </LoaderBlock>
        </div>
    );
}
