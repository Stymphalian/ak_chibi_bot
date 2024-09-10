package room

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
	"google.golang.org/api/option"
)

// TODO
// Ideally we don't create restore points whenever we destroy the server
// We should be using a DB to keep the persistent state. This way even
// if the server goes down and comes back up we won't lose any of the state
// of the User's chibis.

// TODO:
// Make this an interface so that I can save to local file instead
// of to firestore

type SaveData struct {
	Rooms []SaveDataRoom `firestore:"rooms"`
}

type SaveDataRoom struct {
	ChannelName string            `firestore:"channel"`
	Chatters    []SaveDataChatter `firestore:"chatters"`
}

type SaveDataChatter struct {
	Username        string            `firestore:"username"`
	UsernameDisplay string            `firestore:"username_display"`
	Operator        string            `firestore:"operator"`
	OperatorFaction spine.FactionEnum `firestore:"operator_faction"`
}

func CreateFirestoreClient(ctx context.Context, projectID string, credsFilePath string) (*firestore.Client, error) {
	client, err := firestore.NewClient(
		ctx,
		projectID,
		option.WithCredentialsFile(credsFilePath))
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return nil, err
	}
	return client, nil
}

func CreateSaveData(rm *RoomsManager) *SaveData {
	log.Println("Creating restore point of Rooms")
	saveData := &SaveData{}
	for roomName, room := range rm.Rooms {

		saveDataChatters := make([]SaveDataChatter, 0)
		for _, chatter := range room.GetChatters() {
			saveDataChatters = append(saveDataChatters, SaveDataChatter{
				Username:        chatter.UserName,
				UsernameDisplay: chatter.UserNameDisplay,
				Operator:        chatter.CurrentOperator.OperatorId,
				OperatorFaction: chatter.CurrentOperator.Faction,
			})
		}

		saveDataRoom := SaveDataRoom{
			ChannelName: roomName,
			Chatters:    saveDataChatters,
		}
		saveData.Rooms = append(saveData.Rooms, saveDataRoom)
	}
	return saveData
}

func RestoreSaveData(rm *RoomsManager, saveData *SaveData) {
	log.Println("Restoring rooms from saved data")

	ctx := context.Background()
	for _, savedRoom := range saveData.Rooms {
		if err := rm.CreateRoomOrNoOp(savedRoom.ChannelName, ctx); err != nil {
			log.Println("Failed to create room", savedRoom.ChannelName, err)
			continue
		}

		for _, chatter := range savedRoom.Chatters {
			room := rm.Rooms[savedRoom.ChannelName]
			err := room.AddOperatorToRoom(
				chatter.Username,
				chatter.UsernameDisplay,
				chatter.Operator,
				chatter.OperatorFaction,
			)
			if err != nil {
				log.Printf("Failed to add chatter %s(%s) to room %s", chatter.Username, chatter.Operator, savedRoom.ChannelName)
				continue
			}
		}
	}
}
