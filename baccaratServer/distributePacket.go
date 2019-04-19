package main

import (
	"bytes"
	"go.uber.org/zap"
	"time"

	. "golang_socketGameServer_codelab/gohipernetFake"

	"golang_socketGameServer_codelab/baccaratServer/connectedSessions"
	"golang_socketGameServer_codelab/baccaratServer/protocol"
)

func (server *ChatServer) DistributePacket(sessionIndex int32,
	sessionUniqueId uint64,
	packetData []byte,
	) {
	packetID := protocol.PeekPacketID(packetData)
	bodySize, bodyData := protocol.PeekPacketBody(packetData)
	NTELIB_LOG_DEBUG("DistributePacket", zap.Int32("sessionIndex", sessionIndex), zap.Uint64("sessionUniqueId", sessionUniqueId), zap.Int16("PacketID", packetID))


	packet := protocol.Packet{Id: packetID}
	packet.UserSessionIndex = sessionIndex
	packet.UserSessionUniqueId = sessionUniqueId
	packet.Id = packetID
	packet.DataSize = bodySize
	packet.Data = make([]byte, packet.DataSize)
	copy(packet.Data, bodyData)

	server.PacketChan <- packet

	NTELIB_LOG_DEBUG("_distributePacket", zap.Int32("sessionIndex", sessionIndex), zap.Int16("PacketId", packetID))
}


func (server *ChatServer) PacketProcess_goroutine() {
	NTELIB_LOG_INFO("start PacketProcess goroutine")

	for {
		if server.PacketProcess_goroutine_Impl() {
			NTELIB_LOG_INFO("Wanted Stop PacketProcess goroutine")
			break
		}
	}

	NTELIB_LOG_INFO("Stop rooms PacketProcess goroutine")
}

func (server *ChatServer) PacketProcess_goroutine_Impl() bool {
	IsWantedTermination := false
	defer PrintPanicStack()

	secondTimeticker := time.NewTicker(time.Second)
	defer secondTimeticker.Stop()


	for {
		select {
		case packet := <-server.PacketChan:
			{
				sessionIndex := packet.UserSessionIndex
				sessionUniqueId := packet.UserSessionUniqueId
				bodySize := packet.DataSize
				bodyData := packet.Data

				if packet.Id == protocol.PACKET_ID_LOGIN_REQ {
					ProcessPacketLogin(sessionIndex, sessionUniqueId, bodySize, bodyData)
				} else if packet.Id == protocol.PACKET_ID_SESSION_CLOSE_SYS {
					ProcessPacketSessionClosed(server,  sessionIndex, sessionUniqueId)
				} else {
					roomNumber, _ := connectedSessions.GetRoomNumber(sessionIndex)
					server.RoomMgr.PacketProcess(roomNumber, packet)
				}
			}
		case _ = <-secondTimeticker.C:
			{
				//TODO 한번에 모든 방을 다 조사할 필요가 없다. 밀리세컨드 단위로 타이머를 돌게 하고 그룹 단위로 방을 조사한다

				//TODO 배팅 대기 중이면 시간 지났는지 체크. 지났으면 카드 배분
				//TODO 게임 종료이면 대기 시간 지났는지 체크. 지났으면 방 상태를 NONE
			}
		}
	}

	return IsWantedTermination
}

func ProcessPacketLogin(sessionIndex int32,
	sessionUniqueId uint64,
	bodySize int16,
	bodyData []byte )  {
	//DB와 연동하지 않으므로 중복 로그인만 아니면 다 성공으로 한다
	var request protocol.LoginReqPacket
	if (&request).Decoding(bodyData) == false {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_PACKET_DECODING_FAIL)
		return
	}

	userID := bytes.Trim(request.UserID[:], "\x00");

	if len(userID) <= 0 {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_LOGIN_USER_INVALID_ID)
		return
	}

	curTime := NetLib_GetCurrnetUnixTime()

	if connectedSessions.SetLogin(sessionIndex, sessionUniqueId, userID, curTime) == false {
		_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_LOGIN_USER_DUPLICATION)
		return
	}

	_sendLoginResult(sessionIndex, sessionUniqueId, protocol.ERROR_CODE_NONE)
}

func _sendLoginResult(sessionIndex int32, sessionUniqueId uint64, result int16) {
	var response protocol.LoginResPacket
	response.Result = result
	sendPacket, _ := response.EncodingPacket()

	NetLibIPostSendToClient(sessionIndex, sessionUniqueId, sendPacket)
	NTELIB_LOG_DEBUG("SendLoginResult", zap.Int32("sessionIndex", sessionIndex), zap.Int16("result", result))
}


func ProcessPacketSessionClosed(server *ChatServer, sessionIndex int32, sessionUniqueId uint64) {
	roomNumber, _ := connectedSessions.GetRoomNumber(sessionIndex)

	if roomNumber > -1 {
		packet := protocol.Packet{
			sessionIndex,
			sessionUniqueId,
			protocol.PACKET_ID_ROOM_LEAVE_REQ,
			0,
			nil,
		}

		server.RoomMgr.PacketProcess(roomNumber, packet)
	}

	connectedSessions.RemoveSession(sessionIndex, true)
}


