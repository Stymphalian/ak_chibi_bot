package api

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/stretchr/testify/assert"
)

func Setup_TestApiServer() *ApiServer {
	roomManager := room.NewFakeRoomsManager()
	return NewApiServer(roomManager)
}
func TestApiServer_HandleRoomUpdate_HappyPath(t *testing.T) {
	assert := assert.New(t)
	sut := Setup_TestApiServer()
	err := sut.roomsManager.CreateRoomOrNoOp(context.TODO(), "test")
	if err != nil {
		assert.Fail(err.Error())
	}

	jsonBody := `{
	"channel_name":"test",
	"min_animation_speed":0.1,
	"max_animation_speed":3,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody := strings.NewReader(jsonBody)

	req := httptest.NewRequest("POST", "http://example.com/api/rooms/update", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleRoomUpdate).ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}

func TestApiServer_HandleRoomUpdate_InvalidConfiguration(t *testing.T) {
	assert := assert.New(t)
	sut := Setup_TestApiServer()
	err := sut.roomsManager.CreateRoomOrNoOp(context.TODO(), "test")
	if err != nil {
		assert.Fail(err.Error())
	}

	// Missing channel name
	jsonBody := `{
	"channel_name":"",
	"min_animation_speed":3,
	"max_animation_speed":0.1,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody := strings.NewReader(jsonBody)
	req := httptest.NewRequest("POST", "http://example.com/api/rooms/update", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleRoomUpdate).ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(resp.StatusCode, 400)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal("Channel name must be provided", string(body))

	// max_animation_speed < min_animation_speed
	jsonBody = `{
	"channel_name":"test",
	"min_animation_speed":3,
	"max_animation_speed":0.1,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody = strings.NewReader(jsonBody)
	req = httptest.NewRequest("POST", "http://example.com/api/rooms/update", reqBody)
	w = httptest.NewRecorder()
	sut.middleware(sut.HandleRoomUpdate).ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(resp.StatusCode, 400)
	body, _ = io.ReadAll(resp.Body)
	assert.Equal("Invalid configuration settings", string(body))

	// Invalid json (string value for min_animation_speed, expecting number)
	jsonBody = `{
	"channel_name":"test",
	"min_animation_speed": "0.1",
	"max_animation_speed": 3,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody = strings.NewReader(jsonBody)
	req = httptest.NewRequest("POST", "http://example.com/api/rooms/update", reqBody)
	w = httptest.NewRecorder()
	sut.middleware(sut.HandleRoomUpdate).ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(resp.StatusCode, 400)
	body, _ = io.ReadAll(resp.Body)
	assert.Equal("Invalid request body", string(body))
}

func TestApiServer_HandleRoomUpdate_RoomDoesNotExist(t *testing.T) {
	assert := assert.New(t)
	sut := Setup_TestApiServer()
	// Missing call to create the room

	jsonBody := `{
	"channel_name":"test",
	"min_animation_speed":3,
	"max_animation_speed":0.1,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody := strings.NewReader(jsonBody)
	req := httptest.NewRequest("POST", "http://example.com/api/rooms/update", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleRoomUpdate).ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(resp.StatusCode, 500)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal("Something went wrong! Please try again.", string(body))
}
