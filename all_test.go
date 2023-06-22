package main

import "testing"

func TestTest(t *testing.T) {
	data := `"{"id":"87a33f41-eff9-4815-9620-27db4584b149","chatroom_id":6486575,"content":"[emote:37244:]","type":"message","created_at":"2023-06-22T04:57:02+00:00","sender":{"id":6634791,"username":"Mirko15","slug":"mirko15","identity":{"color":"#1475E1","badges":[{"type":"moderator","text":"Moderator"}]}}}"`
	_, err := handleChatMessageEvent(data)
	if err != nil {
		t.Error(err)
	}
}

func TestTest2(t *testing.T) {
	data := `"{"id":"044ef81d-8e4a-4e60-b7cb-9364ebcbf389","chatroom_id":1310407,"content":"whaa . \\"","type":"message","created_at":"2023-06-22T05:31:33+00:00","sender":{"id":6644015,"username":"untalpablogod","slug":"untalpablogod","identity":{"color":"#D399FF","badges":[]}}}"`
	_, err := handleChatMessageEvent(data)
	if err != nil {
		t.Error(err)
	}
}

func TestTest3(t *testing.T) {
	data := `"{\"id\":\"a109538a-50e2-4edc-b74a-1c716aed3a3e\",\"chatroom_id\":75062,\"content\":\"Do CREATOR ROUND PLAYLIST 4 it\\u2019s fireeeee\",\"type\":\"message\",\"created_at\":\"2023-06-22T07:38:47+00:00\",\"sender\":{\"id\":549307,\"username\":\"Vap3rz\",\"slug\":\"vap3rz\",\"identity\":{\"color\":\"#B9D6F6\",\"badges\":[]}}}"`
	_, err := handleChatMessageEvent(data)
	if err != nil {
		t.Error(err)
	}
}
