import { useState } from 'react';
import { Alert } from 'react-bootstrap';
import { useForm, SubmitHandler, SubmitErrorHandler } from "react-hook-form"
import * as yup from "yup"
import { yupResolver } from "@hookform/resolvers/yup"

import { ChannelSettings } from '../models/models';
import {Code} from "./Code";
import { updateUserChannelSettings } from '../api/real';
import { useAuth } from '../contexts/auth';

const schema = yup
.object({
    channelName: yup.string().required(),
    minAnimationSpeed: yup.number().positive().min(0).max(10).required(),
    maxAnimationSpeed: yup.number().positive().min(0).max(10).required(),
    minVelocity: yup.number().positive().min(0).max(10).required(),
    maxVelocity: yup.number().positive().min(0).max(10).required(),
    minSpriteScale: yup.number().positive().min(0).max(10).required(),
    maxSpriteScale: yup.number().positive().min(0).max(10).required(),
    maxSpritePixelSize: yup.number().positive().min(100).max(2000).required(),
})
.required()

function RoomSettingsForm(props: {
    channelSettings: ChannelSettings
}) {
    const auth = useAuth();
    const [showAlert, setShowAlert] = useState(false);
    const [requestSuccessful, setRequestSuccessful] = useState(false);
    const cs = props.channelSettings;
    const {
        register,
        handleSubmit,
        formState,
    } = useForm<ChannelSettings>({
        resolver: yupResolver(schema)
    });
    const { errors } = formState;
    const resetAlert = () => {
        setShowAlert(false);
        setRequestSuccessful(false);
    };

    const onSubmit: SubmitHandler<ChannelSettings> = async (data) => {
        let accessToken = await auth.getAccessToken();
        let resp = await updateUserChannelSettings(accessToken, data.channelName, data);
        setShowAlert(true);
        setRequestSuccessful(resp === null ? false: true);
        setTimeout(resetAlert, 5000);
    }
    const onError: SubmitErrorHandler<ChannelSettings> = async (errors) => {
        setShowAlert(true);
        setRequestSuccessful(false);
        setTimeout(resetAlert, 5000);
    }

    return (
        <div className="container bg-light border rounded-3 m-1 p-3">
            <div className="lead">Update your channel settings.<br />
                This controls the min/max values for commands such 
                <Code>!chibi speed</Code>,
                <Code>!chibi velocity</Code> 
                and <Code>!chibi size</Code>
            </div>
            <hr />
            <form className="container" onSubmit={handleSubmit(onSubmit, onError)}>
                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Channel Name</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control text-muted" 
                            placeholder="channel name" 
                            defaultValue={cs.channelName} 
                            {...register("channelName")} readOnly />
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Min Animation Speed</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="0.5" 
                            defaultValue={cs.minAnimationSpeed} 
                            {...register("minAnimationSpeed")} />
                        {errors.minAnimationSpeed && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.minAnimationSpeed?.message}</div>}
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Max Animation Speed</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="2.0" 
                            defaultValue={cs.maxAnimationSpeed} 
                            {...register("maxAnimationSpeed")} />
                        {errors.maxAnimationSpeed && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.maxAnimationSpeed?.message}</div>}
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Min Velocity</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="0.5" 
                            defaultValue={cs.minVelocity}
                            {...register("minVelocity")} />
                        {errors.minVelocity && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.minVelocity?.message}</div>}
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Max Velocity</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="2.0" 
                            defaultValue={cs.maxVelocity} 
                            {...register("maxVelocity")} />
                        {errors.maxVelocity && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.maxVelocity?.message}</div>}
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Min Sprite Size</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="0.5" 
                            defaultValue={cs.minSpriteScale} 
                            {...register("minSpriteScale")} />
                        {errors.minSpriteScale && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.minSpriteScale?.message}</div>}
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Max Sprite Size</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="1.5" 
                            defaultValue={cs.maxSpriteScale} 
                            {...register("maxSpriteScale")} />
                        {errors.maxSpriteScale && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.maxSpriteScale?.message}</div>}
                    </div>
                </div>

                <div className="form-group row pb-1">
                    <label className="col-form-label col-sm-2">Max Sprite Pixel Size</label>
                    <div className="col-sm-10">
                        <input 
                            className="form-control" 
                            placeholder="350" 
                            defaultValue={cs.maxSpritePixelSize} 
                            {...register("maxSpritePixelSize")} />
                        {errors.maxSpritePixelSize && <div className="p-1 bg-warning bg-gradient text-black rounded">{errors.maxSpritePixelSize?.message}</div>}
                    </div>
                </div>

                <div className="form-group pt-2">
                    <input 
                        className="btn btn-primary" type="submit" value="Save"></input>
                </div>
            </form>

            
            <div>
                {
                    showAlert
                    && (
                        formState.isSubmitSuccessful && requestSuccessful
                        ? <div><hr /><Alert variant="success">Saved successfully!</Alert></div>
                        : <div><hr /><Alert variant="warning">Failed to save. Please try again!</Alert></div>
                    )
                }
            </div>
            
            <div hidden>
                <div>isSubmitted: {formState.isSubmitted ? "true": "false"}</div>
                <div>isSubmitSuccessful: {formState.isSubmitSuccessful ? "true": "false"}</div>
                <div>submitCount: {formState.submitCount}</div>
                <div>isValid: {formState.isValid ? "true": "false"}</div>
            </div>
        </div>
    )

}
export default RoomSettingsForm;