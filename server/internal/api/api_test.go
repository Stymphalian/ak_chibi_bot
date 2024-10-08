package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/stretchr/testify/assert"
)

func Setup_TestApiServer() (*ApiServer, *auth.FakeAuthService) {
	roomManager := room.NewFakeRoomsManager()
	authService := auth.NewFakeAuthService()
	roomsRepo := room.NewRoomRepositoryPsql()
	authService.IsAuthenticated = true
	authService.Username = "test"
	return NewApiServer(roomManager, authService, roomsRepo), authService
}

func TestApiServer_HandleRoomUpdate_NoAuth(t *testing.T) {
	assert := assert.New(t)
	sut, authService := Setup_TestApiServer()
	authService.IsAuthenticated = false
	err := sut.roomsManager.CreateRoomOrNoOp(context.TODO(), "test")
	if err != nil {
		assert.Fail(err.Error())
	}

	jsonBody := `{
	"channel_name":"test2",
	"min_animation_speed":0.1,
	"max_animation_speed":3,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody := strings.NewReader(jsonBody)

	req := httptest.NewRequest("POST", "http://example.com/api/rooms/settings/", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleUpdateRoomSettings).ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(401, resp.StatusCode)
}

func TestApiServer_HandleRoomUpdate_HappyPath(t *testing.T) {
	assert := assert.New(t)
	sut, _ := Setup_TestApiServer()
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

	req := httptest.NewRequest("POST", "http://example.com/api/rooms/settings/", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleUpdateRoomSettings).ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(200, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}

func TestApiServer_HandleRoomUpdate_InvalidConfiguration(t *testing.T) {
	assert := assert.New(t)
	sut, _ := Setup_TestApiServer()
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
	req := httptest.NewRequest("POST", "http://example.com/api/rooms/settings/", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleUpdateRoomSettings).ServeHTTP(w, req)
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
	req = httptest.NewRequest("POST", "http://example.com/api/rooms/settings/", reqBody)
	w = httptest.NewRecorder()
	sut.middleware(sut.HandleUpdateRoomSettings).ServeHTTP(w, req)
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
	req = httptest.NewRequest("POST", "http://example.com/api/rooms/settings/", reqBody)
	w = httptest.NewRecorder()
	sut.middleware(sut.HandleUpdateRoomSettings).ServeHTTP(w, req)
	resp = w.Result()
	assert.Equal(resp.StatusCode, 400)
	body, _ = io.ReadAll(resp.Body)
	assert.Equal("Invalid request body", string(body))
}

func TestApiServer_HandleRoomUpdate_RoomDoesNotExist(t *testing.T) {
	assert := assert.New(t)
	sut, auth := Setup_TestApiServer()
	auth.Username = "TestApiServer_HandleRoomUpdate_RoomDoesNotExist"
	// Missing call to create the room

	jsonBody := `{
	"channel_name": "TestApiServer_HandleRoomUpdate_RoomDoesNotExist",
	"min_animation_speed":3,
	"max_animation_speed":0.1,
	"min_velocity":0.1,
	"max_velocity":3,
	"min_sprite_scale":0.5,
	"max_sprite_scale":2,
	"max_sprite_pixel_size":300
	}`
	reqBody := strings.NewReader(jsonBody)
	req := httptest.NewRequest("POST", "http://example.com/api/rooms/settings/", reqBody)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleUpdateRoomSettings).ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(resp.StatusCode, 404)
}

func TestApiServer_HandleGetRoomSettings_CannotModifyOtherUsersRoom(t *testing.T) {
	assert := assert.New(t)
	sut, auth := Setup_TestApiServer()
	auth.Username = "TestApiServer_HandleGetRoomSettings_CannotModifyOtherUsersRoom"
	err := sut.roomsManager.CreateRoomOrNoOp(context.TODO(), "TestApiServer_HandleGetRoomSettings_CannotModifyOtherUsersRoom")
	if err != nil {
		assert.Fail(err.Error())
	}

	req := httptest.NewRequest(
		"GET",
		"http://example.com/api/rooms/settings/?channel_name=test",
		nil,
	)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleGetRoomSettings).ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(resp.StatusCode, 400)
}

func TestApiServer_HandleGetRoomSettings_RoomExistsButNotActiveInManager(t *testing.T) {
	assert := assert.New(t)
	sut, auth := Setup_TestApiServer()
	auth.Username = "TestApiServer_HandleGetRoomSettings_RoomExistsButNotActiveInManager"
	err := sut.roomsManager.CreateRoomOrNoOp(context.TODO(), auth.Username)
	if err != nil {
		assert.Fail(err.Error())
	}
	delete(sut.roomsManager.Rooms, auth.Username)

	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("http://example.com/api/rooms/settings/?channel_name=%s", auth.Username),
		nil,
	)
	w := httptest.NewRecorder()
	sut.middleware(sut.HandleGetRoomSettings).ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(resp.StatusCode, 200)
	body, _ := io.ReadAll(resp.Body)
	respObj := GetRoomSettingsResponse{
		MinAnimationSpeed:  0.1,
		MaxAnimationSpeed:  5,
		MinMovementSpeed:   0.1,
		MaxMovementSpeed:   2,
		MinSpriteSize:      0.5,
		MaxSpriteSize:      1.5,
		MaxSpritePixelSize: 350,
	}
	var gotObj GetRoomSettingsResponse
	err = json.Unmarshal(body, &gotObj)
	assert.NoError(err, string(body))
	assert.Equal(respObj, gotObj)
}
