package roomPkg

import (
	"go.uber.org/zap"

	"main/connectedSessions"
	"main/protocol"
	. "gohipernetFake"
)

// TODO 방 입장 중에 유저가 연결을 끊을 수 있으므로 이때 문제가 없는지 꼼꼼한 확인 필요
func (room *baseRoom) _packetProcess_EnterUser(inValidUser *roomUser, packet protocol.Packet) int16 {
	curTime := NetLib_GetCurrnetUnixTime()
	sessionIndex := packet.UserSessionIndex
	sessionUniqueId := packet.UserSessionUniqueId
	NTELIB_LOG_INFO("[[[Room _packetProcess_EnterUser]]]")

	var requestPacket protocol.RoomEnterReqPacket
	(&requestPacket).Decoding(packet.Data)

	userID := packet.UserID
	if userID == nil {
		_sendRoomEnterResult(sessionIndex, sessionUniqueId, 0,0, protocol.ERROR_CODE_ENTER_ROOM_INVALID_USER_ID)
		return protocol.ERROR_CODE_ENTER_ROOM_INVALID_USER_ID
	}

	userInfo := addRoomUserInfo{
		userID,
		sessionIndex,
		sessionUniqueId,
	}
	newUser, addResult := room.addUser(userInfo)

	if addResult != protocol.ERROR_CODE_NONE {
		_sendRoomEnterResult(sessionIndex, sessionUniqueId, 0, 0, addResult)
		return addResult
	}

	if connectedSessions.SetRoomNumber(sessionIndex, sessionUniqueId, room.getNumber(), curTime) == false {
		_sendRoomEnterResult(sessionIndex, sessionUniqueId, 0, 0, protocol.ERROR_CODE_ENTER_ROOM_INVALID_SESSION_STATE)
		return protocol.ERROR_CODE_ENTER_ROOM_INVALID_SESSION_STATE
	}

	if room.getCurUserCount() > 1 {
		//룸의 다른 유저에게 통보한다.
		room._sendNewUserInfoPacket(newUser)

		// 지금 들어온 유저에게 이미 채널에 있는 유저들의 정보를 보낸다
		room._sendUserInfoListPacket(newUser)
	}


	roomNumebr := room.getNumber()
	_sendRoomEnterResult(sessionIndex, sessionUniqueId, roomNumebr, newUser.RoomUniqueId, protocol.ERROR_CODE_NONE)
	return protocol.ERROR_CODE_NONE
}

func _sendRoomEnterResult(sessionIndex int32, sessionUniqueId uint64, roomNumber int32, userUniqueId uint64, result int16) {
	response := protocol.RoomEnterResPacket{
		result,
		roomNumber,
		userUniqueId,
	}
	sendPacket, _ := response.EncodingPacket()
	NetLibIPostSendToClient(sessionIndex, sessionUniqueId, sendPacket)
}


func (room *baseRoom) _sendUserInfoListPacket(user *roomUser) {
	NTELIB_LOG_DEBUG("Room _sendUserInfoListPacket", zap.Uint64("SessionUniqueId", user.netSessionUniqueId))

	userCount, userInfoListSize, userInfoListBuffer := room.allocAllUserInfo(user.netSessionUniqueId)

	var response protocol.RoomUserListNtfPacket
	response.UserCount = userCount
	response.UserList = userInfoListBuffer
	sendBuf, _ := response.EncodingPacket(userInfoListSize)
	NetLibIPostSendToClient(user.netSessionIndex, user.netSessionUniqueId, sendBuf)
}

func (room *baseRoom) _sendNewUserInfoPacket(user *roomUser) {
	NTELIB_LOG_DEBUG("Room _sendNewUserInfoPacket", zap.Uint64("SessionUniqueId", user.netSessionUniqueId))

	userInfoSize, userInfoListBuffer := room._allocUserInfo(user)

	var response protocol.RoomNewUserNtfPacket
	response.User = userInfoListBuffer
	sendBuf, packetSize := response.EncodingPacket(userInfoSize)
	room.broadcastPacket(int16(packetSize), sendBuf, user.netSessionUniqueId) // 자신을 제외하고 모든 유저에게 Send
}



func (room *baseRoom) _packetProcess_LeaveUser(user *roomUser, packet protocol.Packet) int16 {
	NTELIB_LOG_DEBUG("[[Room _packetProcess_LeaveUser ]")

	room._leaveUserProcess(user)

	sessionIndex := packet.UserSessionIndex
	sessionUniqueId := packet.UserSessionUniqueId
	_sendRoomLeaveResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_NONE)
	return protocol.ERROR_CODE_NONE
}

func (room *baseRoom) _leaveUserProcess(user *roomUser) {
	NTELIB_LOG_DEBUG("[[Room _leaveUserProcess]")

	roomUserUniqueId := user.RoomUniqueId
	userSessionIndex := user.netSessionIndex
	userSessionUniqueId := user.netSessionUniqueId

	room._removeUser(user)

	room._sendRoomLeaveUserNotify(roomUserUniqueId, userSessionUniqueId)

	curTime := NetLib_GetCurrnetUnixTime()
	connectedSessions.SetRoomNumber(userSessionIndex, userSessionUniqueId, -1, curTime)
}


func _sendRoomLeaveResult(sessionIndex int32, sessionUniqueId uint64, result int16) {
	response := protocol.RoomLeaveResPacket{ result }
	sendPacket, _ := response.EncodingPacket()
	NetLibIPostSendToClient(sessionIndex, sessionUniqueId, sendPacket)
}

func (room *baseRoom) _sendRoomLeaveUserNotify(roomUserUniqueId uint64, userSessionUniqueId uint64) {
	NTELIB_LOG_DEBUG("Room _sendRoomLeaveUserNotify", zap.Uint64("userSessionUniqueId", userSessionUniqueId), zap.Int32("RoomIndex", room._index))

	notifyPacket := protocol.RoomLeaveUserNtfPacket{roomUserUniqueId	}
	sendBuf, packetSize := notifyPacket.EncodingPacket()
	room.broadcastPacket(int16(packetSize), sendBuf, userSessionUniqueId)
}