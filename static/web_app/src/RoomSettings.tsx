import React from 'react';
import { useState } from 'react';
import { Outlet, Link } from "react-router-dom";
import { Button, Col, Form, Row, FloatingLabel, Container } from 'react-bootstrap';


function RoomFormFloatingElem(props: {
    controlId: string;
    label: string;
    defaultValue: string;
}) {
    return (
        <>
            <FloatingLabel
                controlId={props.controlId}
                label={props.label}
                className="mb-3"
            >
                <Form.Control
                    type="text"
                    defaultValue={props.defaultValue}
                    required
                />
            </FloatingLabel>
        </>
    )
}

type IsValidFunction = (val: string) => boolean;

function RoomFormNumberElem(props: {
    controlId: string;
    label: string;
    defaultValue: string;
    isValidFn?: IsValidFunction
}) {
    const IsNumberFn = (val: string, min: number = 0, max: number = 10.1) => {
        console.log(val, !isNaN(parseFloat(val)));
        return !isNaN(parseFloat(val));
    }
    const isValidFn = props.isValidFn ? props.isValidFn : IsNumberFn;
    const [valid, setValid] = useState(isValidFn(props.defaultValue));

    return (
        <Form.Group as={Row} className="mb-1" controlId={props.controlId}>
            <Form.Label column>
                {props.label}
            </Form.Label>
            <Col>
                <Form.Control type="text"
                    placeholder={props.defaultValue}
                    defaultValue={props.defaultValue}
                    onChange={e => {
                        const val = e.target.value;
                        console.log(val, isValidFn(val));
                        setValid(isValidFn(val));
                    }}
                    isValid={valid}
                    isInvalid={!valid}
                    required
                />
                <Form.Control.Feedback type="invalid">
                    Value must be a number betweeen 0 and 10
                </Form.Control.Feedback>
            </Col>
        </Form.Group>
    );
}

function RoomSettings(props: any) {
    const [validated, setValidated] = useState(true);

    function handleSubmit(event: React.FormEvent<any>) {
        event.preventDefault();
        event.stopPropagation();

        if (!event.currentTarget.checkValidity()) {
            setValidated(false);
            return;
        }
        // const form  = event.currentTarget;
        // setValidated(true);
        // event.preventDefault();
        console.log(event);
    }

    return (
        <Container className="p-3">
            <Form>
                <RoomFormFloatingElem controlId="formChannelName" label="Channel Name" defaultValue="stymphalian__" />
                <RoomFormNumberElem controlId="formMinAnimationSpeed" label="Minimum Animation Speed" defaultValue="0.5" />
                <RoomFormNumberElem controlId="formMaxAnimationSpeed" label="Maximum Animation Speed" defaultValue="3.0" />
                <RoomFormNumberElem controlId="formMinVelocity" label="Minimum Velocity" defaultValue="0.1" />
                <RoomFormNumberElem controlId="formMaxVelocity" label="Maximum Velocity" defaultValue="3.0" />
                <RoomFormNumberElem controlId="formMinSpriteScale" label="Minimum Sprite Scale" defaultValue="0.5" />
                <RoomFormNumberElem controlId="formMaxSpriteScale" label="Maximum Sprite Scale" defaultValue="2.0" />
                <RoomFormNumberElem controlId="formMaxSpritePixelSize" label="Maximum Sprite Size" defaultValue="350" />

                <Button variant="primary" type="submit" onClick={handleSubmit}>
                    Save
                </Button>
            </Form>
            <Link to={"/"}>Back</Link>
        </Container>

    )
}
export default RoomSettings;