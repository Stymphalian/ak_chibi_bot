import React from "react"
import { AdminInfo, AdminRoomInfo, AdminChatterInfo } from "../models/models"
import { getAdminInfo, removeUserFromRoom, removeRoom } from "../api/admin"
import { LoaderBlock } from "../components/LoaderBlock"
import "./Admin.css"

function GeneralInfo(props: {
    metrics: Object,
    gcTime: string
}) {
    const metricsList = Object.entries(props.metrics).map(([key, value]) => (
        <tr key={key}>
            <td>{key}</td>
            <td>{value}</td>
        </tr>
    ));

    return (
        <div className="container bg-light border rounded-3 m-1 p-3">
            <div className="display-5 fw-bold">General Info</div>
            <hr />
            <div className="container">
                <table className="table">
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Value</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>Next Rooms GC Time</td>
                            <td>{props.gcTime}</td>
                        </tr>
                        {metricsList}
                    </tbody>
                </table>
            </div>
        </div>
    )
}

function ConnectionsInfo(props: { 
    data: Map<string, number>,
    num_conns: number,
    refresh: () => void
 }) {
    return (
        <div>
        <p>Num Connections: {props.num_conns}</p>

        <table className="sub-table">
            <thead>
                <tr>
                    <th className="sub-th">ConnectionId</th>
                    <th className="sub-th">Info</th>
                </tr>
            </thead>
            <tbody>
                {Object.entries(props.data).map(([connId, avgFps]) => (
                    <tr key={connId}>
                        <td className="sub-td">{connId}</td>
                        <td className="sub-td">Average FPS: {avgFps.toFixed(2)}</td>
                    </tr>
                ))}
            </tbody>
        </table>
        </div>
    )
}

function ChattersList(props: { 
    roomId: string,
    chatters: AdminChatterInfo[],
    refresh: () => void 
}) {
    async function handleRemoveUser(event: any) {
        const roomId = event.target.getAttribute('data-room-id');
        const userName = event.target.getAttribute('data-user-id');
        removeUserFromRoom(roomId, userName);
        props.refresh();
    }

    return (

        <div>
            <p>Number Chatters: {props.chatters.length}</p>
            <table className="sub-table">
                <thead>
                    <tr>
                        <th className="sub-th">Username</th>
                        <th className="sub-th">Operator</th>
                        <th className="sub-th">LastChatTime</th>
                        <th className="sub-th">Action</th>
                    </tr>
                </thead>
                <tbody>
                    {props.chatters.map(chatter => (
                        <tr key={chatter.username}>
                            <td className="sub-td">{chatter.username}</td>
                            <td className="sub-td">{chatter.operator}</td>
                            <td className="sub-td">{chatter.last_chat_time}</td>
                            <td>
                                <button 
                                    className="btn btn-danger" 
                                    data-room-id={props.roomId}
                                    data-user-id={chatter.username}
                                    onClick={handleRemoveUser} >
                                    Kick
                                </button>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}

function RoomRow(props: { data: AdminRoomInfo , refresh: () => void }) {
    function handleRemoveRoom(event: any) {
        const roomId = event.target.getAttribute('data-room-id');
        removeRoom(roomId);
        props.refresh();
    }

    return (
        <tr>
            <td>
                <a href={`https://twitch.tv/${props.data.channel_name}`} 
                    target="_blank" 
                    rel="noopener noreferrer" >
                    {props.data.channel_name}
                </a>
            </td>
            <td>
                <p>CreatedAt: {props.data.created_at}</p>
                <p>LastTime: {props.data.last_time_used}</p>
                <p>NextGCTime: {props.data.next_gc_time}</p>
            </td>
            <td>
                <ConnectionsInfo 
                    data={props.data.connection_average_fps}
                    num_conns={props.data.num_websocket_connections}
                    refresh={props.refresh} />
            </td>
            <td>
                <ChattersList 
                    roomId={props.data.channel_name} 
                    chatters={props.data.chatters}
                    refresh={props.refresh} />
            </td>
            <td>
                <button className="btn btn-danger" 
                    data-room-id={props.data.channel_name}
                    onClick={handleRemoveRoom} >
                    Remove
                </button>
            </td>
        </tr>
    )

}

function RoomsList(props: {
    roomData: AdminRoomInfo[]
    refresh: () => void
}) {
    const data = props.roomData;
    return (
        <div className="container bg-light border rounded-3 m-1 p-3">
            <div>
                <div className="lead display-5 fw-bold">Rooms List</div>
                <div>
                    <button className="btn btn-primary" onClick={props.refresh}>Refresh</button>
                </div>
            </div>
            
            <hr />
            <table id="roomsTable" className="table">
                <thead>
                    <tr>
                        <th scope="col">ChannelName</th>
                        <th scope="col">Info</th>
                        <th scope="col">Connections</th>
                        <th scope="col">Chatters</th>
                        <th scope="col">Action</th>
                    </tr>
                </thead>
                <tbody>
                    {data.map((d: AdminRoomInfo) => <RoomRow data={d} refresh={props.refresh} />)}
                </tbody>
            </table>
        </div>
    )
}

export function AdminPage() {
    const [adminInfo, setAdminInfo] = React.useState<AdminInfo>({} as AdminInfo);
    const [loading, setLoading] = React.useState(true)

    React.useEffect(() => {
        async function fetchData() {
            const resp: AdminInfo | null = await getAdminInfo();
            if (resp) {
                setAdminInfo(resp);
                setLoading(false);
            }
        }
        fetchData();
    }, [loading])

    return (
        <div className="container">
            <LoaderBlock loading={loading}>
                <GeneralInfo metrics={adminInfo.metrics} gcTime={adminInfo.next_gc_time} />
                <RoomsList roomData={adminInfo.rooms} refresh={() =>setLoading(true)} />
            </LoaderBlock>
        </div>
    )
}