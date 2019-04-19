package protocol


const (
	PACKET_ID_PING_REQ = 201
	PACKET_ID_PING_RES = 202

	PACKET_ID_ERROR_NTF = 203

	PACKET_ID_SESSION_CLOSE_SYS = 211


	PACKET_ID_LOGIN_REQ = 701
	PACKET_ID_LOGIN_RES = 702

	PACKET_ID_ROOM_ENTER_REQ     = 721
	PACKET_ID_ROOM_ENTER_RES     = 722
	PACKET_ID_ROOM_USER_LIST_NTF = 723
	PACKET_ID_ROOM_NEW_USER_NTF  = 724

	PACKET_ID_ROOM_LEAVE_REQ      = 726
	PACKET_ID_ROOM_LEAVE_RES      = 727
	PACKET_ID_ROOM_LEAVE_USER_NTF = 728

	PACKET_ID_ROOM_CHAT_REQ          = 731
	PACKET_ID_ROOM_CHAT_RES          = 732
	PACKET_ID_ROOM_CHAT_NOTIFY       = 733

	PACKET_ID_ROOM_RELAY_REQ          = 741
	PACKET_ID_ROOM_RELAY_NTF          = 742

	PACKET_ID_GAME_START_REQ = 751
	PACKET_ID_GAME_START_RES = 752
	PACKET_ID_GAME_START_NTF = 753

	PACKET_ID_GAME_BATTING_REQ = 761
	PACKET_ID_GAME_BATTING_RES = 762
	PACKET_ID_GAME_BATTING_NTF = 753

	PACKET_ID_GAME_RESULT_NTF = 764

	INTERNAL_PACKET_ID_DISCONNECTED_USER_TO_ROOM = 1602
)


